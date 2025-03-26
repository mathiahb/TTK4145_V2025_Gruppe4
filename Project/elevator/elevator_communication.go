package elevator

import (
	"elevator_project/elevio"
)
type ElevatorChannels struct {
	Button        chan elevio.ButtonEvent
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