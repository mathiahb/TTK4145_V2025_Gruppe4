package network

import (
	"strconv"
	"testing"
	"time"

	peerToPeer "elevator_project/network/Peer_to_Peer"
)

func CreateTestNetworkCommunicationChannels() NetworkCommunicationChannels {
	return NetworkCommunicationChannels{
		ToNetwork: CommunicationToNetwork{
			Discovery: struct{}{},
			Synchronization: struct {
				RespondToInformationRequest chan string
				RespondWithInterpretation   chan string
			}{
				RespondToInformationRequest: make(chan string, 1),
				RespondWithInterpretation:   make(chan string, 1),
			},
			TwoPhaseCommit: struct{ RequestCommit chan string }{
				RequestCommit: make(chan string, 1),
			},
		},
		FromNetwork: CommunicationFromNetwork{
			Discovery: struct{ UpdatedAliveNodes chan []string }{
				UpdatedAliveNodes: make(chan []string, 1),
			},
			Synchronization: struct {
				ProtocolRequestInformation     chan bool
				ProtocolRequestsInterpretation chan map[string]string
				ResultFromSynchronization      chan string
			}{
				ProtocolRequestInformation:     make(chan bool, 1),
				ProtocolRequestsInterpretation: make(chan map[string]string, 1),
				ResultFromSynchronization:      make(chan string, 1),
			},
			TwoPhaseCommit: struct{ ProtocolCommited chan string }{
				ProtocolCommited: make(chan string, 1),
			},
		},
	}
}

// Test Discovery Protocol

func stringSlicesEqual(slice1 []string, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := 0; i < len(slice1); i++ {
		if slice1[i] != slice2[i] {
			return false
		}
	}

	return true
}

func TestDiscovery(t *testing.T) {
	name1 := "Node1"
	name2 := "Node2"

	result1 := []string{name1, name2}
	result2 := []string{name2, name1}

	Node1 := NewNode(name1, CreateTestNetworkCommunicationChannels())
	Node2 := NewNode(name2, CreateTestNetworkCommunicationChannels()) // Node 2

	responseChannel1 := Node1.sharedStateCommunication.FromNetwork.Discovery.UpdatedAliveNodes
	responseChannel2 := Node2.sharedStateCommunication.FromNetwork.Discovery.UpdatedAliveNodes

	defer Node1.Close()
	defer Node2.Close()

	Node1.protocolDispatcher.DoSynchronization()

	time.Sleep(time.Millisecond * 150)

	// Is Node1 a result?
	select {
	case node1Result := <-responseChannel1:
		if !stringSlicesEqual(node1Result, result1) && !stringSlicesEqual(node1Result, result2) {
			t.Errorf("Node 1 result doesn't make sense! %s\n", node1Result)
		}
	default:
		t.Errorf("Received no response from Node 1.\n")
	}

	// Is Node2 a result?
	select {
	case node2Result := <-responseChannel2:
		if !stringSlicesEqual(node2Result, result1) && !stringSlicesEqual(node2Result, result2) {
			t.Errorf("Node 2 result doesn't make sense! %s\n", node2Result)
		}
	default:
		t.Errorf("Received no response from Node 2.\n")
	}
}

func TestDiscoveryMany(t *testing.T) {
	responseChannel := make(chan []string, 12)

	networkCommunication := CreateTestNetworkCommunicationChannels()
	networkCommunication.FromNetwork.Discovery.UpdatedAliveNodes = responseChannel

	Node1 := NewNode("Node0", networkCommunication)
	defer Node1.Close()

	name := "Node"
	for id := 1; id < 10; id++ {
		networkCommunication := CreateTestNetworkCommunicationChannels()
		networkCommunication.FromNetwork.Discovery.UpdatedAliveNodes = responseChannel

		Node := NewNode(name+strconv.Itoa(id), networkCommunication)
		defer Node.Close()
	}

	Node1.protocolDispatcher.DoSynchronization()

	time.Sleep(time.Millisecond * 150)

	var result []string

	select {
	case result = <-responseChannel:
	default:
		t.Fatalf("Didn't receive first response!")
	}

	for i := 1; i < 10; i++ {
		select {
		case other := <-responseChannel:
			if !stringSlicesEqual(result, other) {
				t.Errorf("Response %d does not match first response! %s != %s\n", i, result, other)
			}
		default:
			t.Errorf("Not enough responses! Expected 10, but got %d", i+1)
		}
	}

	select {
	case other := <-responseChannel:
		t.Errorf("Received too many responses! %s\n", other)
	default:
	}
}

func TestSynchronization(t *testing.T) {

	Node1 := NewNode("Node1", CreateTestNetworkCommunicationChannels())
	Node2 := NewNode("Node2", CreateTestNetworkCommunicationChannels())
	defer Node1.Close()
	defer Node2.Close()

	time.Sleep(time.Millisecond * 100)

	Node1.protocolDispatcher.DoSynchronization()

	time.Sleep(time.Millisecond * 150)

	info1 := "Node1 Hi"
	info2 := "Node2 Hello"

	select {
	case <-Node1.sharedStateCommunication.FromNetwork.Synchronization.ProtocolRequestInformation:
		Node1.sharedStateCommunication.ToNetwork.Synchronization.RespondToInformationRequest <- info1
	default:
		t.Fatalf("Node 1 did not request Information!")
	}

	select {
	case <-Node2.sharedStateCommunication.FromNetwork.Synchronization.ProtocolRequestInformation:
		Node2.sharedStateCommunication.ToNetwork.Synchronization.RespondToInformationRequest <- info2
	default:
		t.Fatalf("Node 2 did not request Information!")
	}

	time.Sleep(time.Millisecond * 10)

	successMessage := "Success"

	select {
	case results := <-Node1.sharedStateCommunication.FromNetwork.Synchronization.ProtocolRequestsInterpretation:
		if results[Node1.name] != info1 {
			t.Errorf("Results in Node 1 did not match sent info! %s != %s\n", results[Node1.name], info1)
		}

		if results[Node2.name] != info2 {
			t.Errorf("Results in Node 2 did not match sent info! %s != %s\n", results[Node2.name], info2)
		}

		Node1.sharedStateCommunication.ToNetwork.Synchronization.RespondWithInterpretation <- successMessage
	default:
		t.Fatalf("Node 1 never asked for interpretation!")
	}

	time.Sleep(time.Millisecond * 10)

	select {
	case result := <-Node1.sharedStateCommunication.FromNetwork.Synchronization.ResultFromSynchronization:
		if result != successMessage {
			t.Errorf("Result returned by Node 1 is not the success message! %s != %s\n", result, successMessage)
		}
	default:
		t.Errorf("Did not receive result from Node 1!")
	}

	select {
	case result := <-Node2.sharedStateCommunication.FromNetwork.Synchronization.ResultFromSynchronization:
		if result != successMessage {
			t.Errorf("Result returned by Node 2 is not the success message! %s != %s\n", result, successMessage)
		}
	default:
		t.Errorf("Did not receive result from Node 2!")
	}
}

func Test2PC(t *testing.T) {
	Node1 := NewNode("Node1", CreateTestNetworkCommunicationChannels())
	Node2 := NewNode("Node2", CreateTestNetworkCommunicationChannels())
	defer Node1.Close()
	defer Node2.Close()

	time.Sleep(time.Millisecond * 100)

	Node1.protocolDispatcher.DoSynchronization()

	time.Sleep(time.Millisecond * 150)
	<-Node1.sharedStateCommunication.FromNetwork.Synchronization.ProtocolRequestInformation
	Node1.sharedStateCommunication.ToNetwork.Synchronization.RespondToInformationRequest <- ""
	<-Node2.sharedStateCommunication.FromNetwork.Synchronization.ProtocolRequestInformation
	Node2.sharedStateCommunication.ToNetwork.Synchronization.RespondToInformationRequest <- ""
	<-Node1.sharedStateCommunication.FromNetwork.Synchronization.ProtocolRequestsInterpretation
	Node1.sharedStateCommunication.ToNetwork.Synchronization.RespondWithInterpretation <- ""

	// Do Test Here.
	command := "some command"

	Node1.protocolDispatcher.DoCommand(command)

	time.Sleep(time.Millisecond * 100)

	select {
	case result := <-Node1.sharedStateCommunication.FromNetwork.TwoPhaseCommit.ProtocolCommited:
		if result != command {
			t.Errorf("Received command came back malformed! %s != %s.", command, result)
		}
	default:
		t.Errorf("Did not receive command from Node1!")
	}

	select {
	case result := <-Node2.sharedStateCommunication.FromNetwork.TwoPhaseCommit.ProtocolCommited:
		if result != command {
			t.Errorf("Received command from Node 2 came back malformed! %s != %s.", command, result)
		}
	default:
		t.Errorf("Did not receive command from Node 2!")
	}
}

func testDiscoveryDispatchRetry(Node1 *Node, Node2 *Node, t *testing.T) {
	name1 := Node1.name
	name2 := Node2.name

	result1 := []string{name1, name2}
	result2 := []string{name2, name1}

	// Testing retry of Discovery
	Node1.protocolDispatcher.DoSynchronization()
	time.Sleep(time.Millisecond)

	p2pMessage := <-Node2.p2p.ReadChannel
	message := translateMessage(p2pMessage)
	Node2.abortSynchronization(message)

	go Node2.reader()

	time.Sleep(time.Millisecond * 150)

	// Is Node1 a result?
	select {
	case node1Result := <-Node1.sharedStateCommunication.FromNetwork.Discovery.UpdatedAliveNodes:
		if !stringSlicesEqual(node1Result, result1) && !stringSlicesEqual(node1Result, result2) {
			t.Fatalf("Node 1 result doesn't make sense! %s\n", node1Result)
		}
	default:
		t.Fatalf("Received no response from Node 1.\n")
	}

	// Is Node2 a result?
	select {
	case node2Result := <-Node2.sharedStateCommunication.FromNetwork.Discovery.UpdatedAliveNodes:
		if !stringSlicesEqual(node2Result, result1) && !stringSlicesEqual(node2Result, result2) {
			t.Fatalf("Node 2 result doesn't make sense! %s\n", node2Result)
		}
	default:
		t.Fatalf("Received no response from Node 2.\n")
	}

	t.Logf("Success Discovery redispatch")

	time.Sleep(time.Millisecond * 150)

	info1 := "Node1 Hi"
	info2 := "Node2 Hello"

	select {
	case <-Node1.sharedStateCommunication.FromNetwork.Synchronization.ProtocolRequestInformation:
		Node1.sharedStateCommunication.ToNetwork.Synchronization.RespondToInformationRequest <- info1
	default:
		t.Fatalf("Node 1 did not request Information!")
	}

	select {
	case <-Node2.sharedStateCommunication.FromNetwork.Synchronization.ProtocolRequestInformation:
		Node2.sharedStateCommunication.ToNetwork.Synchronization.RespondToInformationRequest <- info2
	default:
		t.Fatalf("Node 2 did not request Information!")
	}

	time.Sleep(time.Millisecond * 10)

	successMessage := "Success"

	select {
	case results := <-Node1.sharedStateCommunication.FromNetwork.Synchronization.ProtocolRequestsInterpretation:
		if results[Node1.name] != info1 {
			t.Errorf("Results in Node 1 did not match sent info! %s != %s\n", results[Node1.name], info1)
		}

		if results[Node2.name] != info2 {
			t.Errorf("Results in Node 2 did not match sent info! %s != %s\n", results[Node2.name], info2)
		}

		Node1.sharedStateCommunication.ToNetwork.Synchronization.RespondWithInterpretation <- successMessage
	default:
		t.Fatalf("Node 1 never asked for interpretation!")
	}

	time.Sleep(time.Millisecond * 10)

	select {
	case result := <-Node1.sharedStateCommunication.FromNetwork.Synchronization.ResultFromSynchronization:
		if result != successMessage {
			t.Errorf("Result returned by Node 1 is not the success message! %s != %s\n", result, successMessage)
		}
	default:
		t.Errorf("Did not receive result from Node 1!")
	}

	select {
	case result := <-Node2.sharedStateCommunication.FromNetwork.Synchronization.ResultFromSynchronization:
		if result != successMessage {
			t.Errorf("Result returned by Node 2 is not the success message! %s != %s\n", result, successMessage)
		}
	default:
		t.Errorf("Did not receive result from Node 2!")
	}
}

func TestDispatchRetry(t *testing.T) {
	name1 := "Node1"
	name2 := "Node2"

	Node1 := NewNode(name1, CreateTestNetworkCommunicationChannels())

	// Need manual control over node 2, not using NewNode that automatically starts a reader and dispatcher.
	Node2 := Node{
		p2p: peerToPeer.NewP2PNetwork(),

		name: name2,

		nextTxIDNumber: 0,

		aliveNodesManager: AliveNodeManager{
			aliveNodes: make([]string, 0),
		},
		protocolDispatcher: *NewProtocolDispatcher(),

		comm: make(chan Message, 32),

		closeChannel: make(chan bool),

		sharedStateCommunication: CreateTestNetworkCommunicationChannels(),
	}
	defer Node1.Close()
	defer Node2.Close()

	time.Sleep(time.Millisecond * 100)

	testDiscoveryDispatchRetry(Node1, &Node2, t)
}
