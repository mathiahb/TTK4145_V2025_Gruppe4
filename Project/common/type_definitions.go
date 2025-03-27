package common

import (
	"fmt"
	"os"
	"strconv"
)

// --------------- ELEVATOR -------------------- //

// Elevator behavior states
type ElevatorBehaviour string

// Checks if the elevator is stuck based on behaviour
func (behavior ElevatorBehaviour) IsStuck() bool {
	return behavior == EB_Stuck_DoorOpen || behavior == EB_Stuck_Moving
}

// Elevator direction
type Dirn string

// Elevator struct representing an elevator's state
type Elevator struct {
	Behaviour   ElevatorBehaviour `json:"behaviour"`
	Floor       int               `json:"floor"`
	Dirn        Dirn              `json:"direction"`
	CabRequests []bool            `json:"cabRequests"`
}

type HallRequestType [][2]bool

// --------------- SHARED STATES -------------------- //

// HRAType is on the format that the HRA request for its input. The HRA also return a HRAType.
type HRAType struct { 
	HallRequests HallRequestType     `json:"hallRequests"`
	States       map[string]Elevator `json:"states"`
}

var NameExtension int = 0 // Set by main

// GetElevatorID returns a unique identifier for this elevator instance.
func GetElevatorID() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return "unknown_elevator" + strconv.Itoa(NameExtension)
	}
	return hostname + strconv.Itoa(NameExtension)
}
