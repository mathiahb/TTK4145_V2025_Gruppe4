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

	Node1 := New_Node(name1, response_channel1)
	Node2 := New_Node(name2, response_channel2) // Node 2

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

	Node1 := New_Node("Node0", response_channel)

	name := "Node"
	for id := 1; id < 10; id++ {
		New_Node(name+strconv.Itoa(id), response_channel)
	}

	time.Sleep(time.Millisecond * 100)
	Node1.coordinate_Discovery()

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
