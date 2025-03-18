package network

import (
	"Constants"
	"fmt"
	"strconv"
	"strings"
)

// Name:ID
type TxID string

type Message struct {
	message_type string
	id           TxID
	sender       string
	payload      string
}

// Global actions to interact with messages

// Format: message_type txid=id r=payload
func (message Message) String() string {
	return fmt.Sprintf("%s txid=%s s=%s r=%s",
		message.message_type,
		message.id,
		message.sender,
		message.payload,
	)
}

func Message_From_String(message string) Message {
	result := Message{}

	split := strings.SplitN(message, "=", 4) // Returns [garbage, txid + space + s, sender + space + r, payload]
	if len(split) != 4 {
		fmt.Printf("ERROR: Badly formatted message! %s\n", message)
		return result
	}

	result.message_type = message[0:Constants.SIZE_TYPE_FIELD]
	id := split[1][0 : len(split[1])-2]     // Remove last 2 letters since they are a space and r.
	sender := split[2][0 : len(split[2])-2] // Remove last 2 letters since they are a space and s.
	result.id = TxID(id)
	result.sender = sender
	result.payload = split[3]

	return result
}

// Node actions to create a message to be sent

func (node *Node) generateTxID() TxID {
	id := strconv.Itoa(node.next_id)
	node.next_id++

	return TxID(node.name + ":" + id)
}

func (node *Node) isTxIDFromUs(id TxID) bool {
	return strings.HasPrefix(string(id), node.name)
}

func (node *Node) create_Message(message_type string, id TxID, message string) Message {
	return Message{
		message_type: message_type,
		id:           id,
		sender:       node.name,
		payload:      message,
	}
}

func (node *Node) create_Vote_Message(message_type string, message string) Message {
	return node.create_Message(message_type, node.generateTxID(), message)
}
