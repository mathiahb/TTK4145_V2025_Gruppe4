package shared_states

import (
	"elevator_project/common"
	"encoding/json"
	"fmt"
)

// Function that keeps the priority order:
// No requests are lost
// An elevator has control over it's own states.
func resolveDifferences(
	state1 common.HRAType,
	state2 common.HRAType,
	ownerOfState2 string,
) common.HRAType {
	elevatorState, ok := state2.States[ownerOfState2]

	// State2 is authorative over it's own state
	if ok {
		state1.States[ownerOfState2] = elevatorState
	}

	// Add any states missing in state1 from state2
	for name, elevator := range state2.States {
		_, ok := state1.States[name]
		if !ok {
			state1.States[name] = elevator
		}
	}

	// Sanity check the hallrequests
	if len(state1.HallRequests) != 0 && len(state2.HallRequests) != 0 && len(state1.HallRequests) != len(state2.HallRequests) {
		fmt.Printf("ResolveDifferences did not properly parse HallRequests, different nonzero lengths!")
		return state1
	}

	// If both states have HallRequests, do bitwise OR on every request (Make sure no orders are lost)
	if len(state1.HallRequests) != 0 && len(state2.HallRequests) != 0 {
		for i, val := range state2.HallRequests {
			state1.HallRequests[i][0] = state1.HallRequests[i][0] || val[0]
			state1.HallRequests[i][1] = state1.HallRequests[i][1] || val[1]
		}
	}

	// If HallRequests in state 1 is empty, take state 2s HallRequests.
	if len(state1.HallRequests) == 0 {
		state1.HallRequests = state2.HallRequests
	}

	return state1
}

func ResolveSharedStateConflicts(states map[string]string) string {
	result := common.HRAType{
		States:       make(map[string]common.Elevator),
		HallRequests: make(common.HallRequestType, common.N_FLOORS),
	}

	for name, state := range states {
		newState := common.HRAType{}
		err := json.Unmarshal([]byte(state), &newState)

		if err == nil {
			result = resolveDifferences(result, newState, name)
		}
	}

	jsonResult, err := json.Marshal(result)

	if err != nil {
		fmt.Printf("[Error] Shared State Conflict Resolver did not resolve properly!\n")
		for _, state := range states {
			return state
		}
	}

	return string(jsonResult)
}
