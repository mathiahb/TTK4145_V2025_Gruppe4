package elevator

import (
	elevio "Driver-Elevio"
	"fmt"
)

// FSM (Finite State Machine) styrer heisens tilstand og oppførsel basert på knappetrykk, etasjeanløp og dørlukkingshendelser.
// Den håndterer tilstander som Idle, DoorOpen og Moving, og bestemmer heisens retning og handlinger.

var ElevatorID string
var outputDevice elevio.ElevOutputDevice //hva gjør denne?
type HallRequestsType [][2]bool

func elevator(newHallRequestChannel chan NewHallRequest, elevatorStateChannel chan Elevator, hallRequestChannel chan HRAOutput) {
	
	//Alt dette er initialisering
	var localElevator = ElevatorUninitialized()
	var hallRequests = HallRequestsUninitialized()
	
	InitFSM(elevatorStateChannel, localElevator)
	config := elevator.LoadConfig("elevator/elevator.con")
	inputPollRateMs := config.InputPollRateMs
	inputDevice := elevio.GetInputDevice()
	
	// If the elevator starts between floors, handle initialization
	if inputDevice.FloorSensor() == -1 {
		elevator.FSMOnInitBetweenFloors()
	}

	// Declared outside the loop since static doesn't exist in Go for local variables
	var prev [elevator.N_FLOORS][elevator.N_BUTTONS]int // Tracks previous button states
	prevFloor := -1
	
	// Main loop
	for {
		// Sjekker fortløpende om knapper er trykket og oppdaterer systemet deretter.
		for f := 0; f < elevator.N_FLOORS; f++ {
			for b := 0; b < elevator.N_BUTTONS; b++ {
				v := inputDevice.RequestButton(elevio.ButtonType(b), f)
				if v && prev[f][b] == 0 {
					elevator.FSMOnRequestButtonPress(f, elevio.ButtonType(b))
				}
				prev[f][b] = boolToInt(v)
			}
		}
	
			// Floor sensor handling
			f := inputDevice.FloorSensor()
			if f != -1 && f != prevFloor {
				elevator.FSMOnFloorArrival(f)
			}
			prevFloor = f
	
			// Timer handling
			if elevator.TimerTimedOut() {
				elevator.TimerStop()
				elevator.FSMOnDoorTimeout()
			}
			// Sleep for input poll rate
			time.Sleep(time.Duration(inputPollRateMs) * time.Millisecond)
		}
	

}

func InitFSM(elevatorStateChannel chan Elevator, localElevator Elevator) {

	elevio.Init("localhost:15657", N_FLOORS)
	ElevatorID = getElevatorID()
	
	outputDevice = elevio.GetOutputDevice()

	elevatorStateChannel <- localElevator

	fmt.Println("FSM initialized for elevator:", ElevatorID)
}

func HallRequestsUninitialized() HallRequestsType {
	var hallRequests HallRequestsType
	hallRequests[elevator.N_FLOORS][elevator.N_BUTTONS] = false 
	return hallRequests
}

func convertDirnToMotor(d Dirn) elevio.MotorDirection {
	switch d {
	case "up":
		return elevio.MD_Up
	case "down":
		return elevio.MD_Down
	default:
		return elevio.MD_Stop
	}
}

// **setAllLights**: Oppdaterer alle heisknapper med riktig lysstatus
func setAllLights() {
	sharedState := GetSharedState()
	hallRequests := sharedState.HallRequests
	elevator := GetLocalElevator()

	for floor := 0; floor < N_FLOORS; floor++ { // må også få inn posisjonsindikator
		// Cab requests
		outputDevice.RequestButtonLight(floor, elevio.BT_Cab, elevator.CabRequests[floor])

		// Hall requests fra SharedState
		outputDevice.RequestButtonLight(floor, elevio.BT_HallUp, hallRequests[floor][B_HallUp])
		outputDevice.RequestButtonLight(floor, elevio.BT_HallDown, hallRequests[floor][B_HallDown])
	}
}

// **FSMStartMoving**: Kalles for å sjekke om heisen skal starte bevegelse
func FSMStartMoving() {
	elevator := GetLocalElevator()
	sharedState := GetSharedState()

	// Hvis heisen er idle og har forespørsler, velg retning og start motor
	if elevator.Behaviour == EB_Idle && hasRequests(elevator, sharedState.HallRequests) { 
		elevator.Dirn = requestsChooseDirection(elevator)

		if elevator.Dirn != D_Stop {
			elevator.Behaviour = EB_Moving
			outputDevice.MotorDirection(convertDirnToMotor(elevator.Dirn))
		}

		UpdateLocalElevator(elevator)
		UpdateSharedState()
	}
}

// **FSMOnInitBetweenFloors**: Kalles hvis heisen starter mellom etasjer
func FSMOnInitBetweenFloors() {
	elevator := GetLocalElevator()

	outputDevice.MotorDirection(elevio.MD_Down)
	elevator.Dirn = D_Down
	elevator.Behaviour = EB_Moving

	UpdateLocalElevator(elevator) // Oppdater lokalt
	UpdateSharedState()           // Synkroniser med SharedState, omgjøre til kanal!
}

// **FSMOnRequestButtonPress**: Kalles når en knapp trykkes
func FSMOnRequestButtonPress(btnFloor int, btnType elevio.ButtonType, localElevator *Elevator) {
	fmt.Printf("FSMOnRequestButtonPress(%d, %d)\n", btnFloor, btnType)

	if btnType == elevio.BT_Cab {
		localElevator.CabRequests[btnFloor] = true
	} else {
		newHallRequestChannel <- NewHallRequest{int: btnFloor, elevio.ButtonType: btnType}
	}
	elevatorStateChannel <- localElevator&


/*
	// 3. Oppdater SharedState med nye oppdrag
	if assignedRequests, exists := assignments[ElevatorID]; exists {
		GlobalState.mu.Lock()
		GlobalState.HRA.HallRequests = assignedRequests
		GlobalState.mu.Unlock()
		fmt.Printf("Updated assignments for %s: %+v\n", ElevatorID, assignedRequests)
	}
*/

	// 4. Start bevegelse hvis nødvendig
	FSMStartMoving()

	setAllLights()
}

// **FSMOnFloorArrival**: Kalles når heisen ankommer en ny etasje
func FSMOnFloorArrival(newFloor int) {
	fmt.Printf("\nFSMOnFloorArrival(%d)\n", newFloor)

	elevator := GetLocalElevator()
	sharedState := GetSharedState()

	// 1. Oppdater etasjeindikator og lagre ny etasje i lokal state
	elevator.Floor = newFloor
	outputDevice.FloorIndicator(newFloor)

	// 2. Sjekk om heisen skal stoppe
	if elevator.Behaviour == EB_Moving {
		if requestsShouldStop(elevator, sharedState.HallRequests) {
			outputDevice.MotorDirection(elevio.MD_Stop)
			outputDevice.DoorLight(true)
			elevator.Behaviour = EB_DoorOpen

			// Start dør-timer
			TimerStart(DoorOpenDurationS)

			// 3. Klarer requests på nåværende etasje
			elevator = requestsClearAtCurrentFloor(elevator)

		}
	}

	UpdateLocalElevator(elevator)
	UpdateSharedState()
	setAllLights()
}

// **FSMOnDoorTimeout**: Kalles når dør-timeren utløper
func FSMOnDoorTimeout() {
	fmt.Println("\nFSMOnDoorTimeout()")

	elevator := GetLocalElevator()
	sharedState := GetSharedState()

	// Hvis døren er åpen, bestem neste handling
	if elevator.Behaviour == EB_DoorOpen {
		elevator.Behaviour = EB_Idle

		// Be om ny oppdragsfordeling
		AssignRequestChannel <- struct{}{} // Sender tom struct for å trigge assigner
		assignments := <-AssignResultChannel

		// Oppdater hall requests basert på nye oppdrag
		if assignedRequests, exists := assignments[ElevatorID]; exists {
			GlobalState.mu.Lock()
			GlobalState.HRA.HallRequests = assignedRequests
			GlobalState.mu.Unlock()
		}

		// Bestem om heisen skal fortsette eller gå i idle
		if hasRequests(elevator, sharedState.HallRequests) {
			elevator.Behaviour = EB_Moving
			outputDevice.DoorLight(false)
			outputDevice.MotorDirection(convertDirnToMotor(elevator.Dirn))
		} else {
			elevator.Behaviour = EB_Idle
		}

		UpdateLocalElevator(elevator)
		UpdateSharedState() // Synkroniser
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
