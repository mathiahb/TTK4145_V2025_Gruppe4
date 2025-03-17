package network

import (
	"Constants"
	"fmt"
	"strings"
	"sync"

	peer_to_peer "github.com/mathiahb/TTK4145_V2025_Gruppe4/Network_Protocol/Network/Peer_to_Peer"
)

type Node struct {
	p2p *peer_to_peer.P2P_Network

	name string // Elevator ID

	voting_on          TxID
	mu_voting_resource sync.Mutex // TryLock to see if you can vote.
	coordinating       bool

	next_id int

	alive_nodes []string
	comm        chan Message // Kanal for å motta 3PC-meldinger
}

func New_Node(name string) Node {

	network := Node{
		p2p: peer_to_peer.New_P2P_Network(),

		name: name,

		coordinating: false,

		next_id: 0,

		alive_nodes: make([]string, 0),
		comm:        make(chan Message, 32), // Velg en passende bufferstørrelse
	}

	go network.reader()

	return network
}

func (node *Node) Broadcast(message Message) {
	p2p_message := node.p2p.Create_Message(message.String(), peer_to_peer.MESSAGE)
	node.p2p.Broadcast(p2p_message)
}

func (node *Node) Broadcast_Response(message Message, responding_to peer_to_peer.P2P_Message) {
	p2p_message := node.p2p.Create_Message(message.String(), peer_to_peer.MESSAGE)
	p2p_message.Depend_On(responding_to)
	node.p2p.Broadcast(p2p_message)
}

// func generateTxID() string {
// 	panic("unimplemented")
// }

func (node *Node) handleSYN(message peer_to_peer.P2P_Message) {
	// Gjør en vurdering på om heisen kan utføre endringen

	// Sjekk om kommandoen er gyldig
	// Etasjen er innenfor scopet feks
	// Heisen er i en tilstand hvor den kan utføre endringen

	canCommit := true
	// Deretter send PREPARE_ACK eller ABORT

	if canCommit {
		node.PREPARE_ACK()
	} else {
		node.ABORT()
	}
}
func (node *Node) doLocalCommit(cmd Command) {
	// Gjør endringen
}
func (node *Node) doLocalAbort() {
	fmt.Printf("[Local %s] Doing abort.\n", node.name)
}
func parseCommit(msg string) Command {
	// msg example: "COMMIT floor=3"

	// Drop "COMMIT " from the beginning.
	payload := strings.TrimPrefix(msg, "COMT ")
	// Now payload might be "floor=3"

	// Split on the first '='
	parts := strings.SplitN(payload, "=", 2)

	if len(parts) < 2 {
		fmt.Printf("Error: Could not parse commit message (not two parts): %s\n", msg)
	}

	return Command{
		Field:     parts[0],
		New_Value: parts[1],
	}
}

func (node *Node) reader() {
	for {
		p2p_message := <-node.p2p.Read_Channel
		message := Message_From_String(p2p_message.Message)

		switch message.message_type {

		case Constants.PREPARE: // Received a synchronization request
			node.handleSYN(p2p_message) // Decide whether to commit or abort

		case Constants.PREPARE_ACK: // Received a synchronization acknowledgement
			node.comm <- message

		case Constants.ABORT_COMMIT: // Received an abort commit message
			node.comm <- message
			// TODO: abort current synchronization
			continue

		case Constants.COMMIT: // Received a commit message
			cmd := parseCommit(p2p_message.Message)
			node.doLocalCommit(cmd)
			node.ACK()

		case Constants.ACK:
			node.comm <- message

			// node.active_vote.Add_Vote()
			// if node.active_vote.Is_Committable() {
			// 	node.COMMIT()
			// } else if node.active_vote.Is_Aborted() {
			// 	node.ABORT()
			// }
		}
	}
}

func (node *Node) shared_state_connection(message_from_shared_state chan Command, command_to_shared_state chan Command) {

}

func (node *Node) Get_Alive_Nodes() []string {
	return node.alive_nodes
}
