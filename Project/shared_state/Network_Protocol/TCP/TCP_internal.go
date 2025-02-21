package TCP

import (
	"fmt"
	"net"

	"Constants"
)

func (connection_manager *TCP_Connection_Manager) setup_TCP_Connection(connection net.Conn) {
	// We don't want to delay, send everything asap.
	tcp_conn := connection.(*net.TCPConn)
	tcp_conn.SetNoDelay(true)

	// Add the incoming connection to the connection manager
	connection_name := Constants.SERVER_CREATED_NAME + connection.RemoteAddr().String()
	connection_object := New_TCP_Connection(connection_name, connection_manager.Global_Read_Channel, connection)
	connection_manager.Add_Connection(connection_object)

	// Hand over the accepted connection to a TCP Connection Handler.
	go connection_object.handle_TCP_Connection(connection_manager)
}

func (connection_manager *TCP_Connection_Manager) create_TCP_Server(port string) {
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

		connection_manager.setup_TCP_Connection(connection)
	}
}

func (connection_manager *TCP_Connection_Manager) create_TCP_Client(address string) {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error setting up connection: ", err)
		return
	}

	connection_manager.setup_TCP_Connection(connection)
}
