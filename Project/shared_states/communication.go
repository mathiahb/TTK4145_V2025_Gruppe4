package shared_states

import (
	. "elevator_project/constants"
)

// lager bare kanalene fra shared state til elevator
// lager kanalene fra shared state til elevator
// ergo de kanalene fra nettverket som går til shared state skal defineres på nettverksbiten

type ToElevator struct {
	UpdateHallRequestLights    chan HallRequestType
	ApprovedCabRequestsChannel chan []bool
	ApprovedHRAChannel         chan HallRequestType
}

type FromElevator struct {
	NewHallRequestChannel   chan HallRequestType
	ClearHallRequestChannel chan HallRequestType
	UpdateState             chan Elevator
}

type ToNetwork struct {
	Inform2PC                   chan string
	RespondWithInterpretation   chan string
	RespondToInformationRequest chan string
}

type FromNetwork struct {
	New_alive_nodes                chan []string
	ApprovedBy2PC                  chan string
	ProtocolRequestInformation     chan bool
	ProtocolRequestsInterpretation chan map[string]string
	ResultFromSynchronization      chan string
}

/*
Må få samsvar mellom shared state og nettverket

idé: lagen egen pakke for kanaler og tråder som shared_state skal kommunisere med communication_channels eller module_communication
for er litt redd for circular dependency

synchronizationChannels
------------------
ProtocolRequestInformation  chan bool <-> startSynchChannel
RespondToInformationRequest chan string <-> sendStateForSynchChannel

ProtocolRequestsInterpretation chan map[string]string // kart av nodenavn til sharedstaten på noden
RespondWithInterpretation      chan string // her må jeg returnere tilbake hva det betyr

ResultFromSynchronization chan string <-> updatedSharedStateForSynchChannel // ny felles shared state, HRAInputVariable



twoPhaseCommitChannels
----------------------
det jeg har:
-> Fra shared state
-- notifyNewHallRequestChannel
-- informNewStateChannel

-> Til shared state
-- approvedHallRequestChannel
-- approvedNewelevatorStateChannel:

(hvor cab requests sendes på nettverket som en new state)

requestCommit(chan string)

commitApproved(chan string)


discovery
----------------------
alive nodes // henger ikke med på hvorfor dette er nødvendig
*/
