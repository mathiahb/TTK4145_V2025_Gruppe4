package TCP

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"Constants"
)

// Protected functions

// Reads a string from a net.Conn onto a read channel
func (TCP_Connection TCP_Connection) read(previous_split string) string {
	deadline := time.Now().Add(Constants.TCP_READ_DEADLNE)
	TCP_Connection.connection.SetReadDeadline(deadline)

	data := make([]byte, 4096)

	bytes_received, err := TCP_Connection.connection.Read(data)

	if err == io.EOF {
		TCP_Connection.Close()
	}

	if err == nil {
		message := string(data[0:bytes_received]) // Last split will always be incomplete! Add it to next split...

		split_messages := strings.Split(message, "\000") // Null terminated

		// Handle message that has been split due to buffer size or partial TCP transmission.
		split_messages[0] = previous_split + split_messages[0]

		last_split_id := len(split_messages) - 1
		last_split := split_messages[last_split_id]

		for _, split_message := range split_messages[0:last_split_id] {
			TCP_Connection.Read_Channel <- split_message
		}

		return last_split
	}

	// No data read.
	return previous_split
}

// Writes a string to a net.Conn
func (TCP_Connection TCP_Connection) write(message string) {
	deadline := time.Now().Add(Constants.TCP_READ_DEADLNE)
	TCP_Connection.connection.SetWriteDeadline(deadline)

	data := []byte(message)
	data = append(data, '\000') // Null terminated

	_, err := TCP_Connection.connection.Write(data)

	if err != nil {
		fmt.Println("Write didn't succeed, error: ", err)
	}
}

// Handles a TCP Connection, writing any data from the TCP_Channel onto the connection
// And reads any data from the connection onto the TCP_Channel.
func handler_TCP_Connection_with_Channel(connection TCP_Connection, connection_manager *TCP_Connection_Manager) {
	defer connection.connection.Close()
	defer connection_manager.Remove_Connection(connection)

	when_to_read_ticker := time.NewTicker(Constants.TCP_WAIT_BEFORE_READING_AGAIN)

	var last_read_split string = ""

	for {
		select {
		case <-when_to_read_ticker.C:
			last_read_split = connection.read(last_read_split)

		case message := <-connection.Write_Channel:
			connection.write(message)

		case <-connection.close_channel:
			return
		}

		time.Sleep(time.Microsecond)
	}
}

func create_TCP_Server(port string, connection_manager *TCP_Connection_Manager) {
	// Ensure correct format for net library ":port".
	address_to_listen := port
	if port[0] != ':' {
		address_to_listen = ":" + address_to_listen
	}

	// Create a listener
	listener, err := net.Listen("tcp", address_to_listen)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	// Accept all incoming connections
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error in accepting connection: ", err)
			return
		}

		tcp_conn := connection.(*net.TCPConn)
		tcp_conn.SetNoDelay(true)

		// Add the incoming connection to the connection manager
		// server_ and client_ is only used as a debugging tool to tell if a connection started on the server or client.
		connection_name := Constants.SERVER_CREATED_NAME + connection.RemoteAddr().String()
		connection_object := New_TCP_Connection(connection_name, connection_manager.Global_Read_Channel, connection)
		connection_manager.Add_Connection(connection_object)

		// Hand over the accepted connection to the TCP Handler.
		go handler_TCP_Connection_with_Channel(connection_object, connection_manager)
	}
}

func create_TCP_Client(address string, connection_manager *TCP_Connection_Manager) {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error setting up connection: ", err)
		return
	}

	tcp_conn := connection.(*net.TCPConn)
	tcp_conn.SetNoDelay(true)

	connection_name := Constants.CLIENT_CREATED_NAME + connection.RemoteAddr().String()
	connection_object := New_TCP_Connection(connection_name, connection_manager.Global_Read_Channel, connection)
	connection_manager.Add_Connection(connection_object)

	go handler_TCP_Connection_with_Channel(connection_object, connection_manager)
}
