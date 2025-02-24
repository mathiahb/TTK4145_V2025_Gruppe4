package elevator

import (
	"fmt"

	elevio "Driver-Elevio"
)

// FSM (Finite State Machine) styrer heisens tilstand og oppførsel basert på knappetrykk, etasjeanløp og dørlukkingshendelser.
// Den håndterer tilstander som Idle, DoorOpen og Moving, og bestemmer heisens retning og handlinger.

// Må opprette en assigner, fjerne funksjoner som bestemmer logikk for valg av etasje - erstattes av kostfunksjonen, beholde logikk som gir retning.
// En kanal for assigning og en for meldinger. Enveis kommunikasjon.

// Global heistilstand og output-enhet fra elevio
var elevator Elevator
var outputDevice elevio.ElevOutputDevice

// **init**: Initialiserer heisens tilstandsmaskin
func init() {
	elevator = ElevatorUninitialized()

	// Henter konfigurasjon fra fil
	config := LoadConfig("elevator.con")
	elevator.Config.ClearRequestVariant = config.ClearRequestVariant
	elevator.Config.DoorOpenDurationS = config.DoorOpenDurationS

	// Henter hardware output-funksjoner fra elevio
	outputDevice = elevio.GetOutputDevice()
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
func setAllLights(e Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			outputDevice.RequestButtonLight(floor, elevio.ButtonType(btn), e.Requests[floor][btn] != 0)
		}
	}
}

// **FSMOnInitBetweenFloors**: Kalles hvis heisen starter mellom etasjer
func FSMOnInitBetweenFloors() {
	outputDevice.MotorDirection(elevio.MD_Down)
	elevator.Dirn = "down"
	elevator.Behaviour = EB_Moving
}

// **FSMOnRequestButtonPress**: Kalles når en knapp i heisen trykkes
func FSMOnRequestButtonPress(btnFloor int, btnType elevio.ButtonType) {
	fmt.Printf("\n\nFSMOnRequestButtonPress(%d, %d)\n", btnFloor, btnType)
	ElevatorPrint(elevator)

	switch elevator.Behaviour { // Hva gjør heisen nå?
	case EB_DoorOpen:
		if requestsShouldClearImmediately(elevator, btnFloor, btnType) {
			TimerStart(elevator.Config.DoorOpenDurationS)
		} else {
			elevator.Requests[btnFloor][btnType] = 1
		}

	case EB_Moving:
		elevator.Requests[btnFloor][btnType] = 1

	case EB_Idle:
		elevator.Requests[btnFloor][btnType] = 1
		pair := requestsChooseDirection(elevator)

		fmt.Printf("requestsChooseDirection returned Dirn=%v, Behaviour=%v\n", pair.Dirn, pair.Behaviour)

		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case EB_DoorOpen:
			outputDevice.DoorLight(true)
			TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requestsClearAtCurrentFloor(elevator)

		case EB_Moving:
			fmt.Printf("FSMOnRequestButtonPress: Setting motor direction to %v\n", elevio.MotorDirection(convertDirnToMotor(elevator.Dirn)))
			//outputDevice.MotorDirection(elevio.MotorDirection(elevator.Dirn))
			outputDevice.MotorDirection(elevio.MotorDirection(convertDirnToMotor(elevator.Dirn)))

		case EB_Idle:
			// Ikke gjør noe
		}
	}

	setAllLights(elevator)

	fmt.Println("\nNew state:")
	ElevatorPrint(elevator)
}

// **FSMOnFloorArrival**: Kalles når heisen ankommer en ny etasje
func FSMOnFloorArrival(newFloor int) {
	fmt.Printf("\n\nFSMOnFloorArrival(%d)\n", newFloor)
	ElevatorPrint(elevator)

	elevator.Floor = newFloor
	outputDevice.FloorIndicator(newFloor)

	if elevator.Behaviour == EB_Moving {
		if requestsShouldStop(elevator) {
			outputDevice.MotorDirection(elevio.MD_Stop)
			outputDevice.DoorLight(true)
			elevator = requestsClearAtCurrentFloor(elevator)
			TimerStart(elevator.Config.DoorOpenDurationS)
			setAllLights(elevator)
			elevator.Behaviour = EB_DoorOpen
		}
	}

	fmt.Println("\nNew state:")
	ElevatorPrint(elevator)
}

// **FSMOnDoorTimeout**: Kalles når dør-timeren utløper
func FSMOnDoorTimeout() {
	fmt.Println("\n\nFSMOnDoorTimeout()")
	ElevatorPrint(elevator)

	if elevator.Behaviour == EB_DoorOpen { // Hvis døra var åpen så bestemmer vi hva som skal skje videre.
		pair := requestsChooseDirection(elevator) // returnerer DirnBehaviourPair med ny retning og oppførsel
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour

		switch elevator.Behaviour { // Hva skal heisen gjøre nå - med oppdatert retning og oppførsel?
		case EB_DoorOpen:
			TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requestsClearAtCurrentFloor(elevator)
			setAllLights(elevator)

		case EB_Moving, EB_Idle:
			outputDevice.DoorLight(false)
			//outputDevice.MotorDirection(elevio.MotorDirection(elevator.Dirn))
			outputDevice.MotorDirection(elevio.MotorDirection(convertDirnToMotor(elevator.Dirn)))

		}
	}

	fmt.Println("\nNew state:")
	ElevatorPrint(elevator)
}
