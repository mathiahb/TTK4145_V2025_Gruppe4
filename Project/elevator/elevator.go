package elevator

import (
	"fmt"
)

// Konstante verdier som tilsvarer N_FLOORS og N_BUTTONS.
const (
	N_FLOORS  = 4
	N_BUTTONS = 3
)

// Definerer mulige retninger for heisen - type safety
type Dirn string

const (
	D_Stop Dirn = "up"
	D_Up        = "down"
	D_Down      = "stop"
)

// Definerer knappetyper (for eksempel hall up, hall down, cab).
const (
	B_HallUp = iota
	B_HallDown
	B_Cab
)

// Konverterer en Dirn-verdi til tekst for utskrift.
// func elevioDirnToString(d Dirn) string {
// 	switch d {
// 	case D_Stop:
// 		return "Stop"
// 	case D_Up:
// 		return "Up"
// 	case D_Down:
// 		return "Down"
// 	default:
// 		return "Undefined"
// 	}
// }

// ElevatorBehaviour beskriver heisens hovedtilstander.
type ElevatorBehaviour string

const (
	EB_Idle     ElevatorBehaviour = "EB_Idle"
	EB_DoorOpen                   = "EB_DoorOpen"
	EB_Moving                     = "EB_Moving"
)

// Konverterer ElevatorBehaviour til en tekstlig representasjon.
// func elevatorBehaviourToString(eb ElevatorBehaviour) string {
// 	switch eb {
// 	case EB_Idle:
// 		return "EB_Idle"
// 	case EB_DoorOpen:
// 		return "EB_DoorOpen"
// 	case EB_Moving:
// 		return "EB_Moving"
// 	default:
// 		return "EB_UNDEFINED"
// 	}
// }

// ClearRequestVariant definerer hvordan bestillinger (requests) "ryddes opp" når døren åpner.
type ClearRequestVariant int

const (
	CV_All    ClearRequestVariant = iota // Alle som venter, går inn.
	CV_InDirn                            // Kun de som skal i samme retning, går inn.
)

// Elevator beskriver hele heisens tilstand.
type Elevator struct {
	Behaviour   ElevatorBehaviour `json:"behaviour"`
	Floor       int               `json:"floor"`
	Dirn        Dirn              `json:"direction"`
	CabRequests []bool            `json:"cabRequests"`

	Config struct {
		ClearRequestVariant ClearRequestVariant
		DoorOpenDurationS   float64
	}
}

// Hall Request Assigner Input
type HRAInput struct {
	HallRequests [][2]bool           `json:"hallRequests"`
	States       map[string]Elevator `json:"states"`
}

// ElevatorPrint skriver ut heisens tilstand til konsollen.
func ElevatorPrint(es Elevator) {
	fmt.Println("  +--------------------+")
	fmt.Printf("  |floor = %-2d          |\n", es.Floor) // %-2d = to siffer, venstrejustert.
	fmt.Printf("  |dirn  = %-12.12s|\n", es.Dirn)        // %-12.12s = minst 12 tegn i bredde, maks 12 tegn, venstrejustert.
	fmt.Printf("  |behav = %-12.12s|\n", es.Behaviour)
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")

	// Går gjennom etasjer fra øverste til nederste.
	for f := N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < N_BUTTONS; btn++ {
			// Hopp over 'umulige' knapper (f.eks. "HallUp" i øverste etasje).
			if (f == N_FLOORS-1 && btn == B_HallUp) ||
				(f == 0 && btn == B_HallDown) {
				fmt.Print("|     ")
			} else {
				if es.CabRequests[f] == true {
					fmt.Print("|  #  ")
				} else {
					fmt.Print("|  -  ")
				}
			}
		}
		fmt.Println("|")
	}
	fmt.Println("  +--------------------+")
}

// ElevatorUninitialized oppretter en heis med standardverdier, og returnerer den.
func ElevatorUninitialized() Elevator {
	var e Elevator
	e.Floor = -1
	e.Dirn = "stop"
	e.Behaviour = "EB_Idle"
	e.Config.ClearRequestVariant = CV_All
	e.Config.DoorOpenDurationS = 3.0
	return e
}
