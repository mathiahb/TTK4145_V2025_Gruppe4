package network

import (
	Constants "elevator_project/constants"
	peer_to_peer "elevator_project/network/Peer_to_Peer"
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

	p2p_message peer_to_peer.P2P_Message
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

func translate_Message(p2p_message peer_to_peer.P2P_Message) Message {
	result := Message{}
	result.p2p_message = p2p_message

	message := p2p_message.Message

	split := strings.SplitN(message, "=", 4) // Returns [garbage, txid + space + s, sender + space + r, payload]
	if len(split) != 4 {
		fmt.Printf("ERROR: Badly formatted message! %s\n", p2p_message.To_String())
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
	id := strconv.Itoa(node.next_TxID_number)
	node.next_TxID_number++

	return TxID(node.name + ":" + id)
}

func (node *Node) isTxIDFromUs(id TxID) bool {
	return strings.HasPrefix(string(id), node.name)
}

func (node *Node) create_Message(message_type string, id TxID, message string) Message {
	result := Message{
		message_type: message_type,
		id:           id,
		sender:       node.name,
		payload:      message,
	}

	result.p2p_message = node.p2p.Create_Message(result.String())

	return result
}

func (node *Node) create_Vote_Message(message_type string, message string) Message {
	return node.create_Message(message_type, node.generateTxID(), message)
}
