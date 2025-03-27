package elevator

import (
	"elevator_project/constants"
)

func HallRequestsUninitialized() constants.HallRequestType {
	hallRequests := make(constants.HallRequestType, constants.N_FLOORS)
	return hallRequests
}

// requestsShouldStop checks if the elevator should stop at the current floor
// by checking if there are any requests for the current floor.
func requestsShouldStop(
	localElevator constants.Elevator,
	hallRequests constants.HallRequestType,
) bool {

	// 1. Always stop if there is a cab request at this floor
	if localElevator.CabRequests[localElevator.Floor] {
		return true
	}

	// 2. Stop if a hall request is assigned to this elevator at this floor
	if localElevator.Dirn == constants.D_Up && hallRequests[localElevator.Floor][constants.B_HallUp] {
		return true
	}
	if localElevator.Dirn == constants.D_Down && hallRequests[localElevator.Floor][constants.B_HallDown] {
		return true
	}

	// 3. If there are no more hall requests in this direction, stop
	if localElevator.Dirn == constants.D_Up {
		return !requests_above(localElevator, hallRequests)
	}

	if localElevator.Dirn == constants.D_Down {
		return !requests_below(localElevator, hallRequests)
	}

	return false
}

// requestsClearAtCurrentFloor clears hall- and cab requests at the current floor
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

// requests_above checks if there are any requests above the current floor
func requests_above(
	localElevator constants.Elevator,
	hallRequests constants.HallRequestType,
) bool {
	for f := localElevator.Floor + 1; f < constants.N_FLOORS; f++ {
		if hallRequests[f][constants.B_HallUp] || hallRequests[f][constants.B_HallDown] || localElevator.CabRequests[f] {
			return true
		}
	}
	return false
}

// requests_below checks if there are any requests below the current floor
func requests_below(
	localElevator constants.Elevator,
	hallRequests constants.HallRequestType,
) bool {
	for f := 0; f < localElevator.Floor; f++ {
		if hallRequests[f][constants.B_HallUp] || hallRequests[f][constants.B_HallDown] || localElevator.CabRequests[f] {
			return true
		}
	}
	return false
}

// hasRequests checks if the elevator has any requests
func hasRequests(
	localElevator constants.Elevator,
	hallRequests [][2]bool,
) bool {

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

// requestsChooseDirection chooses the direction the elevator should move
// based on the current floor and the requests in the system
func requestsChooseDirection(
	localElevator constants.Elevator,
	hallRequests constants.HallRequestType,
) constants.Dirn {

	// Checks if there are any requests above the elevator -> move up
	for f := localElevator.Floor + 1; f < constants.N_FLOORS; f++ {
		if hallRequests[f][constants.B_HallUp] || hallRequests[f][constants.B_HallDown] || localElevator.CabRequests[f] {
			return constants.D_Up
		}
	}

	// Checks if there are any requests below the elevator -> move down
	for f := 0; f < localElevator.Floor; f++ {
		if hallRequests[f][constants.B_HallUp] || hallRequests[f][constants.B_HallDown] || localElevator.CabRequests[f] {
			return constants.D_Down
		}
	}
	// If there are no requests, the elevator should stop
	return constants.D_Stop
}
