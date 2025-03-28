package network

import (
	"elevator_project/common"
	"fmt"
	"sync"

	peerToPeer "elevator_project/network/Peer_to_Peer"

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
	p2p *peerToPeer.P2P_Network

	connectedToNetwork bool
	name               string // Elevator ID
	nextTxIDNumber     int

	muVotingResource sync.Mutex // TryLock to see if you can vote.

	aliveNodesManager  AliveNodeManager
	protocolDispatcher ProtocolDispatcher

	comm chan Message

	//peers
	txEnable     chan bool
	peerUpdateCh chan peers.PeerUpdate

	muCommunicationChannels sync.Mutex
	communicationChannels   map[TxID]chan Message

	closeChannel chan bool

	// Shared State connection
	sharedStateCommunication NetworkCommunicationChannels
}

func NewNode(name string, communicationChannels NetworkCommunicationChannels) *Node {

	node := Node{
		p2p: peerToPeer.NewP2PNetwork(),

		name:               name,
		connectedToNetwork: false,

		nextTxIDNumber: 0,

		aliveNodesManager: AliveNodeManager{
			aliveNodes: make([]string, 0),
		},
		protocolDispatcher: *NewProtocolDispatcher(),

		txEnable:     make(chan bool),
		peerUpdateCh: make(chan peers.PeerUpdate),

		comm:                  make(chan Message, 32), // Velg en passende bufferstørrelse
		communicationChannels: make(map[TxID]chan Message),

		closeChannel: make(chan bool),

		sharedStateCommunication: communicationChannels,
	}

	go peers.Receiver(common.PEERS_PORT, node.peerUpdateCh)
	go peers.Transmitter(common.PEERS_PORT, name, node.txEnable)

	node.startReader()
	node.startDispatcher()

	return &node
}

func (node *Node) Connect() {
	node.protocolDispatcher.DoSynchronization()
}

func (node *Node) Close() {
	node.p2p.Close()
	close(node.closeChannel)
}

func (node *Node) Broadcast(message Message) {
	node.p2p.Broadcast(message.p2pMessage)
}

func (node *Node) BroadcastResponse(message Message, respondingTo Message) {
	message.p2pMessage.DependOn(respondingTo.p2pMessage)
	node.p2p.Broadcast(message.p2pMessage)
}

func (node *Node) protocolTimedOut() {
	node.Connect() // Reconnect to the others
}

func (node *Node) startReader() {
	go node.reader()
}

func (node *Node) forwardToNetwork(message Message) {
	node.muCommunicationChannels.Lock()
	defer node.muCommunicationChannels.Unlock()

	comm, ok := node.communicationChannels[message.id]
	if ok {
		go func() { comm <- message }()
	}
}

func (node *Node) reader() {
	fmt.Printf("[%s]: Began reading on Node %s\n", node.name, node.name)

	for {
		select {
		case <-node.closeChannel:
			return

		case commit := <-node.sharedStateCommunication.ToNetwork.TwoPhaseCommit.RequestCommit:
			fmt.Printf("[%s] Got command: %+v\n\n", node.name, commit)

			if !node.connectedToNetwork {
				// We are not connected, just accept any change to shared state.
				node.sharedStateCommunication.FromNetwork.TwoPhaseCommit.ProtocolCommited <- commit
			} else {
				node.protocolDispatcher.DoCommand(commit)
			}

		case peerUpdate := <-node.peerUpdateCh:
			node.connectedToNetwork = false
			for _, peer := range peerUpdate.Peers {
				if peer == node.name {
					node.connectedToNetwork = true
				}
			}

			if !node.connectedToNetwork {
				peerUpdate.Peers = []string{node.name}
			} else {
				node.protocolDispatcher.DoSynchronization()
			}

			node.aliveNodesManager.SetAliveNodes(peerUpdate.Peers)
			node.sharedStateCommunication.FromNetwork.Discovery.Updated_Alive_Nodes <- node.aliveNodesManager.GetAliveNodes()

		case p2pMessage := <-node.p2p.ReadChannel:
			message := translateMessage(p2pMessage)

			fmt.Printf("[%s] Received message: %s, decoded to \"%s: %s %s\"\n",
				node.name, p2pMessage.Message, message.id, message.messageType, message.payload)

			switch message.messageType {
			// SYNCHRONIZATION
			case common.SYNC_REQUEST:
				go node.participateInSynchronization(message)

				// 2PC
			case common.PREPARE: // Received a synchronization request
				go node.participate2PC(message)

			default:
				node.forwardToNetwork(message)
			}
		}
	}
}

func (node *Node) GetAliveNodes() []string {
	return node.aliveNodesManager.GetAliveNodes()
}
