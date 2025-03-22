package shared_states

import (
	"elevator_project/constants"
	"encoding/json"
	"testing"
)

func TestConflictResolver(t *testing.T) {
	name1 := "Node1"
	name2 := "Node2"
	name3 := "Node3"

	hallRequests1 := [][2]bool{{true, false}, {true, false}, {true, false}, {true, false}}
	hallRequests2 := [][2]bool{}
	hallRequests3 := [][2]bool{{false, true}, {false, true}, {false, true}, {false, true}}
	expected_bool_result := [][2]bool{{true, true}, {true, true}, {true, true}, {true, true}}

	state1 := constants.HRAType{
		States:       make(map[string]constants.Elevator),
		HallRequests: hallRequests1,
	}

	state2 := constants.HRAType{
		States:       make(map[string]constants.Elevator),
		HallRequests: hallRequests2,
	}

	state3 := constants.HRAType{
		States:       make(map[string]constants.Elevator),
		HallRequests: hallRequests3,
	}

	elevator1 := constants.Elevator{
		Behaviour:   constants.EB_Idle,
		Floor:       2,
		Dirn:        constants.D_Down,
		CabRequests: []bool{false, false, false, false},
	}

	elevator2 := constants.Elevator{
		Behaviour:   constants.EB_DoorOpen,
		Floor:       3,
		Dirn:        constants.D_Up,
		CabRequests: []bool{true, true, true, true},
	}

	fake_elevator2 := elevator1

	elevator3 := constants.Elevator{
		Behaviour:   constants.EB_Moving,
		Floor:       1,
		Dirn:        constants.D_Stop,
		CabRequests: []bool{false, true, false, true},
	}

	fake_elevator3 := elevator2

	// state1.States[name1] = delete
	state1.States[name2] = fake_elevator2
	state1.States[name3] = fake_elevator3

	state2.States[name1] = elevator1
	state2.States[name2] = elevator2
	state2.States[name3] = fake_elevator3

	state3.States[name2] = fake_elevator2
	state3.States[name3] = elevator3

	expected_result := constants.HRAType{
		States:       make(map[string]constants.Elevator),
		HallRequests: expected_bool_result,
	}

	expected_result.States[name1] = elevator1
	expected_result.States[name2] = elevator2
	expected_result.States[name3] = elevator3

	expected_string_result, err := json.Marshal(expected_result)
	if err != nil {
		t.Fatalf("Failed to interpret expected result!")
	}

	var testData map[string]string = make(map[string]string)

	byte1, err := json.Marshal(state1)
	if err != nil {
		t.Fatalf("Failed to interpret state1!.")
	}

	byte2, err := json.Marshal(state2)
	if err != nil {
		t.Fatalf("Failed to interpret state2!.")
	}

	byte3, err := json.Marshal(state3)
	if err != nil {
		t.Fatalf("Failed to interpret state3!.")
	}

	testData[name1] = string(byte1)
	testData[name2] = string(byte2)
	testData[name3] = string(byte3)

	resultstring := ResolveSharedStateConflicts(testData)

	if err != nil {
		t.Fatalf("Failed to interpret result!")
	}

	if string(expected_string_result) != string(resultstring) {
		t.Fatalf("Result was not the expected one! %s != %s", expected_string_result, resultstring)
	}

}
