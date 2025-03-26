package elevator

import (
	"elevator_project/constants"
	"elevator_project/shared_states"
	"fmt"
	"time"
)

func ElevatorRoutine(portElevio int, initElevator constants.Elevator, elevatorChannels ElevatorChannels, fromSharedState shared_states.ToElevator, toSharedState shared_states.FromElevator) {

	localElevator := InitFSM(portElevio, initElevator, toSharedState, elevatorChannels)
	var hallRequests = HallRequestsUninitialized()
	var isObstructed = false
	var doorTimer = time.NewTimer(time.Second * constants.DoorOpenDurationS)
	var isStuckTimer = time.NewTimer(time.Second * constants.IsStuckDurationS)

	for {
		select {

		case buttonEvent := <-elevatorChannels.Button:
			localElevator = FSMButtonPress(buttonEvent.Floor, buttonEvent.Button, localElevator, toSharedState.UpdateState, toSharedState.NewHallRequest)

		case newFloor := <-elevatorChannels.Floor:
			localElevator, hallRequests = FSMOnFloorArrival(
				newFloor,
				localElevator,
				hallRequests,
				toSharedState.ClearHallRequest,
				toSharedState.UpdateState,
				doorTimer,
				isStuckTimer,
			)
			localElevator = FSMStartMoving(
				localElevator,
				hallRequests,
				toSharedState.UpdateState,
				toSharedState.ClearHallRequest,
				toSharedState.UpdateState,
				doorTimer,
				isStuckTimer,
			)

		case isObstructed = <-elevatorChannels.Obstruction:
			fmt.Printf("Obstruction switch: %v\n", isObstructed)

			if localElevator.Behaviour == constants.EB_DoorOpen || localElevator.Behaviour == constants.EB_Stuck_DoorOpen {

				// Stop timer first to avoid channel blocking.
				if !doorTimer.Stop() {
					select {
					case <-doorTimer.C:
					default:
					}
				}

				if !isObstructed {
					fmt.Printf("Door is not obstructed, closing door\n")
					doorTimer.Reset(3 * time.Second)
				} else if localElevator.Behaviour == constants.EB_DoorOpen { // Check to not flood the 2PC channel.
					localElevator.Behaviour = constants.EB_Stuck_DoorOpen
					toSharedState.UpdateState <- localElevator
				}
			}

			// -- Timeren går ut etter 3 sek --
			// Skal alltid trigges når døren lukkes
		case <-doorTimer.C:
			fmt.Printf("Door timer expired, obstruction: %v\n", isObstructed)
			if !isObstructed {
				fmt.Printf("Door is not obstructed, closing door\n")

				if localElevator.Behaviour == constants.EB_Stuck_DoorOpen {
					localElevator.Behaviour = constants.EB_DoorOpen
				}

				localElevator = FSMCloseDoors(localElevator, hallRequests, toSharedState.UpdateState, doorTimer, isStuckTimer, toSharedState.ClearHallRequest, toSharedState.UpdateState)
			}

		case <-isStuckTimer.C:
			if localElevator.Behaviour == constants.EB_Moving {
				fmt.Printf("\n\n[%s] ELEVATOR IS STUCK\n\n", constants.GetElevatorID())
				localElevator.Behaviour = constants.EB_Stuck_Moving
				toSharedState.UpdateState <- localElevator
			}

		case hallRequests = <-fromSharedState.ApprovedHRA: // fordi alle ordre kommer fra shared states
			localElevator = FSMStartMoving(localElevator, hallRequests, toSharedState.UpdateState, toSharedState.ClearHallRequest, toSharedState.UpdateState, doorTimer, isStuckTimer,)

		case sharedHallRequests := <-fromSharedState.UpdateHallRequestLights:
			setHallLights(sharedHallRequests)

		case cabRequests := <-fromSharedState.ApprovedCabRequests:
			setCabLights(cabRequests)
		}
	}
}
