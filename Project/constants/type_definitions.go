package constants

import (
	"fmt"
	"os"
	"strconv"
)

// --------------- ELEVATOR -------------------- //

// Elevator behavior states
type ElevatorBehaviour string

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
type HRAType struct { // Hall request assignment type
	HallRequests HallRequestType     `json:"hallRequests"`
	States       map[string]Elevator `json:"states"`
}

// getElevatorID returns a unique identifier for this elevator instance.
var NameExtension int = 0 // Set by main

func GetElevatorID() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return "unknown_elevator" + strconv.Itoa(NameExtension)
	}
	return hostname + strconv.Itoa(NameExtension)
}
