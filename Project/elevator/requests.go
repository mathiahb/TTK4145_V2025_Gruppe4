package elevator

import (
	. "elevator_project/constants"
)

// DirnBehaviourPair strukturen holder retningen og oppførselen for heisen
type DirnBehaviourPair struct {
	Dirn      Dirn
	Behaviour ElevatorBehaviour
}

func HallRequestsUninitialized() HallRequestType {
	var hallRequests HallRequestType
	hallRequests = make(HallRequestType, N_FLOORS)
	return hallRequests
	//
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
		return !requests_above(localElevator, hallRequests)
	}

	if localElevator.Dirn == D_Down {
		return !requests_below(localElevator, hallRequests)
	}

	return false
}

// requestsClearAtCurrentFloor rydder bestillinger på gjeldende etasje
// localElevator, hallRequests, clearHallRequestChannel, updateStateChannel
func requestsClearAtCurrentFloor(localElevator Elevator, hallRequests HallRequestType, clearHallRequestChannel chan HallRequestType, updateStateChannel chan Elevator) (Elevator, HallRequestType) {

	// Clear cab request at this floor
	if localElevator.CabRequests[localElevator.Floor] == true {

		localElevator.CabRequests[localElevator.Floor] = false
		updateStateChannel <- localElevator
	}

	switch localElevator.Dirn {
	case D_Up:
		if !requests_above(localElevator, hallRequests) && !hallRequests[localElevator.Floor][B_HallUp] == true {
			if hallRequests[localElevator.Floor][B_HallDown] == true {
				//clear hallrequest locally and update network
				clearHallRequest := make(HallRequestType, N_FLOORS)
				clearHallRequest[localElevator.Floor][B_HallDown] = true
				clearHallRequestChannel <- clearHallRequest
			}
		}
		if hallRequests[localElevator.Floor][B_HallUp] == true {
			//clear hallrequest locally and update network
			clearHallRequest := make(HallRequestType, N_FLOORS)
			clearHallRequest[localElevator.Floor][B_HallUp] = true
			clearHallRequestChannel <- clearHallRequest
		}
	case D_Down:
		if !requests_below(localElevator, hallRequests) && !hallRequests[localElevator.Floor][B_HallDown] == true {
			if hallRequests[localElevator.Floor][B_HallUp] == true {
				//clear hallrequest locally and update network
				clearHallRequest := make(HallRequestType, N_FLOORS)
				clearHallRequest[localElevator.Floor][B_HallUp] = true
				clearHallRequestChannel <- clearHallRequest
			}
		}
		if hallRequests[localElevator.Floor][B_HallDown] == true {
			//clear hallrequest locally and update network
			clearHallRequest := make(HallRequestType, N_FLOORS)
			clearHallRequest[localElevator.Floor][B_HallDown] = true
			clearHallRequestChannel <- clearHallRequest
		}
	case D_Stop:
		if hallRequests[localElevator.Floor][B_HallUp] == true {
			//clear hallrequest locally and update network
			clearHallRequest := make(HallRequestType, N_FLOORS)
			clearHallRequest[localElevator.Floor][B_HallUp] = true
			clearHallRequestChannel <- clearHallRequest
		}
		if hallRequests[localElevator.Floor][B_HallDown] == true {
			//clear hallrequest locally and update network
			clearHallRequest := make(HallRequestType, N_FLOORS)
			clearHallRequest[localElevator.Floor][B_HallDown] = true
			clearHallRequestChannel <- clearHallRequest
		}
	}

	return localElevator, hallRequests
}

func requests_above(localElevator Elevator, hallRequests HallRequestType) bool {
	for f := localElevator.Floor + 1; f < N_FLOORS; f++ {
		if hallRequests[f][B_HallUp] || hallRequests[f][B_HallDown] || localElevator.CabRequests[f] {
			return true
		}
	}
	return false
}

func requests_below(localElevator Elevator, hallRequests HallRequestType) bool {
	for f := 0; f < localElevator.Floor; f++ {
		if hallRequests[f][B_HallUp] || hallRequests[f][B_HallDown] || localElevator.CabRequests[f] {
			return true
		}
	}
	return false
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
func requestsChooseDirection(e Elevator, hallRequests HallRequestType) Dirn {

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
