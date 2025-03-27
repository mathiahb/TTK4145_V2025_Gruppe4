package shared_states

import (
	"elevator_project/constants"
	"encoding/json"
	"fmt"
	"os/exec"
)

// Command2PC represents a two-phase commit command structure with
// a command type, a name identifier, and associated data.
type Command2PC struct {
	Command string
	Name    string
	Data    string
}

// makeHRAInputVariable creates a new HRAType input variable by filtering out
// stuck elevators from the shared state and including only alive nodes' states
// and hall requests.
func makeHRAInputVariable(sharedState constants.HRAType, aliveNodes []string) constants.HRAType {
	result := constants.HRAType{
		States:       make(map[string]constants.Elevator),
		HallRequests: sharedState.HallRequests,
	}

	for _, nodeID := range aliveNodes {
		if state, ok := sharedState.States[nodeID]; ok {
			if !state.Behaviour.IsStuck() {
				result.States[nodeID] = sharedState.States[nodeID]
			}
		}
	}

	return result
}

// updateSharedStateByCommand updates the shared state based on a given two-phase commit command.
// It handles adding and removing hall requests, and updating elevator states.
func updateSharedStateByCommand(command Command2PC, sharedState constants.HRAType) constants.HRAType {

	switch command.Command {

	case constants.ADD:
		newHallRequest := translateFromNetwork[constants.HallRequestType](command.Data)

		for i, value := range newHallRequest {
			sharedState.HallRequests[i][0] = sharedState.HallRequests[i][0] || value[0]
			sharedState.HallRequests[i][1] = sharedState.HallRequests[i][1] || value[1]
		}

	case constants.REMOVE:
		removeHallRequest := translateFromNetwork[constants.HallRequestType](command.Data)

		for i, value := range removeHallRequest {
			sharedState.HallRequests[i][0] = sharedState.HallRequests[i][0] && (!value[0])
			sharedState.HallRequests[i][1] = sharedState.HallRequests[i][1] && (!value[1])
		}

	case constants.UPDATE_STATE:

		newState := translateFromNetwork[constants.Elevator](command.Data)
		sharedState.States[command.Name] = newState
	}
	return sharedState

}

// reactToSharedStateUpdate
// 1) processes updates to the shared state,
// 2) computes hall request assignments
// and 3) sends the results to the elevator system.
// It also updates hall request lights and approved cab requests for the local elevator.
func reactToSharedStateUpdate(sharedState constants.HRAType, aliveNodes []string, localID string, toElevator ToElevator) {

	HRAInputVariable := makeHRAInputVariable(sharedState, aliveNodes)
	HRAResults := getHallRequestAssignments(HRAInputVariable)
	approvedCabRequests := sharedState.States[localID].CabRequests // må sende cabRequest separat fra resten av states for å sørge for at heisen ikke "tar" en bestilling uten bekreftelse fra nettverket

	if HRAResults != nil && HRAResults[localID] != nil {
		toElevator.ApprovedHRA <- HRAResults[localID]
	}
	toElevator.UpdateHallRequestLights <- sharedState.HallRequests
	if approvedCabRequests != nil {
		toElevator.ApprovedCabRequests <- approvedCabRequests
	}
}

// getHallRequestAssignments calls an external executable to compute hall
// request assignments based on the provided HRAType input variable. It
// returns the assignments as a map of elevator IDs to hall request arrays.
func getHallRequestAssignments(HRAInputVariable constants.HRAType) map[string][][2]bool {

	// Convert to JSON
	jsonBytes, err := json.Marshal(HRAInputVariable)
	if err != nil {
		fmt.Println("json.Marshal error:", err)
		return nil
	}

	// Call `hall_request_assigner`
	ret, err := exec.Command("./hall_request_assigner/hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error:", err)
		fmt.Println(string(ret))
		return nil
	}

	// Debugging: Print raw response
	fmt.Println("Raw response from hall_request_assigner:", string(ret))

	// Parse JSON response
	output := make(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error:", err)
		return nil
	}

	return output
}

// ===================== TRANSLATION TO/FROM NETWORK ===================== //
// translateToNetwork serializes a given variable into a JSON string for
// network transmission.
func translateToNetwork(variable any) string {

	translatedVariable, err := json.Marshal(variable)

	if err != nil {
		fmt.Println("json.Marshal error:", err)
		return ""
	}

	return string(translatedVariable)
}

// translateFromNetwork deserializes a JSON string into a specified type,
// converting it back to its original structure.
func translateFromNetwork[T any](variable string) T {
	var translatedVariable T
	err := json.Unmarshal([]byte(variable), &translatedVariable)
	if err != nil {
		fmt.Println("json.Unmarshal error:", err)
	}
	return translatedVariable
}
