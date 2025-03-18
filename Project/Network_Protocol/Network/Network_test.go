package network

import (
	"strconv"
	"testing"
	"time"
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

	go Node1.coordinate_Discovery()

	time.Sleep(time.Second)

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
	go Node1.coordinate_Discovery()

	time.Sleep(time.Second)

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

	go Node1.coordinate_Discovery()

	time.Sleep(time.Millisecond * 150)

	go Node1.coordinate_Synchronization()

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
