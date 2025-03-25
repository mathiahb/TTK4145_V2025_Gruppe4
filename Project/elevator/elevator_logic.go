package elevator

import (
	. "elevator_project/constants"
	"elevator_project/elevio"
	. "elevator_project/shared_states"
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
func ElevatorThread(initElevator Elevator, elevatorChannels ElevatorChannels, fromSharedState ToElevator, toSharedState FromElevator) {

	var hallRequests = HallRequestsUninitialized() // lager et tomt request-objekt
	var isObstructed = false
	threeSecTimer := time.NewTimer(time.Second * 3) // lager en timer som går i 3 sekunder
	threeSecTimer.Stop()
	isStuckTimer := time.NewTimer(time.Second * 5)                            // Så den ikke utløses før vi selv resetter den
	localElevator := InitFSM(initElevator, elevatorChannels, toSharedState, isStuckTimer)

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
				isStuckTimer,
			)

		case isObstructed = <-elevatorChannels.Obstruction:
			fmt.Printf("Obstruction switch: %v\n", isObstructed)

			// Tømme kanalen for å unngå blokkering
			if localElevator.Behaviour == EB_DoorOpen || localElevator.Behaviour == EB_Stuck_DoorOpen {


				if !threeSecTimer.Stop() {
					<-threeSecTimer.C
				}

				if !isObstructed {
					fmt.Printf("Door is not obstructed, closing door\n")
					threeSecTimer.Reset(3 * time.Second)
				
				} else if isObstructed{
					localElevator.Behaviour = EB_Stuck_DoorOpen 
					toSharedState.UpdateState <- localElevator
				}
			}

			// -- Timeren går ut etter 3 sek --
			// Skal alltid trigges når døren lukkes
		case <-threeSecTimer.C:
			fmt.Printf("Door timer expired, obstruction: %v\n", isObstructed)
			if !isObstructed {
				fmt.Printf("Door is not obstructed, closing door\n")
				
				if(localElevator.Behaviour == EB_Stuck_DoorOpen){
					localElevator.Behaviour = EB_DoorOpen 
				}

				localElevator = FSMCloseDoors(localElevator, hallRequests, toSharedState.UpdateState, isStuckTimer)
			}
		
		case <- isStuckTimer.C: // dersom heisen har vært kontinuerlig i bevegelse i mer enn 5 sek uten å 
			
			if(localElevator.Behaviour == EB_Moving){ // ekstra sjekk
				localElevator.Behaviour = EB_Stuck_Moving
				toSharedState.UpdateState <- localElevator 
			} 
			
			
		case hallRequests = <-fromSharedState.ApprovedHRAChannel: // fordi alle ordre kommer fra shared states
			localElevator = FSMStartMoving(localElevator, hallRequests, toSharedState.UpdateState, isStuckTimer)

		case sharedHallRequests := <-fromSharedState.UpdateHallRequestLights:
			setHallLights(sharedHallRequests)

		case cabRequests := <-fromSharedState.ApprovedCabRequestsChannel:
			setCabLights(cabRequests)
		}
	}
}
