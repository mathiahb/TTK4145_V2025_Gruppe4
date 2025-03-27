package shared_states

import (
	"elevator_project/constants"
)

type ToElevator struct {
	UpdateHallRequestLights chan constants.HallRequestType
	ApprovedCabRequests     chan []bool
	ApprovedHRA             chan constants.HallRequestType
}

type FromElevator struct {
	NewHallRequest   chan constants.HallRequestType
	ClearHallRequest chan constants.HallRequestType
	UpdateState      chan constants.Elevator
}

type ToNetwork struct {
	Inform2PC                   chan string
	RespondWithInterpretation   chan string
	RespondToInformationRequest chan string
}

type FromNetwork struct {
	NewAliveNodes                  chan []string
	ApprovedBy2PC                  chan string
	ProtocolRequestInformation     chan bool
	ProtocolRequestsInterpretation chan map[string]string
	ResultFromSynchronization      chan string
}
