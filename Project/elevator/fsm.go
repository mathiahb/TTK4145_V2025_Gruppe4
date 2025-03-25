package elevator

import (
	. "elevator_project/constants"
	"elevator_project/elevio"
	"elevator_project/shared_states"
	"fmt"
	"os"
	"time"
)

// FSM (Finite State Machine) styrer heisens tilstand og oppførsel basert på knappetrykk, etasjeanløp og dørlukkingshendelser.
// Den håndterer tilstander som Idle, DoorOpen og Moving, og bestemmer heisens retning og handlinger.

func InitFSM(localElevator Elevator, elevatorChannels ElevatorChannels, toSharedState shared_states.FromElevator, isStuckTimer *time.Timer) Elevator {

	elevio.Init("localhost:15657", N_FLOORS)
	fmt.Println("FSM initialized for elevator:", getElevatorID())
	
	turnOffAllLights() // starter med alle lys avslått

	if localElevator.Floor == -1 {
		localElevator = FSMOnInitBetweenFloors(localElevator, toSharedState.UpdateState, isStuckTimer)
	}else{
		elevatorChannels.Floor <- localElevator.Floor
	}
	return localElevator
}

// **FSMOnInitBetweenFloors**: Kalles hvis heisen starter mellom etasjer
func FSMOnInitBetweenFloors(localElevator Elevator, UpdateState chan Elevator, isStuckTimer *time.Timer) Elevator {

	elevio.SetMotorDirection(elevio.MD_Down)
	localElevator.Dirn = D_Down
	localElevator.Behaviour = EB_Moving

	isStuckTimer.Reset(time.Second * 5)

	UpdateState <- localElevator

	return localElevator
}

// getElevatorID returns a unique identifier for this elevator instance.
func getElevatorID() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return "unknown_elevator"
	}
	return hostname
}

func turnOffAllLights() {
	//starter med alle lys avslått
	for button := 0; button < N_BUTTONS; button++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			elevio.SetButtonLamp(floor, elevio.ButtonType(button), false)
		}
	}
	elevio.SetDoorOpenLamp(false)
}

func ElevatorUninitialized() Elevator { // Initialize a new elevator with default values
	turnOffAllLights()
	return Elevator{
		Floor:       -1,
		Dirn:        D_Stop,
		Behaviour:   EB_Idle,
		CabRequests: make([]bool, N_FLOORS),
	}
}

func convertDirnToMotor(d Dirn) elevio.MotorDirection { // føler ikke at denne funksjonen hører hjemme her
	switch d {
	case "up":
		return elevio.MD_Up
	case "down":
		return elevio.MD_Down
	default:
		return elevio.MD_Stop
	}
}

func setHallLights(hallRequests HallRequestType) {

	for floor := 0; floor < N_FLOORS; floor++ {
		// Hall requests fra SharedState
		elevio.SetButtonLamp(floor, elevio.BT_HallUp, hallRequests[floor][B_HallUp])
		elevio.SetButtonLamp(floor, elevio.BT_HallDown, hallRequests[floor][B_HallDown])
	}
}

func setCabLights(cabRequests []bool) {
	for floor := 0; floor < N_FLOORS; floor++ {
		// Hall requests fra SharedState
		elevio.SetButtonLamp(floor, elevio.BT_Cab, cabRequests[floor])
	}
}

func FSMOpenDoor(
	localElevator Elevator,
	hallRequests HallRequestType,
	threeSecTimer *time.Timer,
	clearHallRequestChannel chan HallRequestType,
	updateStateChannel chan Elevator,
) Elevator {
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(true)
	localElevator.Behaviour = EB_DoorOpen
	threeSecTimer.Reset(3 * time.Second)
	localElevator, _ = requestsClearAtCurrentFloor(localElevator, hallRequests, clearHallRequestChannel, updateStateChannel)
	return localElevator
}

// **FSMStartMoving**: Kalles for å sjekke om heisen skal starte bevegelse
func FSMStartMoving(
	localElevator Elevator,
	hallRequests HallRequestType,
	elevatorStateChannel chan Elevator,
	threeSecTimer *time.Timer,
	clearHallRequestChannel chan HallRequestType,
	updateStateChannel chan Elevator,
) Elevator {

	// Hvis heisen er idle og har forespørsler, velg retning og start motor
	if localElevator.Behaviour == EB_Idle && hasRequests(localElevator, hallRequests) {
		localElevator.Dirn = requestsChooseDirection(localElevator, hallRequests)

		if !isStuckTimer.Stop() {
			select {
			case <-isStuckTimer.C: // Tøm kanalen hvis den ikke er tom
			default: // Fortsett hvis kanalen allerede er tom
			}
		}
		isStuckTimer.Reset(time.Second * 5)

		// Are there any requests at the current floor in the new direction localElevator.Dirn?
		switch localElevator.Dirn {
		case D_Up:
			if hallRequests[localElevator.Floor][B_HallUp] || localElevator.CabRequests[localElevator.Floor] {
				// Stop the elevator and open the door
				localElevator = FSMOpenDoor(localElevator, hallRequests, threeSecTimer, clearHallRequestChannel, updateStateChannel)
			} else {
				localElevator.Behaviour = EB_Moving
				elevio.SetMotorDirection(convertDirnToMotor(localElevator.Dirn))
			}
		case D_Down:
			if hallRequests[localElevator.Floor][B_HallDown] || localElevator.CabRequests[localElevator.Floor] {
				// Stop the elevator and open the door
				localElevator = FSMOpenDoor(localElevator, hallRequests, threeSecTimer, clearHallRequestChannel, updateStateChannel)
			} else {
				localElevator.Behaviour = EB_Moving
				elevio.SetMotorDirection(convertDirnToMotor(localElevator.Dirn))
			}
		case D_Stop:
			localElevator = FSMOpenDoor(localElevator, hallRequests, threeSecTimer, clearHallRequestChannel, updateStateChannel)

		}

		elevatorStateChannel <- localElevator
	}

	return localElevator
}

// **FSMOnRequestButtonPress**: Kalles når en knapp trykkes
func FSMButtonPress(btnFloor int, btnType elevio.ButtonType, localElevator Elevator, updateStateChannel chan Elevator, newHallRequestChannel chan HallRequestType) Elevator {
	fmt.Printf("FSMOnRequestButtonPress(%d, %d)\n", btnFloor, btnType)

	if btnType == elevio.BT_Cab {
		localElevator.CabRequests[btnFloor] = true
		updateStateChannel <- localElevator

	} else if btnType == elevio.BT_HallUp {

		var newHallRequest HallRequestType = make(HallRequestType, N_FLOORS)
		newHallRequest[btnFloor][elevio.BT_HallUp] = true
		newHallRequestChannel <- newHallRequest

	} else if btnType == elevio.BT_HallDown {

		var newHallRequest HallRequestType = make(HallRequestType, N_FLOORS)
		newHallRequest[btnFloor][elevio.BT_HallDown] = true
		newHallRequestChannel <- newHallRequest

	}
	return localElevator
}

// **FSMOnFloorArrival**: Kalles når heisen ankommer en ny etasje
func FSMOnFloorArrival(newFloor int,
	localElevator Elevator,
	hallRequests HallRequestType,
	clearHallRequestChannel chan HallRequestType,
	updateStateChannel chan Elevator,
	threeSecTimer *time.Timer,
	isStuckTimer *time.Timer) (Elevator, HallRequestType) {

	fmt.Printf("\nFSMOnFloorArrival(%d)\n", newFloor)

	if !isStuckTimer.Stop() {
		select {
		case <-isStuckTimer.C: // Tøm kanalen hvis den ikke er tom
		default: // Fortsett hvis kanalen allerede er tom
		}
	}
	isStuckTimer.Reset(time.Second * 5)

	if(localElevator.Behaviour == EB_Stuck_Moving){
		localElevator.Behaviour = EB_Moving
	}

	// 1. lagre ny etasje i lokal state
	localElevator.Floor = newFloor
	elevio.SetFloorIndicator(localElevator.Floor)
	

	// 2. Sjekk om heisen skal stoppe
	if localElevator.Behaviour == EB_Moving { 
		if requestsShouldStop(localElevator, hallRequests) { 
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			localElevator.Behaviour = EB_DoorOpen

			// Start dør-timer
			threeSecTimer.Reset(3 * time.Second)

			// 3. Fjerner requests på nåværende etasje
			localElevator, hallRequests = requestsClearAtCurrentFloor(localElevator, hallRequests, clearHallRequestChannel, updateStateChannel)

		}
	}

	updateStateChannel <- localElevator

	return localElevator, hallRequests
}

// **FSMOnDoorTimeout**: Kalles når dør-timeren utløper
func FSMCloseDoors(
	localElevator Elevator,
	hallRequests HallRequestType,
	elevatorStateChannel chan Elevator,
	threeSecTimer *time.Timer,
	clearHallRequestChannel chan HallRequestType,
	updateStateChannel chan Elevator,

) Elevator {
	fmt.Println("\nFSMOnDoorTimeout()")

	// Hvis døren er åpen, bestem neste handling
	if localElevator.Behaviour == EB_DoorOpen {
		localElevator.Behaviour = EB_Idle

		elevio.SetDoorOpenLamp(false)

		elevatorStateChannel <- localElevator

		localElevator = FSMStartMoving(localElevator, hallRequests, elevatorStateChannel, isStuckTimer, threeSecTimer, clearHallRequestChannel, updateStateChannel) // sjekker om det er noen forespørsel
	}
	return localElevator
}
