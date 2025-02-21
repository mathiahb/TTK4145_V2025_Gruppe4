package TCP

import (
	"fmt"
	"net"
	"sync"

	"Constants"
)

// Implements a TCP framework
// Provides the option to open a TCP channel to an IP
// A TCP channel is represented by a send and receive channel.
// To close a TCP channel use the function close(CHANNEL)

// --------------------------------------------------------------------
// Connection

type TCP_Connection struct {
	// Public
	Write_Channel chan string
	Read_Channel  chan string

	// Protected
	close_channel   chan bool
	connection_name string
	connection      net.Conn
}

// Creates a new TCP conneciton bound to a shared read channel.
func New_TCP_Connection(name string, read_channel chan string, connection net.Conn) TCP_Connection {
	write_channel := make(chan string, Constants.TCP_BUFFER_SIZE)
	close_channel := make(chan bool)

	return TCP_Connection{
		Write_Channel: write_channel,
		Read_Channel:  read_channel,
		close_channel: close_channel,

		connection_name: name,
		connection:      connection,
	}
}

func (connection TCP_Connection) Close() {
	select {
	case <-connection.close_channel:
		// Connection was already closed!
		fmt.Printf("Error: Attempted to close closed channel %s!\n", connection.connection_name)
		return
	default:
		close(connection.close_channel)
	}
}

func (connection TCP_Connection) Get_Name() string {
	return connection.connection_name
}

// ------------------
// Connection Manager

type TCP_Connection_Manager struct {
	mu          sync.Mutex
	Connections map[string]TCP_Connection // TCP Connection defined in TCP_internal.go

	Global_Read_Channel chan string
}

func New_TCP_Connection_Manager() *TCP_Connection_Manager {
	return &TCP_Connection_Manager{
		Connections:         make(map[string]TCP_Connection),
		Global_Read_Channel: make(chan string, Constants.TCP_BUFFER_SIZE),
	}
}

func (manager *TCP_Connection_Manager) Add_Connection(connection TCP_Connection) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.Connections[connection.connection_name] = connection
	fmt.Printf("Connection %s added!\n", connection.connection_name)
}

func (manager *TCP_Connection_Manager) Remove_Connection(connection TCP_Connection) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	delete(manager.Connections, connection.connection_name)
	fmt.Printf("Connection %s removed!\n", connection.connection_name)
}

func (manager *TCP_Connection_Manager) Close_All() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for _, connection := range manager.Connections {
		connection.Close()
	}
}

func (manager *TCP_Connection_Manager) Open_Server(port string) {
	go create_TCP_Server(port, manager)
}

func (manager *TCP_Connection_Manager) Connect_Client(address string) {
	go create_TCP_Client(address, manager)
}

func (manager *TCP_Connection_Manager) Broadcast(message string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for _, connection := range manager.Connections {
		connection.Write_Channel <- message
	}
}
