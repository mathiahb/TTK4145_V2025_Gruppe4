package TCP

import (
	"Constants"
	"fmt"
	"io"
	"strings"
	"time"
)

// Protected functions

// Return: Null-terminated splits of tcp_message which has handled partial TCP transmission.
func (TCP_Connection *TCP_Connection) split_null_terminated_tcp_message(tcp_message string) []string {
	split_messages := strings.Split(tcp_message, "\000") // Null terminated

	// Handle message that has been split due to buffer size or partial TCP transmission.
	//---
	// Add previous split message to the first message:
	split_messages[0] = TCP_Connection.split_message_to_be_handled + split_messages[0]

	// Store next split message to be handled on next read:
	last_split_id := len(split_messages) - 1
	TCP_Connection.split_message_to_be_handled = split_messages[last_split_id]

	//---
	return split_messages[0:last_split_id]
}

// Reads a string from a net.Conn onto a read channel
func (TCP_Connection *TCP_Connection) read() {
	deadline := time.Now().Add(Constants.TCP_READ_DEADLNE)
	TCP_Connection.connection.SetReadDeadline(deadline)

	data := make([]byte, 4096)

	bytes_received, err := TCP_Connection.connection.Read(data)

	if err == io.EOF {
		// Connection terminated from the other side. Disconnect.
		TCP_Connection.Close()
	}

	if err == nil {
		message := string(data[0:bytes_received])

		split_messages := TCP_Connection.split_null_terminated_tcp_message(message)

		for _, split_message := range split_messages {
			TCP_Connection.Read_Channel <- split_message
		}
	}
}

// Writes a string to a net.Conn
func (TCP_Connection *TCP_Connection) write(message string) {
	deadline := time.Now().Add(Constants.TCP_READ_DEADLNE)
	TCP_Connection.connection.SetWriteDeadline(deadline)

	data := []byte(message)
	data = append(data, '\000') // Null terminated

	_, err := TCP_Connection.connection.Write(data)

	if err != nil {
		fmt.Println("Write didn't succeed, error: ", err)
	}
}

// Handles a TCP Connection, writing any data from the Write Channel onto the connection.
// And reads any data from the connection onto the Read Channel.
//
// Will self-remove from the connection manager should close be called.
func (connection *TCP_Connection) handle_TCP_Connection(connection_manager *TCP_Connection_Manager) {
	defer connection.connection.Close()
	defer connection_manager.Remove_Connection(*connection)

	when_to_read_ticker := time.NewTicker(Constants.TCP_WAIT_BEFORE_READING_AGAIN)

	for {
		select {
		case <-when_to_read_ticker.C:
			connection.read()

		case message := <-connection.Write_Channel:
			connection.write(message)

		case <-connection.close_channel:
			return
		}

		time.Sleep(time.Microsecond)
	}
}
