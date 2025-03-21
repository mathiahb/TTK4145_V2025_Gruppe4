package constants
import (
	"fmt"
	"os"
)
// --------------- ELEVATOR -------------------- //

// Elevator behavior states
type ElevatorBehaviour string

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
func GetElevatorID() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return "unknown_elevator"
	}
	return hostname
}