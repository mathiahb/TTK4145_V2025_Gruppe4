package elevator

import (
	"elevator_project/common"
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
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
	toSharedState shared_states.FromElevator,
	elevatorChannels ElevatorChannels,
	doorTimer *time.Timer,
	isStuckTimer *time.Timer,
) common.Elevator {

	port := strconv.Itoa(portElevio)
	elevio.Init("localhost:"+port, common.N_FLOORS)
	fmt.Println("FSM initialized for elevator:", common.GetElevatorID())

	turnOffAllLights()

	if localElevator.Floor == -1 {
		localElevator = FSMOnInitBetweenFloors(localElevator, toSharedState.UpdateState)
	} else {
		if localElevator.Behaviour == common.EB_DoorOpen || localElevator.Behaviour == common.EB_Stuck_DoorOpen {
			// Door open? Close it after 3 seconds.
			localElevator, _ = FSMOpenDoor(localElevator, hallRequests, doorTimer, toSharedState.ClearHallRequest, toSharedState.UpdateState)
		} else if localElevator.Behaviour == common.EB_Moving || localElevator.Behaviour == common.EB_Stuck_Moving {
			// We were moving? Continue moving.
			switch localElevator.Dirn {
			case common.D_Down:
				elevio.SetMotorDirection(elevio.MD_Down)
			case common.D_Up:
				elevio.SetMotorDirection(elevio.MD_Up)
			case common.D_Stop:
				elevio.SetMotorDirection(elevio.MD_Stop)
			}
		} else {
			// We were idle? Figure out if we need to open the door or move
			localElevator, _ = FSMOnFloorArrival(
				localElevator.Floor,
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
		}
	}

	return localElevator
}

// FSMOnInitBetweenFloors is called when the elevator is initialized between floors.
// It sets the motor direction to down and the elevator behaviour to moving.
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

// SetHallLights sets the hall lights based on the hall requests that comes from the shared states module.
func setHallLights(hallRequests common.HallRequestType) {

	for floor := 0; floor < common.N_FLOORS; floor++ {
		elevio.SetButtonLamp(floor, elevio.BT_HallUp, hallRequests[floor][common.B_HallUp])
		elevio.SetButtonLamp(floor, elevio.BT_HallDown, hallRequests[floor][common.B_HallDown])
	}
}

// SetCabLights sets the cab lights based on the cab requests approved by the shared states module.
func setCabLights(cabRequests []bool) {
	for floor := 0; floor < common.N_FLOORS; floor++ {
		elevio.SetButtonLamp(floor, elevio.BT_Cab, cabRequests[floor])
	}
}

// Resets the isStuckTimer
func FSMResetIsStuckTimer(isStuckTimer *time.Timer) {

	if !isStuckTimer.Stop() {
		select {
		case <-isStuckTimer.C:
		default:
		}
	}
	isStuckTimer.Reset(time.Second * common.IsStuckDurationS)

}

// FSMStartMoving is used to start the elevator's motor when it is idle and has requests.
// It resets the isStuckTimer each time the elevator starts moving.
func FSMStartMoving(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
	elevatorStateChannel chan common.Elevator,
	ClearHallRequest chan common.HallRequestType,
	updateStateChannel chan common.Elevator,
	doorTimer *time.Timer,
	isStuckTimer *time.Timer,
) common.Elevator {

	if localElevator.Behaviour == common.EB_Idle && hasRequests(localElevator, hallRequests) {
		localElevator.Dirn = requestsChooseDirection(localElevator, hallRequests)

		FSMResetIsStuckTimer(isStuckTimer)

		switch localElevator.Dirn {
		case common.D_Up:
			if hallRequests[localElevator.Floor][common.B_HallUp] || localElevator.CabRequests[localElevator.Floor] {
				localElevator, _ = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)
			} else {
				localElevator.Behaviour = common.EB_Moving
				elevio.SetMotorDirection(elevio.MD_Up)
			}
		case common.D_Down:
			if hallRequests[localElevator.Floor][common.B_HallDown] || localElevator.CabRequests[localElevator.Floor] {
				localElevator, _ = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)
			} else {
				localElevator.Behaviour = common.EB_Moving
				elevio.SetMotorDirection(elevio.MD_Down)
			}
		case common.D_Stop:
			localElevator, _ = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)

		}

		elevatorStateChannel <- localElevator
	}

	return localElevator
}

// FSMButtonPress communicates new state to the shared state when a button is pressed
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

// FSMOnFloorArrival is called upon when the elevator arrives a new floor.
// It resets the isStuckTimer. If the elevator has "stuck" behaviour when the elevator arrives a new floor,
// we can change it to moving as this is a clear sign of an active elevator.
// The functions also checks if the elevator should stop. If the elevator stops, it opens the door and updates shared states.
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

	if localElevator.Behaviour == common.EB_Stuck_Moving {
		localElevator.Behaviour = common.EB_Moving
	}

	localElevator.Floor = newFloor
	elevio.SetFloorIndicator(localElevator.Floor)

	if localElevator.Behaviour == common.EB_Moving {
		if requestsShouldStop(localElevator, hallRequests) {
			localElevator, hallRequests = FSMOpenDoor(localElevator, hallRequests, doorTimer, ClearHallRequest, updateStateChannel)
		}
	}
	updateStateChannel <- localElevator

	return localElevator, hallRequests
}

// FSMOpenDoor opens the elevator door by turning on the door light.
// It determines the elevator's behavior when the door is open.
func FSMOpenDoor(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
	doorTimer *time.Timer,
	ClearHallRequest chan common.HallRequestType,
	updateStateChannel chan common.Elevator,
) (common.Elevator, common.HallRequestType) {

	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(true)
	localElevator.Behaviour = common.EB_DoorOpen
	doorTimer.Reset(time.Second * common.DoorOpenDurationS)
	localElevator, hallRequests = requestsClearAtCurrentFloor(localElevator, hallRequests, ClearHallRequest, updateStateChannel)
	return localElevator, hallRequests
}

// FSMCloseDoors close the elevator doors by turning of the door light.
// It updates the behaviour and calls FSMStartMoving
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

	if localElevator.Behaviour == common.EB_DoorOpen {
		localElevator.Behaviour = common.EB_Idle
		elevio.SetDoorOpenLamp(false)
		elevatorStateChannel <- localElevator
		localElevator = FSMStartMoving(
			localElevator,
			hallRequests,
			elevatorStateChannel,
			ClearHallRequest,
			updateStateChannel,
			doorTimer,
			isStuckTimer,
		)
	}
	return localElevator
}
