package elevator

import (
	"elevator_project/constants"
	"elevator_project/elevio"
	"elevator_project/shared_states"
	"fmt"
	"strconv"
	"time"
)

// FSM (Finite State Machine) styrer heisens tilstand og oppførsel basert på knappetrykk, etasjeanløp og dørlukkingshendelser.
func InitFSM(portElevio int, localElevator constants.Elevator, toSharedState shared_states.FromElevator, elevatorChannels ElevatorChannels) Elevator {

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

func FSMOpenDoor(
	localElevator constants.Elevator,
	hallRequests constants.HallRequestType,
	doorTimer *time.Timer,
	ClearHallRequest chan constants.HallRequestType,
	updateStateChannel chan constants.Elevator,
) constants.Elevator {

	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(true)
	localElevator.Behaviour = constants.EB_DoorOpen
	doorTimer.Reset(3 * time.Second)
	localElevator, _ = requestsClearAtCurrentFloor(localElevator, hallRequests, ClearHallRequest, updateStateChannel)
	return localElevator
}

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

		if !isStuckTimer.Stop() {
			select {
			case <-isStuckTimer.C:
			default:
			}
		}
		isStuckTimer.Reset(5 * time.Second)

		// Are there any requests at the current floor in the new direction localElevator.Dirn?
		switch localElevator.Dirn {
		case constants.D_Up:
			if hallRequests[localElevator.Floor][constants.B_HallUp] || localElevator.CabRequests[localElevator.Floor] {
				// Stop the elevator and open the door
				localElevator = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)
			} else {
				localElevator.Behaviour = constants.EB_Moving
				elevio.SetMotorDirection(convertDirnToMotor(localElevator.Dirn))
			}
		case constants.D_Down:
			if hallRequests[localElevator.Floor][constants.B_HallDown] || localElevator.CabRequests[localElevator.Floor] {
				// Stop the elevator and open the door
				localElevator = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)
			} else {
				localElevator.Behaviour = constants.EB_Moving
				elevio.SetMotorDirection(convertDirnToMotor(localElevator.Dirn))
			}
		case constants.D_Stop:
			localElevator = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)

		}

		elevatorStateChannel <- localElevator
	}

	return localElevator
}

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
	isStuckTimer *time.Timer)(
constants.Elevator, constants.HallRequestType) {

	fmt.Printf("\nFSMOnFloorArrival(%d)\n", newFloor)

	if !isStuckTimer.Stop() {
		select {
		case <-isStuckTimer.C:
		default:
		}
	}
	isStuckTimer.Reset(time.Second * constants.IsStuckDurationS)

	if localElevator.Behaviour == constants.EB_Stuck_Moving {
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

func convertDirnToMotor(d constants.Dirn) elevio.MotorDirection { // føler ikke at denne funksjonen hører hjemme her, tror kanskje den kan sløyfes
	switch d {
	case "up":
		return elevio.MD_Up
	case "down":
		return elevio.MD_Down
	default:
		return elevio.MD_Stop
	}
}
