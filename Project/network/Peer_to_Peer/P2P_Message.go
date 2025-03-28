package peerToPeer

import (
	"elevator_project/common"
	"fmt"
	"strings"
)

type P2P_Message_Type string

const (
	MESSAGE                    P2P_Message_Type = "MSGSND"
	REQUEST_MISSING_DEPENDENCY P2P_Message_Type = "REQDEP"
)

type P2PMessage struct {
	Sender  string
	Type    P2P_Message_Type
	Message string
	Time    Lamport_Clock

	dependency Dependency
}

// Message format (double new line since body may use new lines):
// SENDER\r\n
// TYPE\r\n
// LAMPORT_CLOCK\r\n
// DEPENDENCY\r\n
// BODY/REQUEST\r\n
func New_P2P_Message(Sender string, Type P2P_Message_Type, Time Lamport_Clock, Message string) P2PMessage {
	return P2PMessage{
		Sender:  Sender,
		Type:    Type,
		Time:    Time,
		Message: Message,

		dependency: Dependency{},
	}
}

func (message *P2PMessage) DependOn(dependency_message P2PMessage) {
	message.dependency = New_Dependency(dependency_message.Sender, dependency_message.Time)
}

func (message P2PMessage) To_String() string {
	var result string = ""

	result = result + message.Sender
	result = result + common.P2P_FIELD_DELIMINATOR + string(message.Type)
	result = result + common.P2P_FIELD_DELIMINATOR + message.Time.String()
	result = result + common.P2P_FIELD_DELIMINATOR + message.dependency.To_String()
	result = result + common.P2P_FIELD_DELIMINATOR + message.Message

	return result
}

func P2P_Message_From_String(tcp_message string) P2PMessage {
	fields := strings.Split(tcp_message, common.P2P_FIELD_DELIMINATOR)

	if len(fields) != 5 {
		fmt.Printf("ERROR: P2PMessage badly formatted! Did you accidentally use \\r\\n in a file? %s\n", tcp_message)
		return P2PMessage{}
	}

	sender := fields[0]
	messageType := P2P_Message_Type(fields[1])
	clock := New_Lamport_Clock_From_String(fields[2])
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
