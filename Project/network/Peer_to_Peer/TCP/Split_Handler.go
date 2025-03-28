package TCP

import (
	"elevator_project/common"
	"strings"
)

// Struct that handles the fact that TCP messages can be cut short in a read statement.
type TCPSplitHandler struct {
	splitMessageToBeHandled string
}

func (handler *TCPSplitHandler) SplitNullTerminatedTCPMessage(tcpMessage string) []string {
	splitMessages := strings.Split(tcpMessage, common.NULL)

	// Handle message that has been split due to buffer size or partial TCP transmission.
	//---
	// Add previous split message to the first message:
	splitMessages[0] = handler.splitMessageToBeHandled + splitMessages[0]

	// Store next split message to be handled on next read:
	lastSplitID := len(splitMessages) - 1
	handler.splitMessageToBeHandled = splitMessages[lastSplitID]

	//---
	// Return all except the incomplete split.
	return splitMessages[0:lastSplitID]
}

func (handler *TCPSplitHandler) MakeNullTerminatedTCPMessage(message string) string {
	return message + common.NULL
}

func NewTCPSplitHandler() TCPSplitHandler {
	return TCPSplitHandler{
		splitMessageToBeHandled: "",
	}
}
