package TCP

import (
	"elevator_project/common"
	"fmt"
	"net"
	"strconv"
	"time"
)

func (connectionManager *TCPConnectionManager) setupTCPConnection(connection net.Conn) {
	// Add the incoming connection to the connection manager
	connectionName := connection.RemoteAddr().String()
	if connectionManager.DoesConnectionExist(connectionName) {
		return // We are already connected...
	}

	connectionObject := NewTCPConnection(connectionName, connectionManager.GlobalReadChannel, connection)
	connectionManager.AddConnection(connectionObject)

	// We don't want to delay, send everything asap.
	tcp_conn := connection.(*net.TCPConn)
	tcp_conn.SetNoDelay(true)

	// Hand over the accepted connection to a TCP Connection Handler.
	go connectionObject.handleTCPConnection(connectionManager)
}

func (connectionManager *TCPConnectionManager) tcpListener(listener *net.Listener) {
	// Accept all incoming connections
	for {
		connection, err := (*listener).Accept()
		if err != nil {
			select {
			case <-connectionManager.stopServerChannel:
				return
			default:
				fmt.Println("Error in accepting connection: ", err)
				continue
			}
		}

		connectionManager.setupTCPConnection(connection)
	}
}

func (connectionManager *TCPConnectionManager) createTCPServer(addrChannel chan string) {
	// Create a listener
	port := common.TCP_PORT + common.NameExtension
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	// Address from a listener is on the form [::]:xxxxx
	const portOffset = 4
	addrChannel <- listener.Addr().String()[portOffset:]

	go connectionManager.tcpListener(&listener)

	<-connectionManager.stopServerChannel
}

func (connectionManager *TCPConnectionManager) createTCPClient(address string) {
	timeout := time.Second

	for {
		connection, err := net.DialTimeout("tcp", address, timeout)

		// Check the error if it's of the type Network Operation Timeout.
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			continue // Retry!
		}
		if err != nil {
			fmt.Println("Error setting up connection: ", err)
			return
		}

		connectionManager.setupTCPConnection(connection)
		return
	}
}
