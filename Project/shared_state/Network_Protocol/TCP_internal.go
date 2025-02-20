package Network_Protocol

import (
	"fmt"
	"net"
	"time"

	"github.com/mathiahb/TTK4145_V2025_Gruppe4/Constants"
)

// Protected functions

// Reads a string from a net.Conn onto a read channel
func read_from_TCP_Connection(TCP_Connection TCP_Connection) {
	deadline := time.Now().Add(Constants.TCP_READ_DEADLNE)
	TCP_Connection.connection.SetReadDeadline(deadline)

	data := make([]byte, 4096)

	bytes_received, err := TCP_Connection.connection.Read(data)

	if err == nil {
		message := string(data[0:bytes_received])
		TCP_Connection.Read_Channel <- message
	}
}

// Writes a string to a net.Conn
func write_to_TCP_Connection(TCP_Connection TCP_Connection, message string) {
	deadline := time.Now().Add(Constants.TCP_READ_DEADLNE)
	TCP_Connection.connection.SetWriteDeadline(deadline)

	data := []byte(message)
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

	for {
		select {
		case <-when_to_read_ticker.C:
			read_from_TCP_Connection(connection)

		case message := <-connection.Write_Channel:
			write_to_TCP_Connection(connection, message)

		case <-connection.close_channel:
			return
		}
	}
}

func create_TCP_Server(connection_manager *TCP_Connection_Manager) {
	address_to_listen := Constants.TCP_PORT

	listener, err := net.Listen("tcp", address_to_listen)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error in accepting connection: ", err)
			return
		}

		tcp_conn := connection.(*net.TCPConn)
		tcp_conn.SetNoDelay(true)

		connection_name := Constants.CLIENT_CREATED_NAME + connection.RemoteAddr().String()
		connection_object := New_TCP_Connection(connection_name, connection)
		connection_manager.Add_Connection(connection_object)

		// Hand over the accepted connection to the TCP Handler.
		go handler_TCP_Connection_with_Channel(connection_object, connection_manager)
	}
}

func create_TCP_Client(address string, connection_manager *TCP_Connection_Manager) {
	address_to_dial := address + Constants.TCP_PORT

	connection, err := net.Dial("tcp", address_to_dial)
	if err != nil {
		fmt.Println("Error setting up connection: ", err)
		return
	}

	tcp_conn := connection.(*net.TCPConn)
	tcp_conn.SetNoDelay(true)

	connection_name := Constants.CLIENT_CREATED_NAME + connection.RemoteAddr().String()
	connection_object := New_TCP_Connection(connection_name, connection)
	connection_manager.Add_Connection(connection_object)

	handler_TCP_Connection_with_Channel(connection_object, connection_manager)
}
