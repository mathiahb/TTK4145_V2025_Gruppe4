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
	Node.Connect()
	defer Node.Close()

	go shared_states.SharedStateThread(
		initResponseChannel,
		fromSharedStateToElevator, toSharedStateFromNetwork,
		fromSharedStateToNetwork, toSharedStateFromElevator,
	)

	initialElevator := <-initResponseChannel

	elevatorChannels := elevator.MakeElevatorChannels()                                             // channels for communication within the different parts of the elevator
	betweenElevatorAndSharedStatesChannels := elevator.MakeBetweenElevatorAndSharedStatesChannels() //elevator <-> shared states communication
	//network <-> shared states communication
	//synchronizationChannels := network.New_SynchronizationChannels() // endre navn til Make, slik at det blir samsvar på tvers av moduler
	//twoPhaseCommitChannels := network.MakeTwoPhaseCommitChannels() // denne må Atle lage

	go elevio.PollButtons(elevatorChannels.Button)
	go elevio.PollFloorSensor(elevatorChannels.Floor)
	go elevio.PollObstructionSwitch(elevatorChannels.Obstruction)
	go elevator.ElevatorThread(elevatorChannels, betweenElevatorAndSharedStatesChannels)
	//go network.NetworkThread(synchronizationChannels) // twoPhaseCommitChannels, skal også sendes til nettverket

	select {}
}
