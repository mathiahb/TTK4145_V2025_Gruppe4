package TCP

import (
	"elevator_project/common"
	"fmt"
	"net"
	"time"
)

// Package TCP
// --------------------------------------------------------------------
//
// Implements TCP_Connection and TCP_Connection_Manager made by New_TCP_Connection(name, readChannel, connection)
// 															and NewTCPConnectionManager()
//
// TCP_Connection_Manager.OpenServer(PORT) opens a server on the PORT that automatically accepts
//		and adds any requested connection to the manager.
// TCP_Connection_Manager.ConnectClient(ADDRESS) attempts to connect a client to the address and
//		adds the resulting connection to the manager.
//
// TCP_Connection_Manager.CloseAll() closes all connections. To be used when crashing gracefully.
//
// TCP_Connection_Manager.Add_Connection(TCP_Connection) will add the connection to the manager.
// TCP_Connection_Manager.Remove_Connection(TCP_Connection) will remove the connection from the manager.
//
// READING:
// From an individually created TCP_Connection: message <- TCP_Connection.ReadChannel
// A connection made by TCP_Connection_Manager: message <- TCP_Connection_Manager.GlobalReadChannel
//
// WRITING:
// Writing to an individual connection: TCP_Connection.Write_Channel <- message
// Broadcasting to all connections: use TCP_Connection_Manager.Broadcast(message)
// --------------------------------------------------------------------
// Connection

type TCPConnection struct {
	// Public
	WriteChannel chan string
	ReadChannel  chan string

	// Protected
	closeChannel   chan bool
	connectionName string
	connection     net.Conn

	splitHandler TCPSplitHandler

	failedWrites    int
	watchdogTimer   *time.Timer
	heartbeatTicker *time.Ticker
}

// Creates a new TCP conneciton bound to a shared read channel.
func NewTCPConnection(name string, readChannel chan string, connection net.Conn) TCPConnection {
	writeChannel := make(chan string, common.TCP_BUFFER_SIZE)
	closeChannel := make(chan bool)

	return TCPConnection{
		WriteChannel: writeChannel,
		ReadChannel:  readChannel,
		closeChannel: closeChannel,

		connectionName: name,
		connection:     connection,

		splitHandler: NewTCPSplitHandler(), // New connection, no split read message yet.

		failedWrites:    0,
		watchdogTimer:   time.NewTimer(common.TCP_HEARTBEAT_TIME),
		heartbeatTicker: time.NewTicker(common.TCP_RESEND_HEARTBEAT),
	}
}

func (connection TCPConnection) Close() {
	select {
	case <-connection.closeChannel:
		// Connection was already closed!
		fmt.Printf("Error: Attempted to close closed channel %s!\n", connection.connectionName)
		return
	default:
		close(connection.closeChannel)
	}
}

func (connection TCPConnection) Get_Name() string {
	return connection.connectionName
}
