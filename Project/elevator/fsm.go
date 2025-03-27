package elevator

import (
	"elevator_project/constants"
	"elevator_project/elevio"
	"elevator_project/shared_states"
	"fmt"
	"strconv"
	"time"
)

// FSM (Finite State Machine) controls the elevator's state and behavior
// based on button presses, floor arrivals, and door closing events.

// InitFSM initializes the FSM for the elevator.
func InitFSM(
	portElevio int,
	localElevator constants.Elevator,
	toSharedState shared_states.FromElevator,
	elevatorChannels ElevatorChannels,
) constants.Elevator {

	port := strconv.Itoa(portElevio)
	elevio.Init("localhost:"+port, constants.N_FLOORS)
	fmt.Println("FSM initialized for elevator:", constants.GetElevatorID())

	turnOffAllLights()

	if localElevator.Floor == -1 {
		localElevator = FSMOnInitBetweenFloors(localElevator, toSharedState.UpdateState)
	} else {
		elevatorChannels.Floor <- localElevator.Floor
	}

	return localElevator
}

// FSMOnInitBetweenFloors is called when the elevator is between floors.
// It sets the motor direction to down and the elevator behavior to moving.
func FSMOnInitBetweenFloors(
	localElevator constants.Elevator,
	UpdateState chan constants.Elevator,
) constants.Elevator {

	elevio.SetMotorDirection(elevio.MD_Down)
	localElevator.Dirn = constants.D_Down
	localElevator.Behaviour = constants.EB_Moving

	UpdateState <- localElevator

	return localElevator
}

func turnOffAllLights() {
	for button := 0; button < constants.N_BUTTONS; button++ {
		for floor := 0; floor < constants.N_FLOORS; floor++ {
			elevio.SetButtonLamp(floor, elevio.ButtonType(button), false)
		}
	}
	elevio.SetDoorOpenLamp(false)
}

// setHallLights sets the hall lights based on the hall requests.
func setHallLights(hallRequests constants.HallRequestType) {

	for floor := 0; floor < constants.N_FLOORS; floor++ {
		elevio.SetButtonLamp(floor, elevio.BT_HallUp, hallRequests[floor][constants.B_HallUp])
		elevio.SetButtonLamp(floor, elevio.BT_HallDown, hallRequests[floor][constants.B_HallDown])
	}
}

func setCabLights(cabRequests []bool) {
	for floor := 0; floor < constants.N_FLOORS; floor++ {
		elevio.SetButtonLamp(floor, elevio.BT_Cab, cabRequests[floor])
	}
}

// FSMOpenDoor determines the elevator's behavior when the door is open.
func FSMOpenDoor(
	localElevator constants.Elevator,
	hallRequests constants.HallRequestType,
	doorTimer *time.Timer,
	ClearHallRequest chan constants.HallRequestType,
	updateStateChannel chan constants.Elevator,
) constants.Elevator {

	elevio.SetMotorDirection(elevio.MD_Stop)
	localElevator.Behaviour = constants.EB_DoorOpen
	elevio.SetDoorOpenLamp(true)
	doorTimer.Reset(time.Second * constants.DoorOpenDurationS)
	localElevator, _ = requestsClearAtCurrentFloor(localElevator, hallRequests, ClearHallRequest, updateStateChannel)
	return localElevator
}

// Resets the isStuckTimer
func FSMResetIsStuckTimer(isStuckTimer *time.Timer) {

	if !isStuckTimer.Stop() {
		select {
		case <-isStuckTimer.C:
		default:
		}
	}
	isStuckTimer.Reset(time.Second * constants.IsStuckDurationS)

}

// FSMStartMoving is used to start the elevator's motor when it is idle and has requests.
func FSMStartMoving(
	localElevator constants.Elevator,
	hallRequests constants.HallRequestType,
	elevatorStateChannel chan constants.Elevator,
	ClearHallRequest chan constants.HallRequestType,
	updateStateChannel chan constants.Elevator,
	doorTimer *time.Timer,
	isStuckTimer *time.Timer,
) constants.Elevator {

	// Hvis heisen er idle og har forespørsler, velg retning og start motor
	if localElevator.Behaviour == constants.EB_Idle && hasRequests(localElevator, hallRequests) {
		localElevator.Dirn = requestsChooseDirection(localElevator, hallRequests)

		FSMResetIsStuckTimer(isStuckTimer) // resetter timeren hver gang man begynner å bevege seg

		// Are there any requests at the current floor in the new direction localElevator.Dirn?
		switch localElevator.Dirn {
		case constants.D_Up:
			if hallRequests[localElevator.Floor][constants.B_HallUp] || localElevator.CabRequests[localElevator.Floor] {
				// Stop the elevator and open the door
				localElevator = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)
			} else {
				localElevator.Behaviour = constants.EB_Moving
				elevio.SetMotorDirection(elevio.MD_Up)
			}
		case constants.D_Down:
			if hallRequests[localElevator.Floor][constants.B_HallDown] || localElevator.CabRequests[localElevator.Floor] {
				// Stop the elevator and open the door
				localElevator = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)
			} else {
				localElevator.Behaviour = constants.EB_Moving
				elevio.SetMotorDirection(elevio.MD_Down)
			}
		case constants.D_Stop:
			localElevator = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)

		}

		elevatorStateChannel <- localElevator
	}

	return localElevator
}

// Communicates new state to the shared state when a button is pressed
func FSMButtonPress(
	btnFloor int,
	btnType elevio.ButtonType,
	localElevator constants.Elevator,
	updateStateChannel chan constants.Elevator,
	NewHallRequest chan constants.HallRequestType,
) constants.Elevator {
	fmt.Printf("FSMOnRequestButtonPress(%d, %d)\n", btnFloor, btnType)

	if btnType == elevio.BT_Cab {
		localElevator.CabRequests[btnFloor] = true
		updateStateChannel <- localElevator

	} else if btnType == elevio.BT_HallUp {

		var newHallRequest constants.HallRequestType = make(constants.HallRequestType, constants.N_FLOORS)
		newHallRequest[btnFloor][elevio.BT_HallUp] = true
		NewHallRequest <- newHallRequest

	} else if btnType == elevio.BT_HallDown {

		var newHallRequest constants.HallRequestType = make(constants.HallRequestType, constants.N_FLOORS)
		newHallRequest[btnFloor][elevio.BT_HallDown] = true
		NewHallRequest <- newHallRequest

	}
	return localElevator
}

func FSMOnFloorArrival(
	newFloor int,
	localElevator constants.Elevator,
	hallRequests constants.HallRequestType,
	ClearHallRequest chan constants.HallRequestType,
	updateStateChannel chan constants.Elevator,
	doorTimer *time.Timer,
	isStuckTimer *time.Timer) (
	constants.Elevator, constants.HallRequestType) {

	fmt.Printf("\nFSMOnFloorArrival(%d)\n", newFloor)

	FSMResetIsStuckTimer(isStuckTimer)

	if localElevator.Behaviour == constants.EB_Stuck_Moving { // dersom vi ankommer en etasje etter å ha brukt motoren, vet vi at vi ikke er stuck
		localElevator.Behaviour = constants.EB_Moving
	}

	// 1. lagre ny etasje i lokal state
	localElevator.Floor = newFloor
	elevio.SetFloorIndicator(localElevator.Floor)

	// 2. Sjekk om heisen skal stoppe
	if localElevator.Behaviour == constants.EB_Moving {
		if requestsShouldStop(localElevator, hallRequests) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			localElevator.Behaviour = constants.EB_DoorOpen

			// Start dør-timer
			doorTimer.Reset(time.Second * constants.DoorOpenDurationS)

			// 3. Fjerner requests på nåværende etasje
			localElevator, hallRequests = requestsClearAtCurrentFloor(localElevator, hallRequests, ClearHallRequest, updateStateChannel)

		}
	}
	updateStateChannel <- localElevator

	return localElevator, hallRequests
}

// **FSMOnDoorTimeout**: Kalles når dør-timeren utløper
func FSMCloseDoors(
	localElevator constants.Elevator,
	hallRequests constants.HallRequestType,
	elevatorStateChannel chan constants.Elevator,
	doorTimer *time.Timer,
	isStuckTimer *time.Timer,
	ClearHallRequest chan constants.HallRequestType,
	updateStateChannel chan constants.Elevator,
) constants.Elevator {
	fmt.Println("\nFSMOnDoorTimeout()")

	// Hvis døren er åpen, bestem neste handling
	if localElevator.Behaviour == constants.EB_DoorOpen {
		localElevator.Behaviour = constants.EB_Idle
		elevio.SetDoorOpenLamp(false)

		elevatorStateChannel <- localElevator

		localElevator = FSMStartMoving(localElevator, hallRequests, elevatorStateChannel, ClearHallRequest, updateStateChannel, doorTimer, isStuckTimer) // sjekker om det er noen forespørsel
	}
	return localElevator
}
