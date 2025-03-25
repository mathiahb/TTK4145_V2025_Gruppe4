package shared_states

import (
	. "elevator_project/constants"
	"fmt"
	//"time"
	//"golang.org/x/text/cases"
)

// ===================== SHARED STATE ===================== //
// Bridge between network and the elevator. The shared states communicates also with the HRA.



func SharedStateThread(initResult chan Elevator, toElevator ToElevator, fromNetwork FromNetwork, toNetwork ToNetwork, fromElevator FromElevator) {
	var sharedState HRAType = HRAType{
		States:       make(map[string]Elevator),
		HallRequests: make(HallRequestType, N_FLOORS),
	}
	// lage et nytt state i shared state?
	var localID string = GetElevatorID()
	var aliveNodes []string = make([]string, 0)
	var initializing bool = true

	fmt.Printf("SharedStateThread Initialized: %s\n", localID)

	for {
		select {
		// 2PC
		case newHallRequest := <-fromElevator.NewHallRequestChannel: 
			command := Command2PC{
				Command: ADD,
				Name:    localID,
				Data:    translateToNetwork(newHallRequest),
			}
			go func() { toNetwork.Inform2PC <- translateToNetwork(command) }()

		case clearHallRequest := <-fromElevator.ClearHallRequestChannel: 
			command := Command2PC{
				Command: REMOVE,
				Name:    localID,
				Data:    translateToNetwork(clearHallRequest),
			}

			go func() { toNetwork.Inform2PC <- translateToNetwork(command) }()

		case newState := <-fromElevator.UpdateState:

			command := Command2PC{
				Command: UPDATE_STATE,
				Name:    localID,
				Data:    translateToNetwork(newState),
			}
			
			go func() { toNetwork.Inform2PC <- translateToNetwork(command) }()


		case commandString := <-fromNetwork.ApprovedBy2PC:
			command := translateFromNetwork[Command2PC](commandString)
			sharedState = updateSharedStateByCommand(command, sharedState)
			go reactToSharedStateUpdate(sharedState, aliveNodes, localID, toElevator)

		// discovery
		case aliveNodes = <-fromNetwork.New_alive_nodes:
			fmt.Printf("SharedStateThread: New alive nodes: %v\n", aliveNodes)
			go reactToSharedStateUpdate(sharedState, aliveNodes, localID, toElevator)

		// synkronisering
		case <-fromNetwork.ProtocolRequestInformation:
			fmt.Printf("SharedStateThread: Responding to information request\n")
			go func() { toNetwork.RespondToInformationRequest <- translateToNetwork(sharedState) }()

		case states := <-fromNetwork.ProtocolRequestsInterpretation:
			go func() { toNetwork.RespondWithInterpretation <- ResolveSharedStateConflicts(states) }()

		case newSharedState := <-fromNetwork.ResultFromSynchronization:
			sharedState = translateFromNetwork[HRAType](newSharedState)

			if initializing {
				initializing = false
				res, ok := sharedState.States[localID]
				if !ok {
					res = Elevator{
						Behaviour:   EB_Idle,
						Dirn:        D_Stop,
						Floor:       -1,
						CabRequests: make([]bool, N_FLOORS),
					}
				}
				go func() { initResult <- res }()
			}

			go reactToSharedStateUpdate(sharedState, aliveNodes, localID, toElevator)
		}
	}
}
