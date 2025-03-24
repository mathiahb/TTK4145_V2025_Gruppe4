package network

import (
	Constants "elevator_project/constants"
	"fmt"
	"sync"

	peer_to_peer "elevator_project/network/Peer_to_Peer"
)

type CommunicationToNetwork struct {
	Discovery struct {
		// Nothing
	}

	Synchronization struct {
		RespondToInformationRequest chan string
		RespondWithInterpretation   chan string
	}

	TwoPhaseCommit struct {
		RequestCommit chan string
	}
}

type CommunicationFromNetwork struct {
	Discovery struct {
		Updated_Alive_Nodes chan []string
	}

	Synchronization struct {
		ProtocolRequestInformation     chan bool
		ProtocolRequestsInterpretation chan map[string]string
		ResultFromSynchronization      chan string
	}

	TwoPhaseCommit struct {
		ProtocolCommited chan string
	}
}

type NetworkCommunicationChannels struct {
	ToNetwork   CommunicationToNetwork
	FromNetwork CommunicationFromNetwork
}

type Node struct {
	p2p *peer_to_peer.P2P_Network

	name             string // Elevator ID
	next_TxID_number int

	mu_voting_resource sync.Mutex // TryLock to see if you can vote.

	alive_nodes_manager AliveNodeManager
	protocol_dispatcher ProtocolDispatcher

	comm       chan Message
	twopc_comm map[TxID]chan Message

	close_channel chan bool

	// Shared State connection
	shared_state_communication NetworkCommunicationChannels
}

func New_Node(name string, communication_channels NetworkCommunicationChannels) *Node {

	node := Node{
		p2p: peer_to_peer.New_P2P_Network(),

		name: name,

		next_TxID_number: 0,

		alive_nodes_manager: AliveNodeManager{
			alive_nodes: make([]string, 0),
		},
		protocol_dispatcher: *New_Protocol_Dispatcher(),

		comm:       make(chan Message, 32), // Velg en passende bufferstørrelse
		twopc_comm: make(map[TxID]chan Message),

		close_channel: make(chan bool),

		shared_state_communication: communication_channels,
	}

	node.start_reader()
	node.start_dispatcher()

	return &node
}

func (node *Node) Connect() {
	node.protocol_dispatcher.Do_Discovery()
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
	node.Connect() // Reconnect to the others
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
		case commit := <-node.shared_state_communication.ToNetwork.TwoPhaseCommit.RequestCommit:
			node.protocol_dispatcher.Do_Command(commit)
		case p2p_message := <-node.p2p.Read_Channel:
			message := Message_From_String(p2p_message.Message)

			fmt.Printf("[%s] Received message: %s, decoded to \"%s: %s %s\"\n",
				node.name, p2p_message.Message, message.id, message.message_type, message.payload)

			switch message.message_type {
			// DISCOVERY
			case Constants.DISCOVERY_BEGIN:
				go node.participate_In_Discovery(p2p_message, message.id)

			case Constants.DISCOVERY_HELLO:
				go func() { node.comm <- message }()

			case Constants.DISCOVERY_COMPLETE:
				go func() { node.comm <- message }()

			// SYNCHRONIZATION
			//case Constants.SYNC_AFTER_DISCOVERY:
			//	go node.participate_In_Synchronization(p2p_message, message.id)

			case Constants.SYNC_RESPONSE:
				go func() { node.comm <- message }()

			case Constants.SYNC_RESULT:
				go func() { node.comm <- message }()

				// 2PC
			case Constants.PREPARE: // Received a synchronization request
				go node.participate_2PC(p2p_message, message)

			case Constants.PREPARE_ACK: // Received a synchronization acknowledgement
				comm, ok := node.twopc_comm[message.id]
				if ok {
					go func() { comm <- message }()
				}

			case Constants.ABORT_COMMIT: // Received an abort commit message
				comm, ok := node.twopc_comm[message.id]
				if ok {
					go func() { comm <- message }()
				} else {
					go func() { node.comm <- message }()
				}

			case Constants.COMMIT: // Received a commit message
				comm, ok := node.twopc_comm[message.id]
				if ok {
					go func() { comm <- message }()
				}

			case Constants.ACK:
				comm, ok := node.twopc_comm[message.id]
				if ok {
					go func() { comm <- message }()
				}
			}
		}
	}
}

func (node *Node) Get_Alive_Nodes() []string {
	return node.alive_nodes_manager.Get_Alive_Nodes()
}
