package elevator

import (
	//. "elevator_project/constants"
	"elevator_project/elevio"
	"elevator_project/shared_states"
)

type ElevatorChannels struct { // channels and threads for communication within the different parts of the elevator
	Button        chan elevio.ButtonEvent // okei, så problemet har vært med importering av elevio
	Floor         chan int
	Obstruction   chan bool
	DoorTimerIsUp chan bool
}

func MakeElevatorChannels() ElevatorChannels {
	return ElevatorChannels{
		Button:        make(chan elevio.ButtonEvent),
		Floor:         make(chan int),
		Obstruction:   make(chan bool),
		DoorTimerIsUp: make(chan bool),
	}
}

// må legge til alle kanalene
func ElevatorThread(elevatorChannels ElevatorChannels, betweenElevatorAndSharedStatesChannels shared_states.BetweenElevatorAndSharedStatesChannels) {
	//føler at det er litt initialisering/konfigurering som mangler

	var localElevator = ElevatorUninitialized()    // lager et lokalt heisobjekt
	var hallRequests = HallRequestsUninitialized() // lager et tomt request-objekt
	var isObstructed = false

	localElevator = InitFSM(elevatorStateChannel, localElevator) // shared state får vite at en heis eksisterer, kjenner ikke helt til poenget med resten av funksjonen
	if inputDevice.FloorSensor() == -1 {                         // fase ut inputDevice?
		localElevator = FSMOnInitBetweenFloors(localElevator, elevatorStateChannel)
	}

	for {
		select {

		case buttonEvent := <-Button:
			localElevator = FSMButtonPress(buttonEvent.btnFloor, buttonEvent.btnType, localElevator, elevatorStateChannel, newHallRequestChannel)

		case newFloor := <-Foor:
			localElevator, hallRequests = FSMOnFloorArrival(newFloor, localElevator, hallRequests, clearHallRequestChannel, clearCabRequestChannel, elevatorStateChannel)

		case isObstructed = <-Obstruction: // dette føles rart, men er nødt til å vite om noen prøver å obstructe døren

			if isObstructed {

			}

		case <-doorTimerIsUp:

			if !isObstructed {
				localElevator = FSMCloseDoors(localElevator, hallRequests, elevatorStateChannel)
			}

		case hallRequests <- approvedHallRequestChannel: // fordi alle ordre kommer fra shared states
			localElevator = FSMStartMoving(localElevator, hallRequests, elevatorStateChannel)

		}
	}
}
