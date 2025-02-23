package TCP

import (
	"Constants"
	"fmt"
	"sync"
)

// TCP Connection Manager
// Implements creating TCP servers and clients and storing them onto a connection.
// Automatic disconnection is handled by TCP Handler

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
	go manager.create_TCP_Server(port)
}

func (manager *TCP_Connection_Manager) Connect_Client(address string) {
	go manager.create_TCP_Client(address)
}

func (manager *TCP_Connection_Manager) Broadcast(message string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for _, connection := range manager.Connections {
		connection.Write_Channel <- message
	}
}
