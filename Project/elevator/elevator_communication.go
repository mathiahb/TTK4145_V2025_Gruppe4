package elevator

import (
	"elevator_project/elevio"
)
// Channels that communicate from elevio to the elevator.
type ElevatorChannels struct {
	Button        chan elevio.ButtonEvent
	Floor         chan int
	Obstruction   chan bool
}

func MakeElevatorChannels() ElevatorChannels {
	return ElevatorChannels{
		Button:        make(chan elevio.ButtonEvent),
		Floor:         make(chan int),
		Obstruction:   make(chan bool),
	}
}