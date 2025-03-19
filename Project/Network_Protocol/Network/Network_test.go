package network

import (
	"strconv"
	"testing"
	"time"

	peer_to_peer "github.com/mathiahb/TTK4145_V2025_Gruppe4/Network_Protocol/Network/Peer_to_Peer"
)

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

	response_channel1 := make(chan []string, 1)
	response_channel2 := make(chan []string, 1)

	Node1 := New_Node(name1, response_channel1, SynchronizationChannels{})
	Node2 := New_Node(name2, response_channel2, SynchronizationChannels{}) // Node 2

	defer Node1.Close()
	defer Node2.Close()

	time.Sleep(time.Millisecond * 100)

	Node1.protocol_dispatcher.Do_Discovery()

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

	Node1 := New_Node("Node0", response_channel, SynchronizationChannels{})
	defer Node1.Close()

	name := "Node"
	for id := 1; id < 10; id++ {
		Node := New_Node(name+strconv.Itoa(id), response_channel, SynchronizationChannels{})
		defer Node.Close()
	}

	time.Sleep(time.Millisecond * 100)
	Node1.protocol_dispatcher.Do_Discovery()

	time.Sleep(time.Millisecond * 150)

	result := make([]string, 0)

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
	SyncChannels1 := New_SynchronizationChannels()
	SyncChannels2 := New_SynchronizationChannels()

	Node1 := New_Node("Node1", make(chan []string, 1), SyncChannels1)
	Node2 := New_Node("Node2", make(chan []string, 1), SyncChannels2)
	defer Node1.Close()
	defer Node2.Close()

	time.Sleep(time.Millisecond * 100)

	Node1.protocol_dispatcher.Do_Discovery()

	time.Sleep(time.Millisecond * 150)

	Node1.protocol_dispatcher.Do_Synchronization()

	info1 := "Node1 Hi"
	info2 := "Node2 Hello"

	time.Sleep(time.Millisecond * 10)

	select {
	case <-SyncChannels1.ProtocolRequestInformation:
		SyncChannels1.RespondToInformationRequest <- info1
	default:
		t.Fatalf("Node 1 did not request Information!")
	}

	select {
	case <-SyncChannels2.ProtocolRequestInformation:
		SyncChannels2.RespondToInformationRequest <- info2
	default:
		t.Fatalf("Node 2 did not request Information!")
	}

	time.Sleep(time.Millisecond * 10)

	success_message := "Success"

	select {
	case results := <-SyncChannels1.ProtocolRequestsInterpretation:
		if results[Node1.name] != info1 {
			t.Errorf("Results in Node 1 did not match sent info! %s != %s\n", results[Node1.name], info1)
		}

		if results[Node2.name] != info2 {
			t.Errorf("Results in Node 2 did not match sent info! %s != %s\n", results[Node2.name], info2)
		}

		SyncChannels1.RespondWithInterpretation <- success_message
	default:
		t.Fatalf("Node 1 never asked for interpretation!")
	}

	time.Sleep(time.Millisecond * 10)

	select {
	case result := <-SyncChannels1.ResultFromSynchronization:
		if result != success_message {
			t.Errorf("Result returned by Node 1 is not the success message! %s != %s\n", result, success_message)
		}
	default:
		t.Errorf("Did not receive result from Node 1!")
	}

	select {
	case result := <-SyncChannels2.ResultFromSynchronization:
		if result != success_message {
			t.Errorf("Result returned by Node 2 is not the success message! %s != %s\n", result, success_message)
		}
	default:
		t.Errorf("Did not receive result from Node 2!")
	}
}

func Test2PC(t *testing.T) {
	Node1 := New_Node("Node1", make(chan []string), New_SynchronizationChannels())
	Node2 := New_Node("Node2", make(chan []string), New_SynchronizationChannels())
	defer Node1.Close()
	defer Node2.Close()

	time.Sleep(time.Millisecond * 100)

	Node1.protocol_dispatcher.Do_Discovery()

	time.Sleep(time.Millisecond * 150)

	// Do Test Here.
}

func testDiscoveryDispatchRetry(Node1 *Node, Node2 *Node, t *testing.T) {
	name1 := Node1.name
	name2 := Node2.name

	result1 := []string{name1, name2}
	result2 := []string{name2, name1}

	// Testing retry of Discovery
	Node1.protocol_dispatcher.Do_Discovery()

	p2p_message := <-Node2.p2p.Read_Channel
	message := Message_From_String(p2p_message.Message)
	Node2.abort_Discovery(message.id)

	go Node2.reader()

	time.Sleep(time.Millisecond * 150)

	// Is Node1 a result?
	select {
	case node1_result := <-Node1.new_alive_nodes:
		if !stringSlicesEqual(node1_result, result1) && !stringSlicesEqual(node1_result, result2) {
			t.Fatalf("Node 1 result doesn't make sense! %s\n", node1_result)
		}
	default:
		t.Fatalf("Received no response from Node 1.\n")
	}

	// Is Node2 a result?
	select {
	case node2_result := <-Node2.new_alive_nodes:
		if !stringSlicesEqual(node2_result, result1) && !stringSlicesEqual(node2_result, result2) {
			t.Fatalf("Node 2 result doesn't make sense! %s\n", node2_result)
		}
	default:
		t.Fatalf("Received no response from Node 2.\n")
	}

	Node2.Close()

	*Node2 = Node{
		p2p: peer_to_peer.New_P2P_Network(),

		name: name2,

		coordinating: false,

		next_id: 0,

		alive_nodes_manager: AliveNodeManager{
			alive_nodes: make([]string, 0),
		},
		protocol_dispatcher: *New_Protocol_Dispatcher(),

		comm: make(chan Message, 32),

		close_channel: make(chan bool),

		new_alive_nodes:          make(chan []string, 10),
		synchronization_channels: New_SynchronizationChannels(),
	}

	time.Sleep(time.Millisecond * 100)
}

func testSynchronizationRetry(Node1 *Node, Node2 *Node, t *testing.T) {
	Node1.protocol_dispatcher.Do_Synchronization()

	info1 := "Node1 Hi"
	info2 := "Node2 Hello"

	time.Sleep(time.Millisecond * 10)

	SyncChannels1 := Node1.synchronization_channels
	SyncChannels2 := Node2.synchronization_channels

	select {
	case <-SyncChannels1.ProtocolRequestInformation:
		SyncChannels1.RespondToInformationRequest <- info1
	default:
		t.Fatalf("Node 1 did not request Information!")
	}

	p2p_message := <-Node2.p2p.Read_Channel
	message := Message_From_String(p2p_message.Message)
	Node2.abort_Synchronization(message.id)

	go Node2.reader()

	// Should now redispatch a discovery and reattempt.
	time.Sleep(time.Millisecond * 100)

	select {
	case <-SyncChannels1.ProtocolRequestInformation:
		SyncChannels1.RespondToInformationRequest <- info1
	default:
		t.Fatalf("Node 1 did not request Information!")
	}

	select {
	case <-SyncChannels2.ProtocolRequestInformation:
		SyncChannels2.RespondToInformationRequest <- info2
	default:
		t.Fatalf("Node 2 did not request Information!")
	}

	time.Sleep(time.Millisecond * 10)

	success_message := "Success"

	select {
	case results := <-SyncChannels1.ProtocolRequestsInterpretation:
		if results[Node1.name] != info1 {
			t.Errorf("Results in Node 1 did not match sent info! %s != %s\n", results[Node1.name], info1)
		}

		if results[Node2.name] != info2 {
			t.Errorf("Results in Node 2 did not match sent info! %s != %s\n", results[Node2.name], info2)
		}

		SyncChannels1.RespondWithInterpretation <- success_message
	default:
		t.Fatalf("Node 1 never asked for interpretation!")
	}

	time.Sleep(time.Millisecond * 10)

	select {
	case result := <-SyncChannels1.ResultFromSynchronization:
		if result != success_message {
			t.Errorf("Result returned by Node 1 is not the success message! %s != %s\n", result, success_message)
		}
	default:
		t.Errorf("Did not receive result from Node 1!")
	}

	select {
	case result := <-SyncChannels2.ResultFromSynchronization:
		if result != success_message {
			t.Errorf("Result returned by Node 2 is not the success message! %s != %s\n", result, success_message)
		}
	default:
		t.Errorf("Did not receive result from Node 2!")
	}

	Node2.Close()

	*Node2 = Node{
		p2p: peer_to_peer.New_P2P_Network(),

		name: Node2.name,

		coordinating: false,

		next_id: 0,

		alive_nodes_manager: AliveNodeManager{
			alive_nodes: make([]string, 0),
		},
		protocol_dispatcher: *New_Protocol_Dispatcher(),

		comm: make(chan Message, 32),

		close_channel: make(chan bool),

		new_alive_nodes:          make(chan []string),
		synchronization_channels: New_SynchronizationChannels(),
	}
}

func TestDispatchRetry(t *testing.T) {
	name1 := "Node1"
	name2 := "Node2"

	alive_channel_1 := make(chan []string, 1)
	alive_channel_2 := make(chan []string, 1)

	synch_channel_1 := New_SynchronizationChannels()
	synch_channel_2 := New_SynchronizationChannels()

	Node1 := New_Node(name1, alive_channel_1, synch_channel_1)

	// Need manual control over node 2, not using NewNode that automatically starts a reader and dispatcher.
	Node2 := Node{
		p2p: peer_to_peer.New_P2P_Network(),

		name: name2,

		coordinating: false,

		next_id: 0,

		alive_nodes_manager: AliveNodeManager{
			alive_nodes: make([]string, 0),
		},
		protocol_dispatcher: *New_Protocol_Dispatcher(),

		comm: make(chan Message, 32),

		close_channel: make(chan bool),

		new_alive_nodes:          alive_channel_2,
		synchronization_channels: synch_channel_2,
	}
	defer Node1.Close()
	defer Node2.Close()

	time.Sleep(time.Millisecond * 100)

	testDiscoveryDispatchRetry(Node1, &Node2, t)

	// Synchronization

	testSynchronizationRetry(Node1, &Node2, t)
}
