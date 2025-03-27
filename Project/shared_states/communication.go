package shared_states
import(
	"elevator_project/common"
)

type ToElevator struct {
	UpdateHallRequestLights chan common.HallRequestType
	ApprovedCabRequests     chan []bool
	ApprovedHRA             chan common.HallRequestType
}

type FromElevator struct {
	NewHallRequest   chan common.HallRequestType
	ClearHallRequest chan common.HallRequestType
	UpdateState      chan common.Elevator
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
