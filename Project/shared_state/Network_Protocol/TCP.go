package Network_Protocol

import (
	"fmt"
	"net"
	"sync"

	"github.com/mathiahb/TTK4145_V2025_Gruppe4/Constants"
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

func New_TCP_Connection(name string, connection net.Conn) TCP_Connection {
	write_channel := make(chan string, Constants.TCP_BUFFER_SIZE)
	read_channel := make(chan string, Constants.TCP_BUFFER_SIZE)
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
	close(connection.close_channel)
}

func (connection TCP_Connection) Get_Name() string {
	return connection.connection_name
}

// ------------------
// Connection Manager

type TCP_Connection_Manager struct {
	mu          sync.Mutex
	Connections map[string]TCP_Connection // TCP Connection defined in TCP_internal.go
}

func New_TCP_Connection_Manager() *TCP_Connection_Manager {
	return &TCP_Connection_Manager{
		Connections: make(map[string]TCP_Connection),
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

func (manager *TCP_Connection_Manager) Open_Server() {
	go create_TCP_Server(manager)
}

func (manager *TCP_Connection_Manager) Connect_Client(address string) {
	go create_TCP_Client(address, manager)
}
