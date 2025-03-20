package main

import (
	"Constants"
	elevator "Driver-Elevator"
	"fmt"
	"os"
)

func main() {
	var is_testing bool = false
	var id string = ""
	for i, arg := range os.Args {
		if i == 0 {
			continue
		}

		switch arg {
		case Constants.ARGV_TEST:
			is_testing = true
		case Constants.ARGV_LISTENER_ONLY:
			// Should not connect an elevator, and should start a listener node.
			// Will print the shared state to screen, additionally will log the messages sent on the network.
		case Constants.ARGV_BACKUP:
			// Spawned by local_heartbeat, will be listening to and sending a heartbeat to the main program
			// If loss of heartbeat from backup: Kill backup and spawn new
			// If loss of heartbeat from main: Kill main and takeover as new main.
		case Constants.ARGV_ELEVATOR_ID:
			id = os.Args[i+1]
		default:
			fmt.Printf("Unknown Arg: %s", arg)
		}
	}

	if id == "" {
		fmt.Println("Error. No id!")
		return
	}

	if is_testing {
		// tests.Test_Creating_Connection(id)
		// return
		elevator.InitElevator()

		go elevator.RequestAssigner()

		select {}

	}

	// channels and threads for communication within the different parts of the elevator
	buttonChannel := make(chan elevio.ButtonEvent) //het før drv_..., men det er feil konvensjon, må fikses opp
	floorsChannel := make(chan int)
	obstructionChannel := make(chan bool)
	doorTimerIsUp := make(chan bool)

	go elevio.PollButtons(buttonChannel)
	go elevio.PollFloorSensor(floorsChannel)
	go elevio.PollObstructionSwitch(obstructionChannel )



	/* channels for communication between modules */

	//elevator <-> shared states

	elevatorStateChannel  := make(chan Elevator) // fra elevator til shared states, sender tilstandene når en tilstand på heisen endres
	clearCabRequestChannel := make(chan Elevator)
	clearHallRequestChannel := make(chan HallRequestsType)
	approvedClearHallRequestsChannel := make(chan HallRequestsType)
	newHallRequestChannel := make(chan HallRequestType) // fra elevator til shared states, sender ny HallRequest, når knapp trykket inn
	approvedHallRequestChannel  := make(chan HallRequestType) // fra shared state til elevator, sender godkjent HallRequest etter konferering med nettverket


	//network <-> shared states communication
	
	startSynchChannel := make(chan struct{}) // fra nettverk til shared state, ønsker å starte synkronisering
	updatedSharedStateForSynchChannel := make(chan HRAType) // fra nettverk til shared state, inneholder oppdatert shared states
	sendSharedStateForSynchChannel := make(chan HRAType) //fra shared state til nettverk

	notifyNewHallRequestChannel  := make(chan HallRequestType)// shared state ønsker endring i hallRequest, sender til nettverk
	approvedNewHallRequestChannel := make(chan HallRequesType) // network sender en godkjent endring

	informNewStateChannel := make(chan Elevator) // shared state ønsker endring i state, sender til nettverk
	informedNewStateChannel := make(chan Elevator) // network sender en godkjent endring

	/* Threads for the different modules */
	go elevator(elevatorStateChannel, newHallRequestChannel, approvedHallRequestChannel) 
	go sharedState(elevatorStateChannel , newHallRequestChannel , approvedHallRequestChannel , startSynchChannel, updatedSharedStateForSynchChannel, sendSharedStateForSynchChannel, notifyNewHallRequestChannel , approvedNewHallRequestChannel, informNewStateChannel, informedNewStateChannel)
	go network(startSynchChannel, updatedSharedStateForSynchChannel, sendSharedStateForSynchChannel, notifyNewHallRequestChannel , approvedNewHallRequestChannel, informNewStateChannel, informedNewStateChannel)
	
}
