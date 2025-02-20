package elevator

import (
	"Go-driver/elevator/elevio"
)

// DirnBehaviourPair strukturen holder retningen og oppførselen for heisen
type DirnBehaviourPair struct {
	Dirn      Dirn
	Behaviour ElevatorBehaviour
}

// requests_above sjekker om det finnes en bestilling over gjeldende etasje
func requestsAbove(e Elevator) bool {
	for f := e.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] != 0 {
				return true
			}
		}
	}
	return false
}

// requests_below sjekker om det finnes en bestilling under gjeldende etasje
func requestsBelow(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] != 0 {
				return true
			}
		}
	}
	return false
}

// requestsHere sjekker om det finnes en bestilling på nåværende etasje
func requestsHere(e Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] != 0 {
			return true
		}
	}
	return false
}

// requestsChooseDirection velger hvilken retning heisen skal gå i
func requestsChooseDirection(e Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case D_Up:
		// NYTT TESTING
		if e.Floor == N_FLOORS-1 { // Hvis heisen er på toppen, skal den ikke gå opp
			if requestsHere(e) {
				return DirnBehaviourPair{D_Stop, EB_DoorOpen}
			} else if requestsBelow(e) {
				return DirnBehaviourPair{D_Down, EB_Moving}
			}
			return DirnBehaviourPair{D_Stop, EB_Idle}
		}
		// NYTT TESTING
		if requestsAbove(e) {
			return DirnBehaviourPair{D_Up, EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{D_Down, EB_DoorOpen}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{D_Down, EB_Moving}
		}
		return DirnBehaviourPair{D_Stop, EB_Idle}

	case D_Down:
		if requestsBelow(e) {
			return DirnBehaviourPair{D_Down, EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{D_Up, EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{D_Up, EB_Moving}
		}
		return DirnBehaviourPair{D_Stop, EB_Idle}

	case D_Stop:
		if requestsHere(e) {
			return DirnBehaviourPair{D_Stop, EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{D_Up, EB_Moving}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{D_Down, EB_Moving}
		}
		return DirnBehaviourPair{D_Stop, EB_Idle}

	default:
		return DirnBehaviourPair{D_Stop, EB_Idle}
	}
}

// requestsShouldStop sjekker om heisen skal stoppe på nåværende etasje
func requestsShouldStop(e Elevator) bool {
	switch e.Dirn {
	case D_Down:
		return e.Requests[e.Floor][B_HallDown] != 0 ||
			e.Requests[e.Floor][B_Cab] != 0 ||
			!requestsBelow(e)

	case D_Up:
		return e.Requests[e.Floor][B_HallUp] != 0 ||
			e.Requests[e.Floor][B_Cab] != 0 ||
			!requestsAbove(e)

	case D_Stop:
		return true
	default:
		return false
	}
}

// requestsShouldClearImmediately sjekker om en bestilling skal behandles og slettes umiddelbart.
func requestsShouldClearImmediately(e Elevator, btnFloor int, btnType elevio.ButtonType) bool {
	switch e.Config.ClearRequestVariant {
	case CV_All:
		return e.Floor == btnFloor

	case CV_InDirn:
		return e.Floor == btnFloor && ((e.Dirn == D_Up && btnType == B_HallUp) ||
			(e.Dirn == D_Down && btnType == B_HallDown) ||
			e.Dirn == D_Stop || btnType == B_Cab)

	default:
		return false
	}
}

// requestsClearAtCurrentFloor rydder bestillinger på gjeldende etasje
func requestsClearAtCurrentFloor(e Elevator) Elevator {
	switch e.Config.ClearRequestVariant {
	case CV_All:
		for btn := 0; btn < N_BUTTONS; btn++ {
			e.Requests[e.Floor][btn] = 0
		}

	case CV_InDirn:
		e.Requests[e.Floor][B_Cab] = 0
		switch e.Dirn {
		case D_Up:
			if !requestsAbove(e) && e.Requests[e.Floor][B_HallUp] == 0 {
				e.Requests[e.Floor][B_HallDown] = 0
			}
			e.Requests[e.Floor][B_HallUp] = 0

		case D_Down:
			if !requestsBelow(e) && e.Requests[e.Floor][B_HallDown] == 0 {
				e.Requests[e.Floor][B_HallUp] = 0
			}
			e.Requests[e.Floor][B_HallDown] = 0

		case D_Stop:
		default:
			e.Requests[e.Floor][B_HallUp] = 0
			e.Requests[e.Floor][B_HallDown] = 0
		}
	}

	return e
}
