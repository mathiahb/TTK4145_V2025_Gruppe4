package TCP

import (
	"elevator_project/common"
	"fmt"
	"io"
	"time"
)

// Protected functions

// Reads a TCP message from the TCP Connection and puts all detected messages in the TCP message onto the Read Channel.
// Handles splitting messages using the TCP Split Handler.
func (TCPConnection *TCPConnection) read() {
	deadline := time.Now().Add(common.TCP_READ_DEADLNE)
	TCPConnection.connection.SetReadDeadline(deadline)

	data := make([]byte, 4096)

	bytesReceived, err := TCPConnection.connection.Read(data)

	if err == io.EOF {
		// Connection terminated from the other side. Disconnect.
		TCPConnection.Close()
	}

	if err == nil {
		message := string(data[0:bytesReceived])

		splitMessages := TCPConnection.splitHandler.SplitNullTerminatedTCPMessage(message)

		for _, splitMessage := range splitMessages {
			if splitMessage == common.TCP_HEARTBEAT {
				TCPConnection.watchdogTimer.Reset(common.TCP_HEARTBEAT_TIME)
			} else {
				TCPConnection.ReadChannel <- splitMessage
			}
		}
	}
}

// Writes a string onto the TCP Connection, function handles necessary
func (TCPConnection *TCPConnection) write(message string) {
	deadline := time.Now().Add(common.TCP_READ_DEADLNE)
	TCPConnection.connection.SetWriteDeadline(deadline)

	tcp_message := TCPConnection.splitHandler.MakeNullTerminatedTCPMessage(message)
	data := []byte(tcp_message)

	_, err := TCPConnection.connection.Write(data)

	if err != nil {
		fmt.Println("Write didn't succeed, error: ", err)
		TCPConnection.failedWrites++
	} else {
		TCPConnection.failedWrites = 0
	}

	if TCPConnection.failedWrites >= common.TCP_MAX_FAIL_WRITES {
		TCPConnection.Close()
	}
}

// Handles a TCP Connection, writing any data from the Write Channel onto the connection.
// And reads any data from the connection onto the Read Channel.
//
// Will self-remove from the connection manager should close be called.
func (connection *TCPConnection) handleTCPConnection(connectionManager *TCPConnectionManager) {
	defer connection.connection.Close()
	defer connectionManager.RemoveConnection(*connection)

	whenToReadTicker := time.NewTicker(common.TCP_WAIT_BEFORE_READING_AGAIN)

	for {
		select {
		case <-whenToReadTicker.C:
			connection.read()

		case message := <-connection.WriteChannel:
			connection.write(message)

		case <-connection.heartbeatTicker.C:
			connection.write(common.TCP_HEARTBEAT)

		case <-connection.watchdogTimer.C:
			return // No heartbeats heard within watchdog time, disconnect...

		case <-connection.closeChannel:
			return
		}

		time.Sleep(time.Microsecond)
	}
}
