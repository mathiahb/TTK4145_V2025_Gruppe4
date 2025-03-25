package network

import (
	Constants "elevator_project/constants"
	"fmt"
	"sync"

	peer_to_peer "elevator_project/network/Peer_to_Peer"

	peers "Network-go/network/peers"
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

	comm chan Message

	//peers
	txEnable     chan bool
	peerUpdateCh chan peers.PeerUpdate

	mu_communication_channels sync.Mutex
	communication_channels    map[TxID]chan Message

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

		txEnable:     make(chan bool),
		peerUpdateCh: make(chan peers.PeerUpdate),

		comm:                   make(chan Message, 32), // Velg en passende bufferstørrelse
		communication_channels: make(map[TxID]chan Message),

		close_channel: make(chan bool),

		shared_state_communication: communication_channels,
	}

	go peers.Receiver(15647, node.peerUpdateCh)
	go peers.Transmitter(15647, name, node.txEnable)

	node.start_reader()
	node.start_dispatcher()

	return &node
}

func (node *Node) Connect() {
	node.protocol_dispatcher.Do_Synchronization()
}

func (node *Node) Close() {
	node.p2p.Close()
	close(node.close_channel)
}

func (node *Node) Broadcast(message Message) {
	node.p2p.Broadcast(message.p2p_message)
}

func (node *Node) Broadcast_Response(message Message, responding_to Message) {
	message.p2p_message.Depend_On(responding_to.p2p_message)
	node.p2p.Broadcast(message.p2p_message)
}

func (node *Node) protocol_timed_out() {
	node.Connect() // Reconnect to the others
}

func (node *Node) start_reader() {
	go node.reader()
}

func (node *Node) forward_To_Network(message Message) {
	node.mu_communication_channels.Lock()
	defer node.mu_communication_channels.Unlock()

	comm, ok := node.communication_channels[message.id]
	if ok {
		go func() { comm <- message }()
	}
}

func (node *Node) reader() {
	fmt.Printf("[%s]: Began reading on Node %s\n", node.name, node.name)

	for {
		select {
		case <-node.close_channel:
			return
		case commit := <-node.shared_state_communication.ToNetwork.TwoPhaseCommit.RequestCommit:
			fmt.Printf("[%s] Got command: %+v\n\n", node.name, commit)
			node.protocol_dispatcher.Do_Command(commit)
		case peerUpdate := <-node.peerUpdateCh:
			node.alive_nodes_manager.Set_Alive_Nodes(peerUpdate.Peers)
			node.protocol_dispatcher.Do_Synchronization()
			node.shared_state_communication.FromNetwork.Discovery.Updated_Alive_Nodes <- node.alive_nodes_manager.Get_Alive_Nodes()

		case p2p_message := <-node.p2p.Read_Channel:
			message := translate_Message(p2p_message)

			fmt.Printf("[%s] Received message: %s, decoded to \"%s: %s %s\"\n",
				node.name, p2p_message.Message, message.id, message.message_type, message.payload)

			switch message.message_type {
			// SYNCHRONIZATION
			case Constants.SYNC_REQUEST:
				go node.participate_In_Synchronization(message)

				// 2PC
			case Constants.PREPARE: // Received a synchronization request
				go node.participate_2PC(message)

			default:
				node.forward_To_Network(message)
			}
		}
	}
}

func (node *Node) Get_Alive_Nodes() []string {
	return node.alive_nodes_manager.Get_Alive_Nodes()
}
