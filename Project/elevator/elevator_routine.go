package elevator

import (
	"elevator_project/shared_states"
	"elevator_project/common"
	"fmt"
	"time"
)

func ElevatorRoutine(
	portElevio int,
	initElevator common.Elevator,
	elevatorChannels ElevatorChannels,
	fromSharedState shared_states.ToElevator,
	toSharedState shared_states.FromElevator,
) {

	var localElevator = InitFSM(portElevio, initElevator, toSharedState, elevatorChannels)
	var hallRequests = HallRequestsUninitialized()
	var isObstructed = false
	var doorTimer = time.NewTimer(time.Second * common.DoorOpenDurationS)
	var isStuckTimer = time.NewTimer(time.Second * common.IsStuckDurationS)

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

			if localElevator.Behaviour == common.EB_DoorOpen || localElevator.Behaviour == common.EB_Stuck_DoorOpen {

				// Stop timer first to avoid channel blocking.
				if !doorTimer.Stop() {
					select {
					case <-doorTimer.C:
					default:
					}
				}

				if !isObstructed {
					fmt.Printf("Door is not obstructed, closing door\n")
					doorTimer.Reset(time.Second * common.DoorOpenDurationS)
				} else if localElevator.Behaviour == common.EB_DoorOpen { // Check to not flood the 2PC channel.
					localElevator.Behaviour = common.EB_Stuck_DoorOpen // definerer det å ha obstruction mens døren er åpen som at heisen er stuck, ettersom den da ikke kan ta ordre for en uviss fremtid
					toSharedState.UpdateState <- localElevator
				}
			}

		case <-doorTimer.C: // Skal alltid trigges når døren lukkes, både etter dørtimer eller opendoorduration etter at obstuksjonen er ferdig
			fmt.Printf("Door timer expired, obstruction: %v\n", isObstructed)
			if !isObstructed {
				fmt.Printf("Door is not obstructed, closing door\n")

				if localElevator.Behaviour == common.EB_Stuck_DoorOpen {
					localElevator.Behaviour = common.EB_DoorOpen
				}

				localElevator = FSMCloseDoors(localElevator, hallRequests, toSharedState.UpdateState, doorTimer, isStuckTimer, toSharedState.ClearHallRequest, toSharedState.UpdateState)
			}

		case <-isStuckTimer.C: // dersom motoren ikke fungerer selv om heisen er koblet til f.eks.
			if localElevator.Behaviour == common.EB_Moving {
				fmt.Printf("\n\n[%s] ELEVATOR IS STUCK\n\n", common.GetElevatorID())
				localElevator.Behaviour = common.EB_Stuck_Moving
				toSharedState.UpdateState <- localElevator
			}

		case hallRequests = <-fromSharedState.ApprovedHRA: // fordi alle ordre kommer fra shared states
			localElevator = FSMStartMoving(localElevator, hallRequests, toSharedState.UpdateState, toSharedState.ClearHallRequest, toSharedState.UpdateState, doorTimer, isStuckTimer)

		case sharedHallRequests := <-fromSharedState.UpdateHallRequestLights:
			setHallLights(sharedHallRequests)

		case cabRequests := <-fromSharedState.ApprovedCabRequests:
			setCabLights(cabRequests)
		}
	}
}
