package peer_to_peer

import (
	"Constants"
	"fmt"
	"strings"
)

type P2P_Message_Type string

const (
	MESSAGE                    = "MSGSND"
	REQUEST_MISSING_DEPENDENCY = "REQDEP"
	HEARTBEAT                  = "HEARTB"
	REQUEST_LAMPORT_CLOCK      = "REQCLK"
)

type P2P_Message struct {
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
func (message P2P_Message) To_String() string {
	var result string = ""

	result = result + message.Sender
	result = result + Constants.P2P_FIELD_DELIMINATOR + string(message.Type)
	result = result + Constants.P2P_FIELD_DELIMINATOR + message.Time.String()
	result = result + Constants.P2P_FIELD_DELIMINATOR + message.dependency.To_String()
	result = result + Constants.P2P_FIELD_DELIMINATOR + message.Message

	return result
}

func P2P_Message_From_String(tcp_message string) P2P_Message {
	fields := strings.Split(tcp_message, Constants.P2P_FIELD_DELIMINATOR)

	if len(fields) != 5 {
		fmt.Printf("ERROR: P2P_Message badly formatted! Did you accidentally use \\r\\n in a file? %s\n", tcp_message)
		return P2P_Message{}
	}

	sender := fields[0]
	message_type := P2P_Message_Type(fields[1])
	clock := New_Lamport_Clock_From_String(fields[2])
	dependency := Dependency_From_String(fields[3])
	body := fields[4]

	return P2P_Message{
		Sender:     sender,
		Type:       message_type,
		Time:       clock,
		dependency: dependency,
		Message:    body,
	}
}
