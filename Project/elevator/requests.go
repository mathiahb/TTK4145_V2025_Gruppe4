package elevator

import(
	"elevator_project/common"
)

func HallRequestsUninitialized() common.HallRequestType {
	hallRequests := make(common.HallRequestType, common.N_FLOORS)
	return hallRequests
}

// requestsShouldStop sjekker om heisen skal stoppe på nåværende etasje
func requestsShouldStop(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
) bool {

	// 1. Stopp alltid hvis en cab-request er på denne etasjen
	if localElevator.CabRequests[localElevator.Floor] {
		return true
	}

	// 2. Stopp hvis hall-forespørsel er tildelt denne heisen i denne etasjen
	if localElevator.Dirn == common.D_Up && hallRequests[localElevator.Floor][common.B_HallUp] {
		return true
	}
	if localElevator.Dirn == common.D_Down && hallRequests[localElevator.Floor][common.B_HallDown] {
		return true
	}

	// 3. Hvis det ikke er flere hall requests i denne retningen, stopp
	if localElevator.Dirn == common.D_Up {
		return !requests_above(localElevator, hallRequests)
	}

	if localElevator.Dirn == common.D_Down {
		return !requests_below(localElevator, hallRequests)
	}

	return false
}

// requestsClearAtCurrentFloor rydder bestillinger på gjeldende etasje
// localElevator, hallRequests, ClearHallRequest, updateStateChannel
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

// requestsChooseDirection velger retning basert på forespørsler
func requestsChooseDirection(
	localElevator common.Elevator,
	hallRequests common.HallRequestType,
) common.Dirn {

	// Hvis det er noen forespørsler over heisen
	for f := localElevator.Floor + 1; f < common.N_FLOORS; f++ {
		if hallRequests[f][common.B_HallUp] || hallRequests[f][common.B_HallDown] || localElevator.CabRequests[f] {
			return common.D_Up
		}
	}

	// Hvis det er noen forespørsler under heisen
	for f := 0; f < localElevator.Floor; f++ {
		if hallRequests[f][common.B_HallUp] || hallRequests[f][common.B_HallDown] || localElevator.CabRequests[f] {
			return common.D_Down
		}
	}
	// Ingen forespørsler under eller over, stopp
	return common.D_Stop
}
