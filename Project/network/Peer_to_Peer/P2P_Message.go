package peerToPeer

import (
	"elevator_project/common"
	"fmt"
	"strings"
)

type P2PMessageType string

const (
	MESSAGE                    P2PMessageType = "MSGSND"
	REQUEST_MISSING_DEPENDENCY P2PMessageType = "REQDEP"
)

type P2PMessage struct {
	Sender  string
	Type    P2PMessageType
	Message string
	Time    LamportClock

	dependency Dependency
}

// Message format (double new line since body may use new lines):
// SENDER\r\n
// TYPE\r\n
// LAMPORT_CLOCK\r\n
// DEPENDENCY\r\n
// BODY/REQUEST\r\n
func NewP2PMessage(Sender string, Type P2PMessageType, Time LamportClock, Message string) P2PMessage {
	return P2PMessage{
		Sender:  Sender,
		Type:    Type,
		Time:    Time,
		Message: Message,

		dependency: Dependency{},
	}
}

func (message *P2PMessage) DependOn(dependencyMessage P2PMessage) {
	message.dependency = NewDependency(dependencyMessage.Sender, dependencyMessage.Time)
}

func (message P2PMessage) ToString() string {
	var result string = ""

	result = result + message.Sender
	result = result + common.P2P_FIELD_DELIMINATOR + string(message.Type)
	result = result + common.P2P_FIELD_DELIMINATOR + message.Time.String()
	result = result + common.P2P_FIELD_DELIMINATOR + message.dependency.ToString()
	result = result + common.P2P_FIELD_DELIMINATOR + message.Message

	return result
}

func P2PMessageFromString(tcpMessage string) P2PMessage {
	fields := strings.Split(tcpMessage, common.P2P_FIELD_DELIMINATOR)

	if len(fields) != 5 {
		fmt.Printf("ERROR: P2PMessage badly formatted! Did you accidentally use \\r\\n in a file? %s\n", tcpMessage)
		return P2PMessage{}
	}

	sender := fields[0]
	messageType := P2PMessageType(fields[1])
	clock := NewLamportClockFromString(fields[2])
	dependency := Dependency_From_String(fields[3])
	body := fields[4]

	return P2PMessage{
		Sender:     sender,
		Type:       messageType,
		Time:       clock,
		dependency: dependency,
		Message:    body,
	}
}
