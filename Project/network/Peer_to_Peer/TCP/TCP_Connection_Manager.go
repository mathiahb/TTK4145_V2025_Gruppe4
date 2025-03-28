package TCP

import (
	"elevator_project/common"
	"fmt"
	"sync"
)

// TCP Connection Manager
// Implements creating TCP servers and clients and storing them onto a connection.
// Automatic disconnection is handled by TCP Handler
type TCPConnectionManager struct {
	mu          sync.Mutex
	Connections map[string]TCPConnection // TCP Connection defined in TCP_internal.go

	GlobalReadChannel chan string
	stopServerChannel chan bool
}

func NewTCPConnectionManager() *TCPConnectionManager {
	return &TCPConnectionManager{
		Connections:       make(map[string]TCPConnection),
		GlobalReadChannel: make(chan string, common.TCP_BUFFER_SIZE),
		stopServerChannel: make(chan bool),
	}
}

func (manager *TCPConnectionManager) AddConnection(connection TCPConnection) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.doesConnectionExistUnsafe(connection.connectionName) {
		connection.Close()
		return
	}

	manager.Connections[connection.connectionName] = connection
	fmt.Printf("Connection %s added!\n", connection.connectionName)
}

func (manager *TCPConnectionManager) RemoveConnection(connection TCPConnection) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	delete(manager.Connections, connection.connectionName)
	fmt.Printf("Connection %s removed!\n", connection.connectionName)
}

func (manager *TCPConnectionManager) doesConnectionExistUnsafe(connectionName string) bool {
	_, ok := manager.Connections[connectionName]
	return ok
}

func (manager *TCPConnectionManager) DoesConnectionExist(connectionName string) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.doesConnectionExistUnsafe(connectionName)
}

func (manager *TCPConnectionManager) CloseAll() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	close(manager.stopServerChannel)

	for _, connection := range manager.Connections {
		connection.Close()
	}
}

func (manager *TCPConnectionManager) OpenServer() string {
	addressChannel := make(chan string)

	go manager.createTCPServer(addressChannel)

	return <-addressChannel
}

func (manager *TCPConnectionManager) ConnectClient(address string) {
	go manager.createTCPClient(address)
}

func (manager *TCPConnectionManager) Broadcast(message string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for _, connection := range manager.Connections {
		connection.WriteChannel <- message
	}
}

func (manager *TCPConnectionManager) Send(message string, recipient string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if !manager.doesConnectionExistUnsafe(recipient) {
		fmt.Printf("Error, connection %s did not exist! Failed sending %s\n", recipient, message)
		manager.ConnectClient(recipient)
		return
	}

	manager.Connections[recipient].WriteChannel <- message
}
