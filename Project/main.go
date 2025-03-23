package main

import (
	"elevator_project/constants"
	"elevator_project/elevator"
	"elevator_project/elevio"
	"elevator_project/network"
	"elevator_project/shared_states"
)

func main() {
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
	go elevator.ElevatorThread(initialElevator, elevatorChannels, fromSharedStateToElevator, toSharedStateFromElevator)
	//go network.NetworkThread(synchronizationChannels) // twoPhaseCommitChannels, skal også sendes til nettverket

	select {}
}
