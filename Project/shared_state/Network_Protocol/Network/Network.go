package network

import (
	"Constants"
	protocols "Network-Protocol/Network/Protocols"
	peer_to_peer "Network-Protocol/Peer_to_Peer"
)

type Node struct {
	p2p *peer_to_peer.P2P_Network

	name string // Elevator ID

	active_voter_names []string

	// When to Synchronize:
	// When you see someone who shouldn't be there
	// When you suspect someone is disconnected
	// Every so often (A few times a second?)
	active_synchronization     *protocols.Synchronization_Vote
	has_active_synchronization bool

	alive_nodes []string
}

func New_Node(name string) Node {

	network := Node{
		p2p: peer_to_peer.New_P2P_Network(),

		name: name,

		active_voter_names: make([]string, 0),

		active_synchronization:     &protocols.Synchronization_Vote{},
		has_active_synchronization: false,
		alive_nodes: make([]string, 0),
	}

	go network.reader()

	return network
}

func (node Node) reader() {
	for{
		message := <-node.p2p.Read_Channel

		message_type := message.Message[0:4]

		switch message_type{
		case Constants.SYNC_MESSAGE:

		case Constants.ACKNOWLEDGE:
			node.active_vote.Add_Vote()
			if node.active_vote.Is_Committable(){
				node.COMMIT()
			} else if node.active_vote.Is_Aborted(){
				node.ABORT()
			}
		}
	}
}

func (node Node) shared_state_connection(message_from_shared_state chan string, command_to_shared_state chan string){

}


func (node Node) poll_for_alive_nodes() {
	
}

func (node Node) Get_Alive_Nodes() []string{
	return node.alive_nodes
}