package elevator

import (
	"elevator_project/constants"
	"elevator_project/elevio"
	"elevator_project/shared_states"
	"fmt"
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
func ElevatorThread(initElevator constants.Elevator, elevatorChannels ElevatorChannels, fromSharedState shared_states.ToElevator, toSharedState shared_states.FromElevator) {
	//føler at det er litt initialisering/konfigurering som mangler

	var localElevator = initElevator               // lager et lokalt heisobjekt
	var hallRequests = HallRequestsUninitialized() // lager et tomt request-objekt
	var isObstructed = false
	threeSecTimer := time.NewTimer(time.Second * 3) // lager en timer som går i 3 sekunder
	threeSecTimer.Stop()                            // Så den ikke utløses før vi selv resetter den

	InitFSM() // shared state får vite at en heis eksisterer, kjenner ikke helt til poenget med resten av funksjonen

	// FSMOnInitBetweenFloors og turnOffAllLights må kjøres ved første oppstart
	turnOffAllLights() // starter med alle lys avslått

	if localElevator.Floor == -1 {
		localElevator = FSMOnInitBetweenFloors(localElevator, toSharedState.UpdateState)
	}

	for {
		select {

		case buttonEvent := <-elevatorChannels.Button:
			localElevator = FSMButtonPress(buttonEvent.Floor, buttonEvent.Button, localElevator, toSharedState.UpdateState, toSharedState.NewHallRequestChannel)

		case newFloor := <-elevatorChannels.Floor:
			localElevator, hallRequests = FSMOnFloorArrival(
				newFloor,
				localElevator,
				hallRequests,
				toSharedState.ClearHallRequestChannel,
				toSharedState.UpdateState,
				threeSecTimer,
			)

		case isObstructed = <-elevatorChannels.Obstruction:
			fmt.Printf("Obstruction switch: %v\n", isObstructed)

			if localElevator.Behaviour == constants.EB_DoorOpen {

				// Stop timer first to avoid channel blocking.
				if !threeSecTimer.Stop() {
					select {
					case <-threeSecTimer.C:
					default:
					}
				}

				if !isObstructed {
					fmt.Printf("Door is not obstructed, closing door\n")
					threeSecTimer.Reset(3 * time.Second)
				}
			}

			// -- Timeren går ut etter 3 sek --
			// Skal alltid trigges når døren lukkes
		case <-threeSecTimer.C:
			fmt.Printf("Door timer expired, obstruction: %v\n", isObstructed)
			if !isObstructed {
				fmt.Printf("Door is not obstructed, closing door\n")
				localElevator = FSMCloseDoors(localElevator, hallRequests, toSharedState.UpdateState, threeSecTimer, toSharedState.ClearHallRequestChannel, toSharedState.UpdateState)
			}

		case hallRequests = <-fromSharedState.ApprovedHRAChannel: // fordi alle ordre kommer fra shared states
			localElevator = FSMStartMoving(localElevator, hallRequests, toSharedState.UpdateState, threeSecTimer, toSharedState.ClearHallRequestChannel, toSharedState.UpdateState)

		case sharedHallRequests := <-fromSharedState.UpdateHallRequestLights:
			setHallLights(sharedHallRequests)

		case cabRequests := <-fromSharedState.ApprovedCabRequestsChannel:
			setCabLights(cabRequests)
		}
	}
}
