package main

import (
	"elevator_project/constants"
	"elevator_project/elevator"
	"elevator_project/elevio"

	//"elevator_project/Network_Protocol/Network" // denne mappen er rotete! kalle pakke om mappe det samme kanskje??
	"elevator_project/shared_states"
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
		case constants.ARGV_TEST:
			is_testing = true
		case constants.ARGV_LISTENER_ONLY:
			// Should not connect an elevator, and should start a listener node.
			// Will print the shared state to screen, additionally will log the messages sent on the network.
		case constants.ARGV_BACKUP:
			// Spawned by local_heartbeat, will be listening to and sending a heartbeat to the main program
			// If loss of heartbeat from backup: Kill backup and spawn new
			// If loss of heartbeat from main: Kill main and takeover as new main.
		case constants.ARGV_ELEVATOR_ID:
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

	}

	elevatorChannels := elevator.MakeElevatorChannels()                                             // channels for communication within the different parts of the elevator
	betweenElevatorAndSharedStatesChannels := elevator.MakeBetweenElevatorAndSharedStatesChannels() //elevator <-> shared states communication
	//network <-> shared states communication
	//synchronizationChannels := network.New_SynchronizationChannels() // endre navn til Make, slik at det blir samsvar på tvers av moduler
	//twoPhaseCommitChannels := network.MakeTwoPhaseCommitChannels() // denne må Atle lage

	go elevio.PollButtons(elevatorChannels.Button)
	go elevio.PollFloorSensor(elevatorChannels.Floor)
	go elevio.PollObstructionSwitch(elevatorChannels.Obstruction)
	go elevator.ElevatorThread(elevatorChannels, betweenElevatorAndSharedStatesChannels)
	go shared_states.SharedStateThread(betweenElevatorAndSharedStatesChannels)
	//go network.NetworkThread(synchronizationChannels) // twoPhaseCommitChannels, skal også sendes til nettverket

}
