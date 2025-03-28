package network

import (
	"elevator_project/common"
	peerToPeer "elevator_project/network/Peer_to_Peer"
	"fmt"
	"strconv"
	"strings"
)

// Name:ID
type TxID string

type Message struct {
	messageType string
	id          TxID
	sender      string
	payload     string

	p2pMessage peerToPeer.P2PMessage
}

// Global actions to interact with messages

// Format: messageType txid=id r=payload
func (message Message) String() string {
	return fmt.Sprintf("%s txid=%s s=%s r=%s",
		message.messageType,
		message.id,
		message.sender,
		message.payload,
	)
}

func translateMessage(p2pMessage peerToPeer.P2PMessage) Message {
	result := Message{}
	result.p2pMessage = p2pMessage

	message := p2pMessage.Message

	split := strings.SplitN(message, "=", 4) // Returns [garbage, txid + space + s, sender + space + r, payload]
	if len(split) != 4 {
		fmt.Printf("ERROR: Badly formatted message! %s\n", p2pMessage.To_String())
		return result
	}

	result.messageType = message[0:common.SIZE_TYPE_FIELD]
	id := split[1][0 : len(split[1])-2]     // Remove last 2 letters since they are a space and r.
	sender := split[2][0 : len(split[2])-2] // Remove last 2 letters since they are a space and s.
	result.id = TxID(id)
	result.sender = sender
	result.payload = split[3]

	return result
}

// Node actions to create a message to be sent

func (node *Node) generateTxID() TxID {
	id := strconv.Itoa(node.nextTxIDNumber)
	node.nextTxIDNumber++

	return TxID(node.name + ":" + id)
}

func (node *Node) isTxIDFromUs(id TxID) bool {
	return strings.HasPrefix(string(id), node.name)
}

func (node *Node) createMessage(messageType string, id TxID, message string) Message {
	result := Message{
		messageType: messageType,
		id:          id,
		sender:      node.name,
		payload:     message,
	}

	result.p2pMessage = node.p2p.CreateMessage(result.String())

	return result
}

func (node *Node) createVoteMessage(messageType string, message string) Message {
	return node.createMessage(messageType, node.generateTxID(), message)
}
