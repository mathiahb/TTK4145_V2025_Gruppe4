package elevator

import (
	elevio "Driver-Elevio"
	"shared_states" //hvordan koble ting sammen?
	"fmt"
	"time"
)

// FSM (Finite State Machine) styrer heisens tilstand og oppførsel basert på knappetrykk, etasjeanløp og dørlukkingshendelser.
// Den håndterer tilstander som Idle, DoorOpen og Moving, og bestemmer heisens retning og handlinger.


var outputDevice elevio.ElevOutputDevice //hva gjør denne?

func InitFSM(elevatorStateChannel chan Elevator, localElevator Elevator) Elevator{

	elevio.Init("localhost:15657", N_FLOORS)
	outputDevice = elevio.GetOutputDevice()
	fmt.Println("FSM initialized for elevator:", getElevatorID())

	elevatorStateChannel  <- localElevator

	return localElevator
}

// **FSMOnInitBetweenFloors**: Kalles hvis heisen starter mellom etasjer
func FSMOnInitBetweenFloors(localElevator Elevator, elevatorStateChannel chan Elevator, ) Elevator{

	outputDevice.MotorDirection(elevio.MD_Down)
	localElevator.Dirn = D_Down
	localElevator.Behaviour = EB_Moving

	elevatorStateChannel <- localElevator

	return localElevator
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

// **setAllLights**: Oppdaterer alle heisknapper med riktig lysstatus, håndterer ikke etasjeindikator
func setAllLights(localElevator Elevator, hallRequests HallRequestsType) {

	for floor := 0; floor < N_FLOORS; floor++ {
		// Cab requests
		outputDevice.RequestButtonLight(floor, elevio.BT_Cab, localElevator.CabRequests[floor])

		// Hall requests fra SharedState
		outputDevice.RequestButtonLight(floor, elevio.BT_HallUp, hallRequests[floor][B_HallUp])
		outputDevice.RequestButtonLight(floor, elevio.BT_HallDown, hallRequests[floor][B_HallDown])
	}
}

// **FSMStartMoving**: Kalles for å sjekke om heisen skal starte bevegelse
func FSMStartMoving(localElevator Elevator, hallRequests HallRequestsType, elevatorStateChannel chan Elevator) Elevator{

	// Hvis heisen er idle og har forespørsler, velg retning og start motor
	if localElevator.Behaviour == EB_Idle && hasRequests(localElevator, hallRequests) { 
		localElevator.Dirn = requestsChooseDirection(localElevator, hallRequests)

		if localElevator.Dirn != D_Stop {
			localElevator.Behaviour = EB_Moving
			outputDevice.MotorDirection(convertDirnToMotor(localElevator.Dirn))
		}
	}

	elevatorStateChannel <- localElevator
	setAllLights(localElevator, hallRequests)

	return localElevator
}


// **FSMOnRequestButtonPress**: Kalles når en knapp trykkes
func FSMButtonPress(btnFloor int, btnType elevio.ButtonType, localElevator Elevator, elevatorStateChannel chan Elevator, newHallRequestChannel chan HallRequestsType) Elevator {
	fmt.Printf("FSMOnRequestButtonPress(%d, %d)\n", btnFloor, btnType)

	if btnType == elevio.BT_Cab {
		localElevator.CabRequests[btnFloor] = true
		elevatorStateChannel  <- localElevator
	} else {

		var newHallRequest[btnFloor][btnType] HallRequestsType =  true //hmm hvordan lage
		newHallRequestChannel <- newHallRequest
	}
	return localElevator
}

// **FSMOnFloorArrival**: Kalles når heisen ankommer en ny etasje
func FSMOnFloorArrival(newFloor int, localElevator Elevator, hallRequests HallRequestsType, clearHallRequestChannel chan HallRequestsType, clearCabRequestChannel chan Elevator, elevatorStateChannel chan Elevator) (Elevator, HallRequestsType) {
	fmt.Printf("\nFSMOnFloorArrival(%d)\n", newFloor)

	// 1. lagre ny etasje i lokal state
	localElevator.Floor = newFloor

	// 2. Sjekk om heisen skal stoppe
	if localElevator.Behaviour == EB_Moving {
		if requestsShouldStop(localElevator, hallRequests) { // sharedState....
			outputDevice.MotorDirection(elevio.MD_Stop) // endre fra outputDevice?
			outputDevice.DoorLight(true)
			localElevator.Behaviour = EB_DoorOpen

			// Start dør-timer
			TimerStart(DoorOpenDuration)

			three_sec_delay := time.NewTimer(time.Second * 3)
			three_sec_delay.Reset(time.Second * 3) // sette ny timer, hvordan funker det egentlig?? bruke det som allerede er definert?
	

			// 3. Fjerner requests på nåværende etasje
			localElevator, hallRequests = requestsClearAtCurrentFloor(localElevator, hallRequests, clearHallRequestChannel, clearCabRequestChannel) 

		}
	}
	elevatorStateChannel  <- localElevator
	setAllLights(localElevator, hallRequests)

	return localElevator, hallRequests
}

// **FSMOnDoorTimeout**: Kalles når dør-timeren utløper
func FSMCloseDoors(localElevator Elevator, hallRequests HallRequestsType, elevatorStateChannel chan Elevator) Elevator{
	fmt.Println("\nFSMOnDoorTimeout()")

	// Hvis døren er åpen, bestem neste handling
	if localElevator.Behaviour == EB_DoorOpen {
		localElevator.Behaviour = EB_Idle
	
		// Bestem om heisen skal fortsette eller gå i idle
		if hasRequests(localElevator, hallRequests) { // burde ikke denne kalle på FSMStartMoving?
			localElevator.Behaviour = EB_Moving
			outputDevice.DoorLight(false)
			outputDevice.MotorDirection(convertDirnToMotor(elevator.Dirn))
		} else {
			localElevator.Behaviour = EB_Idle
		}
	
		elevatorStateChannel  <- localElevator
	}
	return localElevator
}

func FSMHoldDoors(localElevator Elevator, doorTimerIsUp chan bool, elevatorStateChannel chan Elevator) Elevator {
	
	three_sec_delay := time.NewTimer(time.Second * 3)
	three_sec_delay.Reset(time.Second * 3) // sette ny timer, hvordan funker det egentlig?? bruke det som allerede er definert?
	
	localElevator.Behaviour = EB_DoorOpen
	outputDevice.DoorLight(true)
	
	// må vente på timeren
	doorTimerIsUp <- true
	elevatorStateChannel <- localElevator

	return localElevator
}
	
