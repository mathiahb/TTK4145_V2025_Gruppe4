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

type TCP_Connection struct {
	// Public
	Write_Channel chan string
	ReadChannel   chan string

	// Protected
	closeChannel    chan bool
	connection_name string
	connection      net.Conn

	split_handler TCP_Split_Handler

	failed_writes    int
	watchdog_timer   *time.Timer
	heartbeat_ticker *time.Ticker
}

// Creates a new TCP conneciton bound to a shared read channel.
func New_TCP_Connection(name string, readChannel chan string, connection net.Conn) TCP_Connection {
	write_channel := make(chan string, common.TCP_BUFFER_SIZE)
	closeChannel := make(chan bool)

	return TCP_Connection{
		Write_Channel: write_channel,
		ReadChannel:   readChannel,
		closeChannel:  closeChannel,

		connection_name: name,
		connection:      connection,

		split_handler: New_TCP_Split_Handler(), // New connection, no split read message yet.

		failed_writes:    0,
		watchdog_timer:   time.NewTimer(common.TCP_HEARTBEAT_TIME),
		heartbeat_ticker: time.NewTicker(common.TCP_RESEND_HEARTBEAT),
	}
}

func (connection TCP_Connection) Close() {
	select {
	case <-connection.closeChannel:
		// Connection was already closed!
		fmt.Printf("Error: Attempted to close closed channel %s!\n", connection.connection_name)
		return
	default:
		close(connection.closeChannel)
	}
}

func (connection TCP_Connection) Get_Name() string {
	return connection.connection_name
}
