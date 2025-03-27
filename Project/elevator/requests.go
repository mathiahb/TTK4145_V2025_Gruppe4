package elevator

import (
	"elevator_project/common"
)

func HallRequestsUninitialized() common.HallRequestType {
	hallRequests := make(common.HallRequestType, common.N_FLOORS)
	return hallRequests
}

// requestsShouldStop checks if the elevator should stop at the current floor
// by checking if there are any requests for the current floor.
func requestsShouldStop(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
) bool {

	// 1. Always stop if there is a cab request at this floor
	if localElevator.CabRequests[localElevator.Floor] {
		return true
	}

	// 2. Stop if a hall request is assigned to this elevator at this floor
	if localElevator.Dirn == common.D_Up && hallRequests[localElevator.Floor][common.B_HallUp] {
		return true
	}
	if localElevator.Dirn == common.D_Down && hallRequests[localElevator.Floor][common.B_HallDown] {
		return true
	}

	// 3. If there are no more hall requests in this direction, stop
	if localElevator.Dirn == common.D_Up {
		return !requests_above(localElevator, hallRequests)
	}

	if localElevator.Dirn == common.D_Down {
		return !requests_below(localElevator, hallRequests)
	}

	return false
}

// requestsClearAtCurrentFloor clears hall- and cab requests at the current floor
func requestsClearAtCurrentFloor(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
	ClearHallRequest chan common.HallRequestType,
	UpdateState chan common.Elevator,
) (common.Elevator, common.HallRequestType) {

	// Clear cab request at this floor
	if localElevator.CabRequests[localElevator.Floor] {

		localElevator.CabRequests[localElevator.Floor] = false
		UpdateState <- localElevator
	}

	switch localElevator.Dirn {
	case common.D_Up:
		if !requests_above(localElevator, hallRequests) && !hallRequests[localElevator.Floor][common.B_HallUp] {
			if hallRequests[localElevator.Floor][common.B_HallDown] {
				//clear hallrequest locally and update network
				clearHallRequest := make(common.HallRequestType, common.N_FLOORS)
				clearHallRequest[localElevator.Floor][common.B_HallDown] = true
				ClearHallRequest <- clearHallRequest
			}
		}
		if hallRequests[localElevator.Floor][common.B_HallUp] {
			//clear hallrequest locally and update network
			clearHallRequest := make(common.HallRequestType, common.N_FLOORS)
			clearHallRequest[localElevator.Floor][common.B_HallUp] = true
			ClearHallRequest <- clearHallRequest
		}
	case common.D_Down:
		if !requests_below(localElevator, hallRequests) && !hallRequests[localElevator.Floor][common.B_HallDown] {
			if hallRequests[localElevator.Floor][common.B_HallUp] {
				//clear hallrequest locally and update network
				clearHallRequest := make(common.HallRequestType, common.N_FLOORS)
				clearHallRequest[localElevator.Floor][common.B_HallUp] = true
				ClearHallRequest <- clearHallRequest
			}
		}
		if hallRequests[localElevator.Floor][common.B_HallDown] {
			//clear hallrequest locally and update network
			clearHallRequest := make(common.HallRequestType, common.N_FLOORS)
			clearHallRequest[localElevator.Floor][common.B_HallDown] = true
			ClearHallRequest <- clearHallRequest
		}
	case common.D_Stop:
		if hallRequests[localElevator.Floor][common.B_HallUp] {
			//clear hallrequest locally and update network
			clearHallRequest := make(common.HallRequestType, common.N_FLOORS)
			clearHallRequest[localElevator.Floor][common.B_HallUp] = true
			ClearHallRequest <- clearHallRequest
		}
		if hallRequests[localElevator.Floor][common.B_HallDown] {
			//clear hallrequest locally and update network
			clearHallRequest := make(common.HallRequestType, common.N_FLOORS)
			clearHallRequest[localElevator.Floor][common.B_HallDown] = true
			ClearHallRequest <- clearHallRequest
		}
	}

	return localElevator, hallRequests
}

// requests_above checks if there are any requests above the current floor
func requests_above(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
) bool {
	for f := localElevator.Floor + 1; f < common.N_FLOORS; f++ {
		if hallRequests[f][common.B_HallUp] || hallRequests[f][common.B_HallDown] || localElevator.CabRequests[f] {
			return true
		}
	}
	return false
}

// requests_below checks if there are any requests below the current floor
func requests_below(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
) bool {
	for f := 0; f < localElevator.Floor; f++ {
		if hallRequests[f][common.B_HallUp] || hallRequests[f][common.B_HallDown] || localElevator.CabRequests[f] {
			return true
		}
	}
	return false
}

// hasRequests checks if the elevator has any requests
func hasRequests(
	localElevator common.Elevator,
	hallRequests [][2]bool,
) bool {

	// Sjekk cab requests
	for _, request := range localElevator.CabRequests {
		if request {
			return true
		}
	}

	// Sjekk hall requests
	for floor := 0; floor < common.N_FLOORS; floor++ {
		if hallRequests[floor][common.B_HallUp] || hallRequests[floor][common.B_HallDown] {
			return true
		}
	}

	return false
}

// requestsChooseDirection chooses the direction the elevator should move
// based on the current floor and the requests in the system
func requestsChooseDirection(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
) common.Dirn {

	// Checks if there are any requests above the elevator -> move up
	for f := localElevator.Floor + 1; f < common.N_FLOORS; f++ {
		if hallRequests[f][common.B_HallUp] || hallRequests[f][common.B_HallDown] || localElevator.CabRequests[f] {
			return common.D_Up
		}
	}

	// Checks if there are any requests below the elevator -> move down
	for f := 0; f < localElevator.Floor; f++ {
		if hallRequests[f][common.B_HallUp] || hallRequests[f][common.B_HallDown] || localElevator.CabRequests[f] {
			return common.D_Down
		}
	}
	// If there are no requests, the elevator should stop
	return common.D_Stop
}
