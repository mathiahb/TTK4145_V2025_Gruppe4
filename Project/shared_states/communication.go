package shared_states

import( 
	. "elevator_project/constants"

)

// lager bare kanalene fra shared state til elevator 
// lager kanalene fra shared state til elevator
// ergo de kanalene fra nettverket som går til shared state skal defineres på nettverksbiten

type BetweenElevatorAndSharedStatesChannels struct {
	HallRequestChannel               chan HRAType
	ElevatorStateChannel             chan Elevator
	ClearCabRequestChannel           chan Elevator
	ClearHallRequestChannel          chan HallRequestType
	ApprovedClearHallRequestsChannel chan HallRequestType
	NewHallRequestChannel            chan HallRequestType
	ApprovedHallRequestChannel       chan HallRequestType
}

func MakeBetweenElevatorAndSharedStatesChannels() BetweenElevatorAndSharedStatesChannels {
	return BetweenElevatorAndSharedStatesChannels{
		HallRequestChannel:               make(chan HRAType),  // fra shared state til elevator
		ElevatorStateChannel:             make(chan Elevator), // fra elevator til shared states
		ClearCabRequestChannel:           make(chan Elevator),
		ClearHallRequestChannel:          make(chan HallRequestType),
		ApprovedClearHallRequestsChannel: make(chan HallRequestType),
		NewHallRequestChannel:            make(chan HallRequestType), // fra elevator til shared states, sender ny HallRequest, når knapp trykket inn
		ApprovedHallRequestChannel:       make(chan HallRequestType), // fra shared state til elevator, sender godkjent HallRequest etter konferering med nettverket

	}
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