package shared_states

import (
	. "elevator_project/constants"
)

// ===================== SHARED STATE ===================== //
// Bridge between network and the elevator. The shared states communicates also with the HRA.

func makeHRAInputVariable(sharedState HRAType, aliveNodes []string) HRAType {
	result := HRAType{
		States:       make(map[string]Elevator),
		HallRequests: sharedState.HallRequests,
	}

	for _, nodeID := range aliveNodes {
		result.States[nodeID] = sharedState.States[nodeID]
	}

	return result
}

type Command2PC struct {
	Command string
	Name    string
	Data    string
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

func reactToSharedStateUpdate(sharedState HRAType, aliveNodes []string, localID string, toElevator ToElevator) {

	HRAInputVariable := makeHRAInputVariable(sharedState, aliveNodes)
	HRAResults := getHallRequestAssignments(HRAInputVariable)
	approvedCabRequests := sharedState.States[localID].CabRequests // må sende cabRequest separat fra resten av states for å sørge for at heisen ikke "tar" en bestilling uten bekreftelse fra nettverket

	toElevator.ApprovedHRAChannel <- HRAResults[localID]
	toElevator.UpdateHallRequestLights <- sharedState.HallRequests
	toElevator.ApprovedCabRequestsChannel <- approvedCabRequests
}

func SharedStateThread(toElevator ToElevator, fromNetwork FromNetwork, toNetwork ToNetwork, fromElevator FromElevator) {

	var sharedState HRAType
	var localID string = GetElevatorID()
	var aliveNodes []string = make([]string, 0)

	for {
		select {
		// 2PC
		case newHallRequest := <-fromElevator.NewHallRequestChannel: // får inn en enkelt hallRequest {false, false} {false, false} {true, false} {false, false}
			command := Command2PC{
				Command: ADD,
				Name:    localID,
				Data:    translateToNetwork(newHallRequest),
			}
			toNetwork.Inform2PC <- translateToNetwork(command)

		case clearHallRequest := <-fromElevator.ClearHallRequestChannel: // får inn en enkelt hallRequest {false, false} {false, false} {true, false} {false, false}
			command := Command2PC{
				Command: REMOVE,
				Name:    localID,
				Data:    translateToNetwork(clearHallRequest),
			}

			toNetwork.Inform2PC <- translateToNetwork(command)

		case newState := <-fromElevator.UpdateState:
			command := Command2PC{
				Command: UPDATE_STATE,
				Name:    localID,
				Data:    translateToNetwork(newState),
			}

			toNetwork.Inform2PC <- translateToNetwork(command)

		case commandString := <-fromNetwork.ApprovedBy2PC:
			command := translateFromNetwork[Command2PC](commandString)
			sharedState = updateSharedStateByCommand(command, sharedState)
			reactToSharedStateUpdate(sharedState, aliveNodes, localID, toElevator)

		// discovery
		case aliveNodes = <-fromNetwork.New_alive_nodes:
			reactToSharedStateUpdate(sharedState, aliveNodes, localID, toElevator)

		// synkronisering
		case <-fromNetwork.ProtocolRequestInformation:
			toNetwork.RespondToInformationRequest <- translateToNetwork(sharedState)

		case states := <-fromNetwork.ProtocolRequestsInterpretation:
			toNetwork.RespondWithInterpretation <- ResolveSharedStateConflicts(states)

		case newSharedState := <-fromNetwork.ResultFromSynchronization:
			sharedState = translateFromNetwork[HRAType](newSharedState)
			reactToSharedStateUpdate(sharedState, aliveNodes, localID, toElevator)
		}
	}
}
