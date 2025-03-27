package elevator

import (
	"elevator_project/elevio"
	"elevator_project/shared_states"
	"elevator_project/common"
	"fmt"
	"strconv"
	"time"
)

// FSM (Finite State Machine) styrer heisens tilstand og oppførsel basert på knappetrykk, etasjeanløp og dørlukkingshendelser.

func InitFSM(
	portElevio int,
	localElevator common.Elevator,
	toSharedState shared_states.FromElevator,
	elevatorChannels ElevatorChannels,
) common.Elevator {

	port := strconv.Itoa(portElevio)
	elevio.Init("localhost:"+port, common.N_FLOORS)
	fmt.Println("FSM initialized for elevator:", common.GetElevatorID())

	turnOffAllLights()

	if localElevator.Floor == -1 {
		localElevator = FSMOnInitBetweenFloors(localElevator, toSharedState.UpdateState)
	} else {
		elevatorChannels.Floor <- localElevator.Floor
	}

	return localElevator
}

func FSMOnInitBetweenFloors(
	localElevator common.Elevator,
	UpdateState chan common.Elevator,
) common.Elevator {

	elevio.SetMotorDirection(elevio.MD_Down)
	localElevator.Dirn = common.D_Down
	localElevator.Behaviour = common.EB_Moving

	UpdateState <- localElevator

	return localElevator
}

func turnOffAllLights() {
	for button := 0; button < common.N_BUTTONS; button++ {
		for floor := 0; floor < common.N_FLOORS; floor++ {
			elevio.SetButtonLamp(floor, elevio.ButtonType(button), false)
		}
	}
	elevio.SetDoorOpenLamp(false)
}

func setHallLights(hallRequests common.HallRequestType) {

	for floor := 0; floor < common.N_FLOORS; floor++ {
		elevio.SetButtonLamp(floor, elevio.BT_HallUp, hallRequests[floor][common.B_HallUp])
		elevio.SetButtonLamp(floor, elevio.BT_HallDown, hallRequests[floor][common.B_HallDown])
	}
}

func setCabLights(cabRequests []bool) {
	for floor := 0; floor < common.N_FLOORS; floor++ {
		elevio.SetButtonLamp(floor, elevio.BT_Cab, cabRequests[floor])
	}
}

func FSMOpenDoor(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
	doorTimer *time.Timer,
	ClearHallRequest chan common.HallRequestType,
	updateStateChannel chan common.Elevator,
) common.Elevator {

	elevio.SetMotorDirection(elevio.MD_Stop)
	localElevator.Behaviour = common.EB_DoorOpen
	elevio.SetDoorOpenLamp(true)
	doorTimer.Reset(time.Second * common.DoorOpenDurationS)
	localElevator, _ = requestsClearAtCurrentFloor(localElevator, hallRequests, ClearHallRequest, updateStateChannel)
	return localElevator
}

func FSMResetIsStuckTimer(isStuckTimer *time.Timer) {

	if !isStuckTimer.Stop() {
		select {
		case <-isStuckTimer.C:
		default:
		}
	}
	isStuckTimer.Reset(time.Second * common.IsStuckDurationS)

}

func FSMStartMoving(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
	elevatorStateChannel chan common.Elevator,
	ClearHallRequest chan common.HallRequestType,
	updateStateChannel chan common.Elevator,
	doorTimer *time.Timer,
	isStuckTimer *time.Timer,
) common.Elevator {

	// Hvis heisen er idle og har forespørsler, velg retning og start motor
	if localElevator.Behaviour == common.EB_Idle && hasRequests(localElevator, hallRequests) {
		localElevator.Dirn = requestsChooseDirection(localElevator, hallRequests)

		FSMResetIsStuckTimer(isStuckTimer) // resetter timeren hver gang man begynner å bevege seg

		// Are there any requests at the current floor in the new direction localElevator.Dirn?
		switch localElevator.Dirn {
		case common.D_Up:
			if hallRequests[localElevator.Floor][common.B_HallUp] || localElevator.CabRequests[localElevator.Floor] {
				// Stop the elevator and open the door
				localElevator = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)
			} else {
				localElevator.Behaviour = common.EB_Moving
				elevio.SetMotorDirection(elevio.MD_Up)
			}
		case common.D_Down:
			if hallRequests[localElevator.Floor][common.B_HallDown] || localElevator.CabRequests[localElevator.Floor] {
				// Stop the elevator and open the door
				localElevator = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)
			} else {
				localElevator.Behaviour = common.EB_Moving
				elevio.SetMotorDirection(elevio.MD_Down)
			}
		case common.D_Stop:
			localElevator = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)

		}

		elevatorStateChannel <- localElevator
	}

	return localElevator
}

func FSMButtonPress(
	btnFloor int,
	btnType elevio.ButtonType,
	localElevator common.Elevator,
	updateStateChannel chan common.Elevator,
	NewHallRequest chan common.HallRequestType,
) common.Elevator {
	fmt.Printf("FSMOnRequestButtonPress(%d, %d)\n", btnFloor, btnType)

	if btnType == elevio.BT_Cab {
		localElevator.CabRequests[btnFloor] = true
		updateStateChannel <- localElevator

	} else if btnType == elevio.BT_HallUp {

		var newHallRequest common.HallRequestType = make(common.HallRequestType, common.N_FLOORS)
		newHallRequest[btnFloor][elevio.BT_HallUp] = true
		NewHallRequest <- newHallRequest

	} else if btnType == elevio.BT_HallDown {

		var newHallRequest common.HallRequestType = make(common.HallRequestType, common.N_FLOORS)
		newHallRequest[btnFloor][elevio.BT_HallDown] = true
		NewHallRequest <- newHallRequest

	}
	return localElevator
}

func FSMOnFloorArrival(
	newFloor int,
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
	ClearHallRequest chan common.HallRequestType,
	updateStateChannel chan common.Elevator,
	doorTimer *time.Timer,
	isStuckTimer *time.Timer) (
	common.Elevator, common.HallRequestType) {

	fmt.Printf("\nFSMOnFloorArrival(%d)\n", newFloor)

	FSMResetIsStuckTimer(isStuckTimer)

	if localElevator.Behaviour == common.EB_Stuck_Moving { // dersom vi ankommer en etasje etter å ha brukt motoren, vet vi at vi ikke er stuck
		localElevator.Behaviour = common.EB_Moving
	}

	// 1. lagre ny etasje i lokal state
	localElevator.Floor = newFloor
	elevio.SetFloorIndicator(localElevator.Floor)

	// 2. Sjekk om heisen skal stoppe
	if localElevator.Behaviour == common.EB_Moving {
		if requestsShouldStop(localElevator, hallRequests) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			localElevator.Behaviour = common.EB_DoorOpen

			// Start dør-timer
			doorTimer.Reset(time.Second * common.DoorOpenDurationS)

			// 3. Fjerner requests på nåværende etasje
			localElevator, hallRequests = requestsClearAtCurrentFloor(localElevator, hallRequests, ClearHallRequest, updateStateChannel)

		}
	}
	updateStateChannel <- localElevator

	return localElevator, hallRequests
}

// **FSMOnDoorTimeout**: Kalles når dør-timeren utløper
func FSMCloseDoors(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
	elevatorStateChannel chan common.Elevator,
	doorTimer *time.Timer,
	isStuckTimer *time.Timer,
	ClearHallRequest chan common.HallRequestType,
	updateStateChannel chan common.Elevator,
) common.Elevator {
	fmt.Println("\nFSMOnDoorTimeout()")

	// Hvis døren er åpen, bestem neste handling
	if localElevator.Behaviour == common.EB_DoorOpen {
		localElevator.Behaviour = common.EB_Idle
		elevio.SetDoorOpenLamp(false)

		elevatorStateChannel <- localElevator

		localElevator = FSMStartMoving(localElevator, hallRequests, elevatorStateChannel, ClearHallRequest, updateStateChannel, doorTimer, isStuckTimer) // sjekker om det er noen forespørsel
	}
	return localElevator
}
