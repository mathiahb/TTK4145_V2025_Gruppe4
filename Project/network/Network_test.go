package network

import (
	"strconv"
	"testing"
	"time"

	peer_to_peer "elevator_project/network/Peer_to_Peer"
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
			Discovery: struct{ Updated_Alive_Nodes chan []string }{
				Updated_Alive_Nodes: make(chan []string, 1),
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

	Node1 := New_Node(name1, CreateTestNetworkCommunicationChannels())
	Node2 := New_Node(name2, CreateTestNetworkCommunicationChannels()) // Node 2

	response_channel1 := Node1.shared_state_communication.FromNetwork.Discovery.Updated_Alive_Nodes
	response_channel2 := Node2.shared_state_communication.FromNetwork.Discovery.Updated_Alive_Nodes

	defer Node1.Close()
	defer Node2.Close()

	Node1.protocol_dispatcher.Do_Synchronization()

	time.Sleep(time.Millisecond * 150)

	// Is Node1 a result?
	select {
	case node1_result := <-response_channel1:
		if !stringSlicesEqual(node1_result, result1) && !stringSlicesEqual(node1_result, result2) {
			t.Errorf("Node 1 result doesn't make sense! %s\n", node1_result)
		}
	default:
		t.Errorf("Received no response from Node 1.\n")
	}

	// Is Node2 a result?
	select {
	case node2_result := <-response_channel2:
		if !stringSlicesEqual(node2_result, result1) && !stringSlicesEqual(node2_result, result2) {
			t.Errorf("Node 2 result doesn't make sense! %s\n", node2_result)
		}
	default:
		t.Errorf("Received no response from Node 2.\n")
	}
}

func TestDiscoveryMany(t *testing.T) {
	response_channel := make(chan []string, 12)

	networkCommunication := CreateTestNetworkCommunicationChannels()
	networkCommunication.FromNetwork.Discovery.Updated_Alive_Nodes = response_channel

	Node1 := New_Node("Node0", networkCommunication)
	defer Node1.Close()

	name := "Node"
	for id := 1; id < 10; id++ {
		networkCommunication := CreateTestNetworkCommunicationChannels()
		networkCommunication.FromNetwork.Discovery.Updated_Alive_Nodes = response_channel

		Node := New_Node(name+strconv.Itoa(id), networkCommunication)
		defer Node.Close()
	}

	Node1.protocol_dispatcher.Do_Synchronization()

	time.Sleep(time.Millisecond * 150)

	var result []string

	select {
	case result = <-response_channel:
	default:
		t.Fatalf("Didn't receive first response!")
	}

	for i := 1; i < 10; i++ {
		select {
		case other := <-response_channel:
			if !stringSlicesEqual(result, other) {
				t.Errorf("Response %d does not match first response! %s != %s\n", i, result, other)
			}
		default:
			t.Errorf("Not enough responses! Expected 10, but got %d", i+1)
		}
	}

	select {
	case other := <-response_channel:
		t.Errorf("Received too many responses! %s\n", other)
	default:
	}
}

func TestSynchronization(t *testing.T) {

	Node1 := New_Node("Node1", CreateTestNetworkCommunicationChannels())
	Node2 := New_Node("Node2", CreateTestNetworkCommunicationChannels())
	defer Node1.Close()
	defer Node2.Close()

	time.Sleep(time.Millisecond * 100)

	Node1.protocol_dispatcher.Do_Synchronization()

	time.Sleep(time.Millisecond * 150)

	info1 := "Node1 Hi"
	info2 := "Node2 Hello"

	select {
	case <-Node1.shared_state_communication.FromNetwork.Synchronization.ProtocolRequestInformation:
		Node1.shared_state_communication.ToNetwork.Synchronization.RespondToInformationRequest <- info1
	default:
		t.Fatalf("Node 1 did not request Information!")
	}

	select {
	case <-Node2.shared_state_communication.FromNetwork.Synchronization.ProtocolRequestInformation:
		Node2.shared_state_communication.ToNetwork.Synchronization.RespondToInformationRequest <- info2
	default:
		t.Fatalf("Node 2 did not request Information!")
	}

	time.Sleep(time.Millisecond * 10)

	success_message := "Success"

	select {
	case results := <-Node1.shared_state_communication.FromNetwork.Synchronization.ProtocolRequestsInterpretation:
		if results[Node1.name] != info1 {
			t.Errorf("Results in Node 1 did not match sent info! %s != %s\n", results[Node1.name], info1)
		}

		if results[Node2.name] != info2 {
			t.Errorf("Results in Node 2 did not match sent info! %s != %s\n", results[Node2.name], info2)
		}

		Node1.shared_state_communication.ToNetwork.Synchronization.RespondWithInterpretation <- success_message
	default:
		t.Fatalf("Node 1 never asked for interpretation!")
	}

	time.Sleep(time.Millisecond * 10)

	select {
	case result := <-Node1.shared_state_communication.FromNetwork.Synchronization.ResultFromSynchronization:
		if result != success_message {
			t.Errorf("Result returned by Node 1 is not the success message! %s != %s\n", result, success_message)
		}
	default:
		t.Errorf("Did not receive result from Node 1!")
	}

	select {
	case result := <-Node2.shared_state_communication.FromNetwork.Synchronization.ResultFromSynchronization:
		if result != success_message {
			t.Errorf("Result returned by Node 2 is not the success message! %s != %s\n", result, success_message)
		}
	default:
		t.Errorf("Did not receive result from Node 2!")
	}
}

func Test2PC(t *testing.T) {
	Node1 := New_Node("Node1", CreateTestNetworkCommunicationChannels())
	Node2 := New_Node("Node2", CreateTestNetworkCommunicationChannels())
	defer Node1.Close()
	defer Node2.Close()

	time.Sleep(time.Millisecond * 100)

	Node1.protocol_dispatcher.Do_Synchronization()

	time.Sleep(time.Millisecond * 150)
	<-Node1.shared_state_communication.FromNetwork.Synchronization.ProtocolRequestInformation
	Node1.shared_state_communication.ToNetwork.Synchronization.RespondToInformationRequest <- ""
	<-Node2.shared_state_communication.FromNetwork.Synchronization.ProtocolRequestInformation
	Node2.shared_state_communication.ToNetwork.Synchronization.RespondToInformationRequest <- ""
	<-Node1.shared_state_communication.FromNetwork.Synchronization.ProtocolRequestsInterpretation
	Node1.shared_state_communication.ToNetwork.Synchronization.RespondWithInterpretation <- ""

	// Do Test Here.
	command := "some command"

	Node1.protocol_dispatcher.Do_Command(command)

	time.Sleep(time.Millisecond * 100)

	select {
	case result := <-Node1.shared_state_communication.FromNetwork.TwoPhaseCommit.ProtocolCommited:
		if result != command {
			t.Errorf("Received command came back malformed! %s != %s.", command, result)
		}
	default:
		t.Errorf("Did not receive command from Node1!")
	}

	select {
	case result := <-Node2.shared_state_communication.FromNetwork.TwoPhaseCommit.ProtocolCommited:
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
	Node1.protocol_dispatcher.Do_Synchronization()
	time.Sleep(time.Millisecond)

	p2p_message := <-Node2.p2p.Read_Channel
	message := translate_Message(p2p_message)
	Node2.abort_Synchronization(message)

	go Node2.reader()

	time.Sleep(time.Millisecond * 150)

	// Is Node1 a result?
	select {
	case node1_result := <-Node1.shared_state_communication.FromNetwork.Discovery.Updated_Alive_Nodes:
		if !stringSlicesEqual(node1_result, result1) && !stringSlicesEqual(node1_result, result2) {
			t.Fatalf("Node 1 result doesn't make sense! %s\n", node1_result)
		}
	default:
		t.Fatalf("Received no response from Node 1.\n")
	}

	// Is Node2 a result?
	select {
	case node2_result := <-Node2.shared_state_communication.FromNetwork.Discovery.Updated_Alive_Nodes:
		if !stringSlicesEqual(node2_result, result1) && !stringSlicesEqual(node2_result, result2) {
			t.Fatalf("Node 2 result doesn't make sense! %s\n", node2_result)
		}
	default:
		t.Fatalf("Received no response from Node 2.\n")
	}

	t.Logf("Success Discovery redispatch")

	time.Sleep(time.Millisecond * 150)

	info1 := "Node1 Hi"
	info2 := "Node2 Hello"

	select {
	case <-Node1.shared_state_communication.FromNetwork.Synchronization.ProtocolRequestInformation:
		Node1.shared_state_communication.ToNetwork.Synchronization.RespondToInformationRequest <- info1
	default:
		t.Fatalf("Node 1 did not request Information!")
	}

	select {
	case <-Node2.shared_state_communication.FromNetwork.Synchronization.ProtocolRequestInformation:
		Node2.shared_state_communication.ToNetwork.Synchronization.RespondToInformationRequest <- info2
	default:
		t.Fatalf("Node 2 did not request Information!")
	}

	time.Sleep(time.Millisecond * 10)

	success_message := "Success"

	select {
	case results := <-Node1.shared_state_communication.FromNetwork.Synchronization.ProtocolRequestsInterpretation:
		if results[Node1.name] != info1 {
			t.Errorf("Results in Node 1 did not match sent info! %s != %s\n", results[Node1.name], info1)
		}

		if results[Node2.name] != info2 {
			t.Errorf("Results in Node 2 did not match sent info! %s != %s\n", results[Node2.name], info2)
		}

		Node1.shared_state_communication.ToNetwork.Synchronization.RespondWithInterpretation <- success_message
	default:
		t.Fatalf("Node 1 never asked for interpretation!")
	}

	time.Sleep(time.Millisecond * 10)

	select {
	case result := <-Node1.shared_state_communication.FromNetwork.Synchronization.ResultFromSynchronization:
		if result != success_message {
			t.Errorf("Result returned by Node 1 is not the success message! %s != %s\n", result, success_message)
		}
	default:
		t.Errorf("Did not receive result from Node 1!")
	}

	select {
	case result := <-Node2.shared_state_communication.FromNetwork.Synchronization.ResultFromSynchronization:
		if result != success_message {
			t.Errorf("Result returned by Node 2 is not the success message! %s != %s\n", result, success_message)
		}
	default:
		t.Errorf("Did not receive result from Node 2!")
	}
}

func TestDispatchRetry(t *testing.T) {
	name1 := "Node1"
	name2 := "Node2"

	Node1 := New_Node(name1, CreateTestNetworkCommunicationChannels())

	// Need manual control over node 2, not using NewNode that automatically starts a reader and dispatcher.
	Node2 := Node{
		p2p: peer_to_peer.New_P2P_Network(),

		name: name2,

		next_TxID_number: 0,

		alive_nodes_manager: AliveNodeManager{
			alive_nodes: make([]string, 0),
		},
		protocol_dispatcher: *New_Protocol_Dispatcher(),

		comm: make(chan Message, 32),

		close_channel: make(chan bool),

		shared_state_communication: CreateTestNetworkCommunicationChannels(),
	}
	defer Node1.Close()
	defer Node2.Close()

	time.Sleep(time.Millisecond * 100)

	testDiscoveryDispatchRetry(Node1, &Node2, t)
}
