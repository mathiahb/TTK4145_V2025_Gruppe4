package elevator

// DirnBehaviourPair strukturen holder retningen og oppførselen for heisen
type DirnBehaviourPair struct {
	Dirn      Dirn
	Behaviour ElevatorBehaviour
}

func HallRequestsUninitialized() HallRequestsType {
	var hallRequests HallRequestsType
	hallRequests[elevator.N_FLOORS][elevator.N_BUTTONS] = false 
	return hallRequests
}

// requestsShouldStop sjekker om heisen skal stoppe på nåværende etasje
func requestsShouldStop(localElevator Elevator, hallRequests [][2]bool) bool {

	// 1. Stopp alltid hvis en cab-request er på denne etasjen
	if localElevator.CabRequests[localElevator.Floor] {
		return true
	}

	// 2. Stopp hvis hall-forespørsel er tildelt denne heisen i denne etasjen
	if localElevator.Dirn == D_Up && hallRequests[localElevator.Floor][B_HallUp] {
		return true
	}
	if localElevator.Dirn == D_Down && hallRequests[localElevator.Floor][B_HallDown] {
		return true
	}

	// 3. Hvis det ikke er flere hall requests i denne retningen, stopp
	if localElevator.Dirn == D_Up {
		for floor := localElevator.Floor + 1; floor < N_FLOORS; floor++ {
			if hallRequests[floor][B_HallUp] || hallRequests[floor][B_HallDown] {
				return false // Fortsett å bevege seg
			}
		}
		return true // Stopp siden ingen flere hall requests i denne retningen
	}

	if localElevator.Dirn == D_Down {
		for floor := 0; floor < localElevator.Floor; floor++ {
			if hallRequests[floor][B_HallUp] || hallRequests[floor][B_HallDown] {
				return false // Fortsett å bevege seg
			}
		}
		return true // Stopp siden ingen flere hall requests i denne retningen
	}

	return false
}

// requestsClearAtCurrentFloor rydder bestillinger på gjeldende etasje
func requestsClearAtCurrentFloor(localElevator Elevator, hallRequests HallRequestsType, clearHallRequestChannel chan HallRequestsType, clearCabRequestChannel chan Elevator) (Elevator, HallRequestsType){
	
	// Clear cab request at this floor
	if(localElevator.CabRequests[localElevator.Floor] == true){
		localElevator.CabRequests[localElevator.Floor] = false
		clearCabRequestChannel <- localElevator
	}else {
		//cleare hallrequest locally and update network
		hallRequests[hallRequest.floor][hallRequest.button] = false // hvordan oppdatere?
		clearHallRequestChannel <- localElevator.HallRequests[hallRequest.floor][hallRequest.button]
	}

	return localElevator, hallRequests
}

func hasRequests(e Elevator, hallRequests [][2]bool) bool {

	// Sjekk cab requests
	for _, request := range e.CabRequests {
		if request {
			return true
		}
	}

	// Sjekk hall requests
	for floor := 0; floor < N_FLOORS; floor++ {
		if hallRequests[floor][B_HallUp] || hallRequests[floor][B_HallDown] {
			return true
		}
	}

	return false
}

// requestsChooseDirection velger retning basert på forespørsler
func requestsChooseDirection(e Elevator, hallRequests HallRequestsType) Dirn {

	// Hvis det er noen forespørsler over heisen
	for f := e.Floor + 1; f < N_FLOORS; f++ {
		if hallRequests[f][B_HallUp] || hallRequests[f][B_HallDown] || e.CabRequests[f] {
			return D_Up
		}
	}

	// Hvis det er noen forespørsler under heisen
	for f := 0; f < e.Floor; f++ {
		if hallRequests[f][B_HallUp] || hallRequests[f][B_HallDown] || e.CabRequests[f] {
			return D_Down
		}
	}
	// Ingen forespørsler, stopp
	return D_Stop
}
