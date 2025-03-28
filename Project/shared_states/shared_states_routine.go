package shared_states

import (
	"elevator_project/common"
	"fmt"
)

// SharedStatesRoutine synchronizes the elevator system's shared state between
// local and network components. It handles hall requests, state updates,
// network discovery, and conflict resolution during synchronization.
//
// Runs as a goroutine, continuously processing events to maintain consistency.
func SharedStatesRoutine(
	initResult chan common.Elevator,
	toElevator ToElevator,
	fromElevator FromElevator,
	toNetwork ToNetwork,
	fromNetwork FromNetwork,
) {
	var sharedState common.HRAType = common.HRAType{
		States:       make(map[string]common.Elevator),
		HallRequests: make(common.HallRequestType, common.N_FLOORS),
	}
	var localID string = common.GetElevatorID()
	var aliveNodes []string = make([]string, 0)
	var initializing bool = true

	fmt.Printf("SharedStateThread Initialized: %s\n", localID)

	for {
		select {
		// Two phase commit
		case newHallRequest := <-fromElevator.NewHallRequest:
			fmt.Printf("[%s] Got new HR Request: %+v\n\n", localID, newHallRequest)
			command := Command2PC{
				Command: common.ADD,
				Name:    localID,
				Data:    translateToNetwork(newHallRequest),
			}
			go func() { toNetwork.Inform2PC <- translateToNetwork(command) }()

		case clearHallRequest := <-fromElevator.ClearHallRequest:
			fmt.Printf("[%s] Got clear HR Request: %+v\n\n", localID, clearHallRequest)
			command := Command2PC{
				Command: common.REMOVE,
				Name:    localID,
				Data:    translateToNetwork(clearHallRequest),
			}
			go func() { toNetwork.Inform2PC <- translateToNetwork(command) }()

		case newState := <-fromElevator.UpdateState:
			fmt.Printf("[%s] Got new State Request: %+v\n\n", localID, newState)
			command := Command2PC{
				Command: common.UPDATE_STATE,
				Name:    localID,
				Data:    translateToNetwork(newState),
			}
			go func() { toNetwork.Inform2PC <- translateToNetwork(command) }()

		case commandString := <-fromNetwork.ApprovedBy2PC:
			command := translateFromNetwork[Command2PC](commandString)
			sharedState = updateSharedStateByCommand(command, sharedState)
			go reactToSharedStateUpdate(sharedState, aliveNodes, localID, toElevator)

		// Discovery
		case aliveNodes = <-fromNetwork.NewAliveNodes:
			fmt.Printf("SharedStateThread: New alive nodes: %v\n", aliveNodes)
			go reactToSharedStateUpdate(sharedState, aliveNodes, localID, toElevator)

		// Sycnhronization
		case <-fromNetwork.ProtocolRequestInformation:
			fmt.Printf("SharedStateThread: Responding to information request\n")
			go func() { toNetwork.RespondToInformationRequest <- translateToNetwork(sharedState) }()

		case states := <-fromNetwork.ProtocolRequestsInterpretation:
			go func() { toNetwork.RespondWithInterpretation <- ResolveSharedStateConflicts(states) }()

		case newSharedState := <-fromNetwork.ResultFromSynchronization:
			sharedState = translateFromNetwork[common.HRAType](newSharedState)

			if initializing {
				initializing = false
				result, ok := sharedState.States[localID]
				if !ok {
					result = common.Elevator{
						Behaviour:   common.EB_Idle,
						Dirn:        common.D_Stop,
						Floor:       -1,
						CabRequests: make([]bool, common.N_FLOORS),
					}
				}
				go func() { initResult <- result }()
			}
			go reactToSharedStateUpdate(sharedState, aliveNodes, localID, toElevator)
		}
	}
}
