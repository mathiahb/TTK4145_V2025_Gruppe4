package elevator

import (
	"fmt"
	"os"
)



r


// Retrieve a copy of all elevator states
func GetLocalElevator() Elevator {
	// Function can become useful when/if local elevator is replaced
	return localElevator
}

// Oppdaterer SharedState med denne heisens lokale tilstand
func UpdateSharedState() { //denne må endres til å interagere med en anne modul (kanal)
	GlobalState.mu.Lock()
	defer GlobalState.mu.Unlock()
	GlobalState.HRA.States[getElevatorID()] = localElevator
}


// Henter en kopi av hele SharedState
func GetSharedState() HRAInput {
	GlobalState.mu.Lock()
	defer GlobalState.mu.Unlock()
	return GlobalState.HRA
}

// getElevatorID returns a unique identifier for this elevator instance.
func getElevatorID() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return "unknown_elevator"
	}
	return hostname
}

// ===================== ELEVATOR HELPERS ===================== //
// Initialize a new elevator with default values
func ElevatorUninitialized() Elevator {
	return Elevator{
		Floor:       -1,
		Dirn:        D_Stop,
		Behaviour:   EB_Idle,
		CabRequests: make([]bool, N_FLOORS),
	}
}

// Print elevator state to the console
func ElevatorPrint(es Elevator) {
	fmt.Println("  +--------------------+")
	fmt.Printf("  | Floor  = %-2d        |\n", es.Floor)
	fmt.Printf("  | Dirn   = %-12.12s |\n", es.Dirn)
	fmt.Printf("  | Behav  = %-12.12s |\n", es.Behaviour)
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | Up  | Down | Cab |")

	for f := N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < N_BUTTONS; btn++ {
			if (f == N_FLOORS-1 && btn == B_HallUp) || (f == 0 && btn == B_HallDown) {
				fmt.Print("|     ")
			} else {
				if es.CabRequests[f] {
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


