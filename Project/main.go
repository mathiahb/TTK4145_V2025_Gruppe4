package main

import (
	"elevator_project/constants"
	"elevator_project/elevator"
	"elevator_project/elevio"
	"elevator_project/network"
	"elevator_project/shared_states"
	"flag"
)

func main() {
	var portElevio int
	flag.IntVar(&portElevio, "elevatorPort", 15657, "Connect the elevio to a custom elevator port. Default: 15657")

	var nameExtension int
	flag.IntVar(&nameExtension, "name", 0, "Appends the name to the computer name, to be used when running multiple on the same computer. Default: 0")

	flag.Parse()
	constants.NameExtension = nameExtension

	fromSharedStateToNetwork := newFromSharedStateToNetwork()
	fromSharedStateToElevator := newFromSharedStateToElevator()
	toSharedStateFromNetwork := newToSharedStateFromNetwork()
	toSharedStateFromElevator := newToSharedStateFromElevator()

	initResponseChannel := make(chan constants.Elevator)

	network_channels := transferSharedStateChannelsToNetworkChannels(fromSharedStateToNetwork, toSharedStateFromNetwork)

	Node := network.New_Node(constants.GetElevatorID(), network_channels)
	Node.Connect() // Will start the initializing part.
	defer Node.Close()

	go shared_states.SharedStateThread(
		initResponseChannel,
		fromSharedStateToElevator, toSharedStateFromNetwork,
		fromSharedStateToNetwork, toSharedStateFromElevator,
	)

	initialElevator := <-initResponseChannel

	elevatorChannels := elevator.MakeElevatorChannels() // channels for communication within the different parts of the elevator

	go elevio.PollButtons(elevatorChannels.Button)
	go elevio.PollFloorSensor(elevatorChannels.Floor)
	go elevio.PollObstructionSwitch(elevatorChannels.Obstruction)
	go elevator.ElevatorThread(portElevio, initialElevator, elevatorChannels, fromSharedStateToElevator, toSharedStateFromElevator)
	//go network.NetworkThread(synchronizationChannels) // twoPhaseCommitChannels, skal også sendes til nettverket

	select {}
}
