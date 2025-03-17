package elevator

// ===================== CONSTANTS ===================== //
const (
	N_FLOORS          = 4
	N_BUTTONS         = 3
	DoorOpenDurationS = 3.0
)

// ===================== ENUMS & TYPES ===================== //
// Elevator direction
type Dirn string

const (
	D_Stop Dirn = "stop"
	D_Up        = "up"
	D_Down      = "down"
)

// Elevator behavior states
type ElevatorBehaviour string

const (
	EB_Idle     ElevatorBehaviour = "idle"
	EB_DoorOpen                   = "doorOpen"
	EB_Moving                     = "moving"
)

// Button types (Hall Up, Hall Down, Cab)
const (
	B_HallUp = iota
	B_HallDown
	B_Cab
)

// Elevator struct representing an elevator's state
type Elevator struct {
	Behaviour   ElevatorBehaviour `json:"behaviour"`
	Floor       int               `json:"floor"`
	Dirn        Dirn              `json:"direction"`
	CabRequests []bool            `json:"cabRequests"`
}

