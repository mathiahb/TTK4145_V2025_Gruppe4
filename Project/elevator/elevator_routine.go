package elevator

import (
	"elevator_project/common"
	"elevator_project/shared_states"
	"fmt"
	"time"
)

// ElevatorRoutine starts and runs the main loop for a single elevator,
// handling all local events and state transitions.
//
// Responsibilities:
// - Initializes the elevator’s finite state machine (FSM)
// - Handles button presses, floor arrivals, obstruction detection, and stuck conditions
// - Responds to approved requests from the shared system state
// - Sends state updates back to the shared state
//
// Input Channels:
// - Local button events, floor signals, and obstruction sensor
// - Approved hall and cab requests from the shared state
//
// Output:
// - Updates elevator state and sends it to the shared states module
// - Controls door behavior and movement using FSM logic
// - Sets lights based on current state and active orders
//
// Timers:
// - doorTimer controls how long the door stays open
// - isStuckTimer detects when the elevator is stuck in motion
func ElevatorRoutine(
	portElevio int,
	initElevator common.Elevator,
	elevatorChannels ElevatorChannels,
	fromSharedState shared_states.ToElevator,
	toSharedState shared_states.FromElevator,
) {
	var hallRequests = hallRequestsUninitialized()

	var isObstructed = false

	var doorTimer = time.NewTimer(time.Second * common.DoorOpenDurationS)
	var isStuckTimer = time.NewTimer(time.Second * common.IsStuckDurationS)

	var localElevator = InitFSM(portElevio, initElevator, hallRequests, toSharedState, elevatorChannels, doorTimer, isStuckTimer)

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

		case isObstructed = <-elevatorChannels.Obstruction:
			fmt.Printf("Obstruction switch: %v\n", isObstructed)

			if localElevator.Behaviour == common.EB_DoorOpen || localElevator.Behaviour == common.EB_Stuck_DoorOpen {

				if !doorTimer.Stop() { // Stop timer first to avoid channel blocking.
					select {
					case <-doorTimer.C:
					default:
					}
				}

				if !isObstructed {
					fmt.Printf("Door is not obstructed, closing door\n")
					doorTimer.Reset(time.Second * common.DoorOpenDurationS)
				} else if localElevator.Behaviour == common.EB_DoorOpen { // Check to not flood the 2PC channel.
					localElevator.Behaviour = common.EB_Stuck_DoorOpen // An obstruction will update the elevators behaviour to being stuck.
					toSharedState.UpdateState <- localElevator         // The shared state module will then call upon the HRA without the obstructed (stuck) elevator
				}
			}

		case <-doorTimer.C: // Is triggered when the door closes. Either after the openDoorDuration is up after handling an order, or an openDoorDuration after the obstruction is cleared.
			fmt.Printf("Door timer expired, obstruction: %v\n", isObstructed)
			if !isObstructed {
				fmt.Printf("Door is not obstructed, closing door\n")

				if localElevator.Behaviour == common.EB_Stuck_DoorOpen {
					localElevator.Behaviour = common.EB_DoorOpen
				}

				localElevator = FSMCloseDoors(localElevator, hallRequests, toSharedState.UpdateState, doorTimer, isStuckTimer, toSharedState.ClearHallRequest, toSharedState.UpdateState)
			}

		case <-isStuckTimer.C: // Is triggered when the elevator has behaviour moving, but has not been detected crossing a floor before the isStuckTimer is up.
			// This case is for instance activated when the elevator is connected to the network, but the motor is disconnected.
			if localElevator.Behaviour == common.EB_Moving {
				fmt.Printf("\n\n[%s] ELEVATOR IS STUCK\n\n", common.GetElevatorID())
				localElevator.Behaviour = common.EB_Stuck_Moving
				toSharedState.UpdateState <- localElevator
			}

		case hallRequests = <-fromSharedState.ApprovedHRA: // All hall requests orders are sent from the shared state module
			localElevator = FSMStartMoving(localElevator, hallRequests, toSharedState.UpdateState, toSharedState.ClearHallRequest, toSharedState.UpdateState, doorTimer, isStuckTimer)

		case sharedHallRequests := <-fromSharedState.UpdateHallRequestLights:
			setHallLights(sharedHallRequests)

		case cabRequests := <-fromSharedState.ApprovedCabRequests:
			setCabLights(cabRequests)
		}
	}
}
