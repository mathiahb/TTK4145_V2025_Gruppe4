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

	name    string // Elevator ID
	next_id int

	voting_on          TxID
	mu_voting_resource sync.Mutex // TryLock to see if you can vote.
	coordinating       bool

	alive_nodes_manager AliveNodeManager

	comm chan Message // Kanal for å motta 3PC-meldinger

	close_channel chan bool

	// Shared State connection
	new_alive_nodes chan []string
}

func New_Node(name string, new_alive_nodes_channel chan []string) *Node {

	network := Node{
		p2p: peer_to_peer.New_P2P_Network(),

		name: name,

		coordinating: false,

		next_id: 0,

		alive_nodes_manager: AliveNodeManager{
			alive_nodes: make([]string, 0),
		},

		comm: make(chan Message, 32), // Velg en passende bufferstørrelse

		close_channel: make(chan bool),

		new_alive_nodes: new_alive_nodes_channel,
	}

	go network.reader()

	return &network
}

func (node *Node) Close() {
	node.p2p.Close()
	close(node.close_channel)
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

func (node *Node) handleSYN(message peer_to_peer.P2P_Message) {
	// Gjør en vurdering på om heisen kan utføre endringen

	// Sjekk om kommandoen er gyldig
	// Etasjen er innenfor scopet feks
	// Heisen er i en tilstand hvor den kan utføre endringen

	canCommit := true
	// Deretter send PREPARE_ACK eller ABORT

	if canCommit {
		//node.PREPARE_ACK()
	} else {
		//node.ABORT()
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
	fmt.Printf("[%s]: Began reading on Node %s\n", node.name, node.name)

	for {
		select {
		case <-node.close_channel:
			return
		case p2p_message := <-node.p2p.Read_Channel:
			message := Message_From_String(p2p_message.Message)

			fmt.Printf("[%s] Received message: %s, decoded to \"%s: %s %s\"\n",
				node.name, p2p_message.Message, message.id, message.message_type, message.payload)

			switch message.message_type {
			// DISCOVERY
			case Constants.DISCOVERY_BEGIN:
				go node.participate_In_Discovery(p2p_message, message.id)

			case Constants.DISCOVERY_HELLO:
				fmt.Printf("[%s]: Decoded message %s to Hello! Node coordinating = %t\n", node.name, message.String(), node.coordinating)
				node.comm <- message

			case Constants.DISCOVERY_COMPLETE:
				node.comm <- message
				// SYNCHRONIZATION

			case Constants.SYNC_AFTER_DISCOVERY:
				go node.participate_In_Synchronization(p2p_message, message.id)

			case Constants.SYNC_RESPONSE:
				if node.coordinating {
					node.comm <- message
				}

			case Constants.SYNC_RESULT:
				node.comm <- message

				// 2PC
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
				//node.ACK()

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
}

func (node *Node) shared_state_connection(message_from_shared_state chan Command, command_to_shared_state chan Command) {

}

func (node *Node) Get_Alive_Nodes() []string {
	return node.alive_nodes_manager.Get_Alive_Nodes()
}
