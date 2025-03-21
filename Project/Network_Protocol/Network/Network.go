package network

import (
	Constants "elevator_project/constants"
	"fmt"
	"sync"

	peer_to_peer "elevator_project/Network_Protocol/Network/Peer_to_Peer"
)

type Node struct {
	p2p *peer_to_peer.P2P_Network

	name             string // Elevator ID
	next_TxID_number int

	mu_voting_resource sync.Mutex // TryLock to see if you can vote.

	alive_nodes_manager AliveNodeManager
	protocol_dispatcher ProtocolDispatcher

	comm chan Message // Kanal for å motta 3PC-meldinger

	close_channel chan bool

	// Shared State connection
	new_alive_nodes          chan []string
	synchronization_channels SynchronizationChannels
}

func New_Node(name string, new_alive_nodes_channel chan []string, synchronization_channels SynchronizationChannels) *Node {

	node := Node{
		p2p: peer_to_peer.New_P2P_Network(),

		name: name,

		next_TxID_number: 0,

		alive_nodes_manager: AliveNodeManager{
			alive_nodes: make([]string, 0),
		},
		protocol_dispatcher: *New_Protocol_Dispatcher(),

		comm: make(chan Message, 32), // Velg en passende bufferstørrelse

		close_channel: make(chan bool),

		new_alive_nodes:          new_alive_nodes_channel,
		synchronization_channels: synchronization_channels,
	}

	node.start_reader()
	node.start_dispatcher()

	return &node
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

func (node *Node) protocol_timed_out() {
	node.protocol_dispatcher.should_do_discovery <- true
	node.protocol_dispatcher.should_do_synchronize <- true
}

func (node *Node) start_reader() {
	go node.reader()
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
				node.comm <- message

			case Constants.DISCOVERY_COMPLETE:
				node.comm <- message

			// SYNCHRONIZATION
			case Constants.SYNC_AFTER_DISCOVERY:
				go node.participate_In_Synchronization(p2p_message, message.id)

			case Constants.SYNC_RESPONSE:
				node.comm <- message

			case Constants.SYNC_RESULT:
				node.comm <- message

				// 2PC
			case Constants.PREPARE: // Received a synchronization request
				go node.participate_2PC(p2p_message, message)

			case Constants.PREPARE_ACK: // Received a synchronization acknowledgement
				node.comm <- message

			case Constants.ABORT_COMMIT: // Received an abort commit message
				node.comm <- message
				// TODO: abort current synchronization
				continue

			case Constants.COMMIT: // Received a commit message
				node.comm <- message
				//node.ACK()

			case Constants.ACK:
				node.comm <- message
			}
		}
	}
}

func (node *Node) Get_Alive_Nodes() []string {
	return node.alive_nodes_manager.Get_Alive_Nodes()
}
