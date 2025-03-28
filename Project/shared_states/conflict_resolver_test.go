package shared_states

import (
	"encoding/json"
	"fmt"
	"testing"
	"elevator_project/common"
)

func TestConflictResolver(t *testing.T) {
	name1 := "Node1"
	name2 := "Node2"
	name3 := "Node3"

	hallRequests1 := [][2]bool{{true, false}, {true, false}, {true, false}, {true, false}}
	hallRequests2 := [][2]bool{}
	hallRequests3 := [][2]bool{{false, true}, {false, true}, {false, true}, {false, true}}
	expectedBoolResult := [][2]bool{{true, true}, {true, true}, {true, true}, {true, true}}

	state1 := common.HRAType{
		States:       make(map[string]common.Elevator),
		HallRequests: hallRequests1,
	}

	state2 := common.HRAType{
		States:       make(map[string]common.Elevator),
		HallRequests: hallRequests2,
	}

	state3 := common.HRAType{
		States:       make(map[string]common.Elevator),
		HallRequests: hallRequests3,
	}

	elevator1 := common.Elevator{
		Behaviour:   common.EB_Idle,
		Floor:       2,
		Dirn:        common.D_Down,
		CabRequests: []bool{false, false, false, false},
	}

	elevator2 := common.Elevator{
		Behaviour:   common.EB_DoorOpen,
		Floor:       3,
		Dirn:        common.D_Up,
		CabRequests: []bool{true, true, true, true},
	}

	fakeElevator2 := elevator1

	elevator3 := common.Elevator{
		Behaviour:   common.EB_Moving,
		Floor:       1,
		Dirn:        common.D_Stop,
		CabRequests: []bool{false, true, false, true},
	}

	fakeElevator3 := elevator2

	// state1.States[name1] = delete
	state1.States[name2] = fakeElevator2
	state1.States[name3] = fakeElevator3

	state2.States[name1] = elevator1
	state2.States[name2] = elevator2
	state2.States[name3] = fakeElevator3

	state3.States[name2] = fakeElevator2
	state3.States[name3] = elevator3

	expectedResult := common.HRAType{
		States:       make(map[string]common.Elevator),
		HallRequests: expectedBoolResult,
	}

	expectedResult.States[name1] = elevator1
	expectedResult.States[name2] = elevator2
	expectedResult.States[name3] = elevator3

	expectedStringResult, err := json.Marshal(expectedResult)
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

	resultString := ResolveSharedStateConflicts(testData)

	if err != nil {
		t.Fatalf("Failed to interpret result!")
	}

	if string(expectedStringResult) != string(resultString) {
		t.Fatalf("Result was not the expected one! %s != %s", expectedStringResult, resultString)
	}

}

func TestTranslation(t *testing.T) {

	hallRequests1 := [][2]bool{{true, false}, {true, false}, {true, false}, {true, false}}

	elevator1 := common.Elevator{
		Behaviour:   common.EB_Idle,
		Floor:       2,
		Dirn:        common.D_Down,
		CabRequests: []bool{false, false, false, false},
	}

	state1 := common.HRAType{
		States:       make(map[string]common.Elevator),
		HallRequests: hallRequests1,
	}

	state1.States["Heis nummer 1"] = elevator1

	translatedState1 := translateToNetwork(state1)
	deTranslatedState1 := translateFromNetwork[common.HRAType](translatedState1)
	//translatedHallRequest := translateToNetwork(hallRequests1)
	//deTranslatedHallRequest := translateFromNetwork[HallRequestType](translatedHallRequest) // HallRequestType [][2]bool
	//HRAInputVariable.HallRequests = deTranslatedHallRequest

	result1 := fmt.Sprintf("%+v", state1)
	result2 := fmt.Sprintf("%+v", deTranslatedState1)

	t.Logf("result1 = %s", result1)
	t.Logf("result2 = %s", result2)

	if result1 != result2 {
		t.Errorf("Results did not match, %s != %s", result1, result2)
	}

}
