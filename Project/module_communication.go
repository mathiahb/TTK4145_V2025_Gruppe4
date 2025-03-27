package main

import (
	"elevator_project/network"
	"elevator_project/shared_states"
	"elevator_project/common"
)

func newFromSharedStateToNetwork() shared_states.ToNetwork {
	return shared_states.ToNetwork{
		Inform2PC: make(chan string),

		RespondWithInterpretation:   make(chan string),
		RespondToInformationRequest: make(chan string),
	}
}

func newToSharedStateFromNetwork() shared_states.FromNetwork {
	return shared_states.FromNetwork{
		NewAliveNodes: make(chan []string),

		ProtocolRequestInformation:     make(chan bool),
		ProtocolRequestsInterpretation: make(chan map[string]string),
		ResultFromSynchronization:      make(chan string),

		ApprovedBy2PC: make(chan string),
	}
}

func newFromSharedStateToElevator() shared_states.ToElevator {
	return shared_states.ToElevator{
		UpdateHallRequestLights: make(chan common.HallRequestType),
		ApprovedCabRequests:     make(chan []bool),
		ApprovedHRA:             make(chan common.HallRequestType),
	}
}

func newToSharedStateFromElevator() shared_states.FromElevator {
	return shared_states.FromElevator{
		NewHallRequest:   make(chan common.HallRequestType),
		ClearHallRequest: make(chan common.HallRequestType),
		UpdateState:      make(chan common.Elevator),
	}
}

func transferSharedStateChannelsToNetworkChannels(toNetwork shared_states.ToNetwork, fromNetwork shared_states.FromNetwork) network.NetworkCommunicationChannels {
	return network.NetworkCommunicationChannels{
		ToNetwork: network.CommunicationToNetwork{
			Discovery: struct{}{},
			Synchronization: struct {
				RespondToInformationRequest chan string
				RespondWithInterpretation   chan string
			}{
				RespondToInformationRequest: toNetwork.RespondToInformationRequest,
				RespondWithInterpretation:   toNetwork.RespondWithInterpretation,
			},
			TwoPhaseCommit: struct{ RequestCommit chan string }{
				RequestCommit: toNetwork.Inform2PC,
			},
		},
		FromNetwork: network.CommunicationFromNetwork{
			Discovery: struct{ Updated_Alive_Nodes chan []string }{
				Updated_Alive_Nodes: fromNetwork.NewAliveNodes,
			},
			Synchronization: struct {
				ProtocolRequestInformation     chan bool
				ProtocolRequestsInterpretation chan map[string]string
				ResultFromSynchronization      chan string
			}{
				ProtocolRequestInformation:     fromNetwork.ProtocolRequestInformation,
				ProtocolRequestsInterpretation: fromNetwork.ProtocolRequestsInterpretation,
				ResultFromSynchronization:      fromNetwork.ResultFromSynchronization,
			},
			TwoPhaseCommit: struct{ ProtocolCommited chan string }{
				ProtocolCommited: fromNetwork.ApprovedBy2PC,
			},
		},
	}
}
