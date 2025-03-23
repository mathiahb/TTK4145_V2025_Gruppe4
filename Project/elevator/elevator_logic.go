package elevator

import (
	. "elevator_project/constants"
	"elevator_project/elevio"
	. "elevator_project/shared_states"
	"time"
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
func ElevatorThread(initElevator Elevator, elevatorChannels ElevatorChannels, toElevator ToElevator, fromElevator FromElevator) {
	//føler at det er litt initialisering/konfigurering som mangler

	var localElevator = initElevator               // lager et lokalt heisobjekt
	var hallRequests = HallRequestsUninitialized() // lager et tomt request-objekt
	var isObstructed = false
	threeSecTimer := time.NewTimer(time.Second * 3) // lager en timer som går i 3 sekunder
	threeSecTimer.Stop()                            // Så den ikke utløses før vi selv resetter den

	InitFSM() // shared state får vite at en heis eksisterer, kjenner ikke helt til poenget med resten av funksjonen

	// FSMOnInitBetweenFloors og turnOffAllLights må kjøres ved første oppstart

	for {
		select {

		case buttonEvent := <-elevatorChannels.Button:
			localElevator = FSMButtonPress(buttonEvent.Floor, buttonEvent.Button, localElevator, fromElevator.UpdateState, fromElevator.NewHallRequestChannel)

		case newFloor := <-elevatorChannels.Floor:
			localElevator, hallRequests = FSMOnFloorArrival(
				newFloor,
				localElevator,
				hallRequests,
				fromElevator.ClearHallRequestChannel,
				fromElevator.UpdateState,
				threeSecTimer,
			)

		case isObstructed = <-elevatorChannels.Obstruction:
			// Tømme kanalen for å unngå blokkering
			if localElevator.Behaviour == EB_DoorOpen {

				if !threeSecTimer.Stop() {
					<-threeSecTimer.C
				}

				if !isObstructed {
					threeSecTimer.Reset(3 * time.Second)
				}
			}

			// -- Timeren går ut etter 3 sek --
			// Skal alltid trigges når døren lukkes
		case <-threeSecTimer.C:
			if !isObstructed {
				localElevator = FSMCloseDoors(localElevator, hallRequests, fromElevator.UpdateState)
			}

		case hallRequests = <-toElevator.ApprovedHRAChannel: // fordi alle ordre kommer fra shared states
			localElevator = FSMStartMoving(localElevator, hallRequests, fromElevator.UpdateState)

		case sharedHallRequests := <-toElevator.UpdateHallRequestLights:
			setHallLights(localElevator, sharedHallRequests)

		case cabRequests := <-toElevator.ApprovedCabRequestsChannel:
			setCabLights(localElevator, cabRequests)
		}
	}
}
