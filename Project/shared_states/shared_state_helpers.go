package shared_states

import(
	. "elevator_project/constants"
)


type Command2PC struct {
	Command string
	Name    string
	Data    string
}



func makeHRAInputVariable(sharedState HRAType, aliveNodes []string) HRAType {
	HRAInputVariable := HRAType{
		States:       make(map[string]Elevator),
		HallRequests: sharedState.HallRequests,
	}

	for _, nodeID := range aliveNodes {
		if(!sharedState.States[nodeID].Behaviour.IsStuck()){
			HRAInputVariable.States[nodeID] = sharedState.States[nodeID]
		}	
	}

	return HRAInputVariable
}

func updateSharedStateByCommand(command Command2PC, sharedState HRAType) HRAType {

	switch command.Command {

	case ADD:
		newHallRequest := translateFromNetwork[HallRequestType](command.Data)

		for i, value := range newHallRequest {
			sharedState.HallRequests[i][0] = sharedState.HallRequests[i][0] || value[0]
			sharedState.HallRequests[i][1] = sharedState.HallRequests[i][1] || value[1]
		}

	case REMOVE:
		removeHallRequest := translateFromNetwork[HallRequestType](command.Data)

		for i, value := range removeHallRequest {
			sharedState.HallRequests[i][0] = sharedState.HallRequests[i][0] && (!value[0])
			sharedState.HallRequests[i][1] = sharedState.HallRequests[i][1] && (!value[1])
		}

	case UPDATE_STATE:

		newState := translateFromNetwork[Elevator](command.Data)
		sharedState.States[command.Name] = newState
	}
	return sharedState

}

func reactToSharedStateUpdate(sharedState HRAType, activeNodes []string, localID string, toElevator ToElevator) {

	HRAInputVariable := makeHRAInputVariable(sharedState, activeNodes)
	HRAResults := getHallRequestAssignments(HRAInputVariable)
	approvedCabRequests := sharedState.States[localID].CabRequests // må sende cabRequest separat fra resten av states for å sørge for at heisen ikke "tar" en bestilling uten bekreftelse fra nettverket

	if HRAResults != nil {
		toElevator.ApprovedHRAChannel <- HRAResults[localID]
	}

	toElevator.UpdateHallRequestLights <- sharedState.HallRequests

	if approvedCabRequests != nil {
		toElevator.ApprovedCabRequestsChannel <- approvedCabRequests
	}
}


