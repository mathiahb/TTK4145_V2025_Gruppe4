package shared_states

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"elevator_project/common"
)

type Command2PC struct {
	Command string
	Name    string
	Data    string
}

func makeHRAInputVariable(sharedState common.HRAType, aliveNodes []string) common.HRAType {
	result := common.HRAType{
		States:       make(map[string]common.Elevator),
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

func updateSharedStateByCommand(command Command2PC, sharedState common.HRAType) common.HRAType {

	switch command.Command {

	case common.ADD:
		newHallRequest := translateFromNetwork[common.HallRequestType](command.Data)

		for i, value := range newHallRequest {
			sharedState.HallRequests[i][0] = sharedState.HallRequests[i][0] || value[0]
			sharedState.HallRequests[i][1] = sharedState.HallRequests[i][1] || value[1]
		}

	case common.REMOVE:
		removeHallRequest := translateFromNetwork[common.HallRequestType](command.Data)

		for i, value := range removeHallRequest {
			sharedState.HallRequests[i][0] = sharedState.HallRequests[i][0] && (!value[0])
			sharedState.HallRequests[i][1] = sharedState.HallRequests[i][1] && (!value[1])
		}

	case common.UPDATE_STATE:

		newState := translateFromNetwork[common.Elevator](command.Data)
		sharedState.States[command.Name] = newState
	}
	return sharedState

}

func reactToSharedStateUpdate(sharedState common.HRAType, aliveNodes []string, localID string, toElevator ToElevator) {

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

func getHallRequestAssignments(HRAInputVariable common.HRAType) map[string][][2]bool {

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

func translateToNetwork(variable any) string {

	translatedVariable, err := json.Marshal(variable)

	if err != nil {
		fmt.Println("json.Marshal error:", err)
		return ""
	}

	return string(translatedVariable)
}

func translateFromNetwork[T any](variable string) T {
	var translatedVariable T
	err := json.Unmarshal([]byte(variable), &translatedVariable)
	if err != nil {
		fmt.Println("json.Unmarshal error:", err)
	}
	return translatedVariable
}
