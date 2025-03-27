package elevator

import (
	"elevator_project/constants"
	"elevator_project/shared_states"
)

func HallRequestsUninitialized() constants.HallRequestType {
	hallRequests := make(constants.HallRequestType, constants.N_FLOORS)
	return hallRequests
}

// requestsShouldStop sjekker om heisen skal stoppe på nåværende etasje
func requestsShouldStop(
	localElevator constants.Elevator, 
	hallRequests constants.HallRequestType,
) bool {

	// 1. Stopp alltid hvis en cab-request er på denne etasjen
	if localElevator.CabRequests[localElevator.Floor] {
		return true
	}

	// 2. Stopp hvis hall-forespørsel er tildelt denne heisen i denne etasjen
	if localElevator.Dirn == constants.D_Up && hallRequests[localElevator.Floor][constants.B_HallUp] {
		return true
	}
	if localElevator.Dirn == constants.D_Down && hallRequests[localElevator.Floor][constants.B_HallDown] {
		return true
	}

	// 3. Hvis det ikke er flere hall requests i denne retningen, stopp
	if localElevator.Dirn == constants.D_Up {
		return !requests_above(localElevator, hallRequests)
	}

	if localElevator.Dirn == constants.D_Down {
		return !requests_below(localElevator, hallRequests)
	}

	return false
}

// requestsClearAtCurrentFloor rydder bestillinger på gjeldende etasje
// localElevator, hallRequests, ClearHallRequest, updateStateChannel
func requestsClearAtCurrentFloor(
	localElevator constants.Elevator, 
	hallRequests constants.HallRequestType, 
	ClearHallRequest chan constants.HallRequestType, 
	UpdateState chan constants.Elevator,
) (constants.Elevator, constants.HallRequestType) {

	// Clear cab request at this floor
	if localElevator.CabRequests[localElevator.Floor] {

		localElevator.CabRequests[localElevator.Floor] = false
		UpdateState <- localElevator
	}

	switch localElevator.Dirn {
	case constants.D_Up:
		if !requests_above(localElevator, hallRequests) && !hallRequests[localElevator.Floor][constants.B_HallUp] {
			if hallRequests[localElevator.Floor][constants.B_HallDown] {
				//clear hallrequest locally and update network
				clearHallRequest := make(constants.HallRequestType, constants.N_FLOORS)
				clearHallRequest[localElevator.Floor][constants.B_HallDown] = true
				ClearHallRequest <- clearHallRequest
			}
		}
		if hallRequests[localElevator.Floor][constants.B_HallUp] {
			//clear hallrequest locally and update network
			clearHallRequest := make(constants.HallRequestType, constants.N_FLOORS)
			clearHallRequest[localElevator.Floor][constants.B_HallUp] = true
			ClearHallRequest <- clearHallRequest
		}
	case constants.D_Down:
		if !requests_below(localElevator, hallRequests) && !hallRequests[localElevator.Floor][constants.B_HallDown] {
			if hallRequests[localElevator.Floor][constants.B_HallUp] {
				//clear hallrequest locally and update network
				clearHallRequest := make(constants.HallRequestType, constants.N_FLOORS)
				clearHallRequest[localElevator.Floor][constants.B_HallUp] = true
				ClearHallRequest <- clearHallRequest
			}
		}
		if hallRequests[localElevator.Floor][constants.B_HallDown] {
			//clear hallrequest locally and update network
			clearHallRequest := make(constants.HallRequestType, constants.N_FLOORS)
			clearHallRequest[localElevator.Floor][constants.B_HallDown] = true
			ClearHallRequest <- clearHallRequest
		}
	case constants.D_Stop:
		if hallRequests[localElevator.Floor][constants.B_HallUp] {
			//clear hallrequest locally and update network
			clearHallRequest := make(constants.HallRequestType, constants.N_FLOORS)
			clearHallRequest[localElevator.Floor][constants.B_HallUp] = true
			ClearHallRequest <- clearHallRequest
		}
		if hallRequests[localElevator.Floor][constants.B_HallDown] {
			//clear hallrequest locally and update network
			clearHallRequest := make(constants.HallRequestType, constants.N_FLOORS)
			clearHallRequest[localElevator.Floor][constants.B_HallDown] = true
			ClearHallRequest <- clearHallRequest
		}
	}

	return localElevator, hallRequests
}

func requests_above(localElevator constants.Elevator, hallRequests constants.HallRequestType) bool {
	for f := localElevator.Floor + 1; f < constants.N_FLOORS; f++ {
		if hallRequests[f][constants.B_HallUp] || hallRequests[f][constants.B_HallDown] || localElevator.CabRequests[f] {
			return true
		}
	}
	return false
}

func requests_below(localElevator constants.Elevator, hallRequests constants.HallRequestType) bool {
	for f := 0; f < localElevator.Floor; f++ {
		if hallRequests[f][constants.B_HallUp] || hallRequests[f][constants.B_HallDown] || localElevator.CabRequests[f] {
			return true
		}
	}
	return false
}

func hasRequests(localElevator constants.Elevator, hallRequests [][2]bool) bool {

	// Sjekk cab requests
	for _, request := range localElevator.CabRequests {
		if request {
			return true
		}
	}

	// Sjekk hall requests
	for floor := 0; floor < constants.N_FLOORS; floor++ {
		if hallRequests[floor][constants.B_HallUp] || hallRequests[floor][constants.B_HallDown] {
			return true
		}
	}

	return false
}

// requestsChooseDirection velger retning basert på forespørsler
func requestsChooseDirection(localElevator constants.Elevator, hallRequests constants.HallRequestType) constants.Dirn {

	// Hvis det er noen forespørsler over heisen
	for f := localElevator.Floor + 1; f < constants.N_FLOORS; f++ {
		if hallRequests[f][constants.B_HallUp] || hallRequests[f][constants.B_HallDown] || localElevator.CabRequests[f] {
			return constants.D_Up
		}
	}

	// Hvis det er noen forespørsler under heisen
	for f := 0; f < localElevator.Floor; f++ {
		if hallRequests[f][constants.B_HallUp] || hallRequests[f][constants.B_HallDown] || localElevator.CabRequests[f] {
			return constants.D_Down
		}
	}
	// Ingen forespørsler under eller over, stopp
	return constants.D_Stop
}
