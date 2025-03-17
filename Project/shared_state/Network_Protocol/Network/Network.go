package network

import (
	"Constants"
	peer_to_peer "Network-Protocol/Peer_to_Peer"
	"fmt"
	"strings"
	"sync"
)

type Node struct {
	p2p *peer_to_peer.P2P_Network

	name string // Elevator ID

	active_voter_names []string

	voting_on   string
	busy_voting sync.Mutex // TryLock to see if you can vote.

	alive_nodes []string
	comm        chan string // Kanal for å motta 3PC-meldinger
}

func New_Node(name string) Node {

	network := Node{
		p2p: peer_to_peer.New_P2P_Network(),

		name: name,

		active_voter_names: make([]string, 0),

		alive_nodes: make([]string, 0),
		comm:        make(chan string, 32), // Velg en passende bufferstørrelse
	}

	go network.reader()

	return network
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
		message := <-node.p2p.Read_Channel
		message_type := message.Message[0:Constants.SIZE_TYPE_FIELD]

		switch message_type {

		case Constants.PREPARE: // Received a synchronization request
			node.handleSYN(message) // Decide whether to commit or abort

		case Constants.PREPARE_ACK: // Received a synchronization acknowledgement
			node.comm <- Constants.PREPARE_ACK

		case Constants.ABORT_COMMIT: // Received an abort commit message
			node.comm <- Constants.ABORT_COMMIT
			// TODO: abort current synchronization
			continue

		case Constants.COMMIT: // Received a commit message
			cmd := parseCommit(message.Message)
			node.doLocalCommit(cmd)
			node.ACK()

		case Constants.ACK:
			node.comm <- Constants.ACK

			// node.active_vote.Add_Vote()
			// if node.active_vote.Is_Committable() {
			// 	node.COMMIT()
			// } else if node.active_vote.Is_Aborted() {
			// 	node.ABORT()
			// }
		}
	}
}

func (node *Node) shared_state_connection(message_from_shared_state chan string, command_to_shared_state chan string) {

}

func (node *Node) poll_for_alive_nodes() {

}

func (node *Node) Get_Alive_Nodes() []string {
	return node.alive_nodes
}
