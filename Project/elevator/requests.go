package elevator

import (
	"elevator_project/common"
)

// hallRequestsUninitialized returns an empty hall request object.
func hallRequestsUninitialized() common.HallRequestType {
	hallRequests := make(common.HallRequestType, common.N_FLOORS)
	return hallRequests
}

// RequestsShouldStop checks if the elevator should stop at the current floor
// by checking if there are any requests for the current floor.
func requestsShouldStop(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
) bool {

	if localElevator.CabRequests[localElevator.Floor] {
		return true
	}

	if localElevator.Dirn == common.D_Up && hallRequests[localElevator.Floor][common.B_HallUp] {
		return true
	}
	if localElevator.Dirn == common.D_Down && hallRequests[localElevator.Floor][common.B_HallDown] {
		return true
	}

	if localElevator.Dirn == common.D_Up {
		return !requestsAbove(localElevator, hallRequests)
	}

	if localElevator.Dirn == common.D_Down {
		return !requestsBelow(localElevator, hallRequests)
	}

	return false
}

// RequestsClearAtCurrentFloor clears hall- and cab requests at the current floor locally and updates the network
func requestsClearAtCurrentFloor(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
	ClearHallRequest chan common.HallRequestType,
	UpdateState chan common.Elevator,
) (common.Elevator, common.HallRequestType) {

	if localElevator.CabRequests[localElevator.Floor] {

		localElevator.CabRequests[localElevator.Floor] = false
		UpdateState <- localElevator
	}

	switch localElevator.Dirn {
	case common.D_Up:
		if !requestsAbove(localElevator, hallRequests) && !hallRequests[localElevator.Floor][common.B_HallUp] {
			if hallRequests[localElevator.Floor][common.B_HallDown] {
				clearHallRequest := make(common.HallRequestType, common.N_FLOORS)
				clearHallRequest[localElevator.Floor][common.B_HallDown] = true
				ClearHallRequest <- clearHallRequest
			}
		}
		if hallRequests[localElevator.Floor][common.B_HallUp] {
			clearHallRequest := make(common.HallRequestType, common.N_FLOORS)
			clearHallRequest[localElevator.Floor][common.B_HallUp] = true
			ClearHallRequest <- clearHallRequest
		}
	case common.D_Down:
		if !requestsBelow(localElevator, hallRequests) && !hallRequests[localElevator.Floor][common.B_HallDown] {
			if hallRequests[localElevator.Floor][common.B_HallUp] {

				clearHallRequest := make(common.HallRequestType, common.N_FLOORS)
				clearHallRequest[localElevator.Floor][common.B_HallUp] = true
				ClearHallRequest <- clearHallRequest
			}
		}
		if hallRequests[localElevator.Floor][common.B_HallDown] {
			clearHallRequest := make(common.HallRequestType, common.N_FLOORS)
			clearHallRequest[localElevator.Floor][common.B_HallDown] = true
			ClearHallRequest <- clearHallRequest
		}
	case common.D_Stop:
		if hallRequests[localElevator.Floor][common.B_HallUp] {
			clearHallRequest := make(common.HallRequestType, common.N_FLOORS)
			clearHallRequest[localElevator.Floor][common.B_HallUp] = true
			ClearHallRequest <- clearHallRequest
		}
		if hallRequests[localElevator.Floor][common.B_HallDown] {
			clearHallRequest := make(common.HallRequestType, common.N_FLOORS)
			clearHallRequest[localElevator.Floor][common.B_HallDown] = true
			ClearHallRequest <- clearHallRequest
		}
	}

	return localElevator, hallRequests
}

// RequestsAbove checks if there are any requests above the current floor
func requestsAbove(
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

// RequestsBelow checks if there are any requests below the current floor
func requestsBelow(
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

// HasRequests checks if the elevator has any requests
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

// RequestsChooseDirection chooses the direction the elevator should move
// based on the current floor and the requests in the system.
// If there are no requests, the elevator should stop.
func requestsChooseDirection(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
) common.Dirn {

	if requestsAbove(localElevator, hallRequests) {
		return common.D_Up
	}

	if requestsBelow(localElevator, hallRequests) {
		return common.D_Down
	}

	return common.D_Stop
}
