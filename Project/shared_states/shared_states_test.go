package shared_states

import (
	"elevator_project/common"
	network "elevator_project/network"
	"fmt"
	"testing"
	"time"
)

func transferToNetworkChannels(toNetwork ToNetwork, fromNetwork FromNetwork) network.NetworkCommunicationChannels {
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

// fakeElevator simulates an elevator's interaction with shared states.
type fakeElevator struct {
	toSharedState   FromElevator
	fromSharedState ToElevator

	getInformation chan bool
	closeChannel   chan bool

	informationLocalHr  chan common.HallRequestType
	informationGlobalHr chan common.HallRequestType
	informationGlobalCr chan []bool
}

// run handles the fake elevator's state updates and communication with shared states.
func (elevator fakeElevator) run() {
	globalCr := make([]bool, 4)
	localHr := make(common.HallRequestType, 4)
	globalHr := make(common.HallRequestType, 4)

	for {
		select {
		case <-elevator.closeChannel:
			return
		case <-elevator.getInformation:
			elevator.informationLocalHr <- localHr
			elevator.informationGlobalHr <- globalHr
			elevator.informationGlobalCr <- globalCr

		case globalCr = <-elevator.fromSharedState.ApprovedCabRequests:
		case localHr = <-elevator.fromSharedState.ApprovedHRA:
		case globalHr = <-elevator.fromSharedState.UpdateHallRequestLights:
		}
	}
}

// TestSharedStateUpdate tests the shared state update mechanism by simulating elevator requests and verifying results.
func TestSharedStateUpdate(t *testing.T) {
	name1 := common.GetElevatorID()

	toNetwork := ToNetwork{
		RespondWithInterpretation:   make(chan string, 1),
		RespondToInformationRequest: make(chan string, 1),

		Inform2PC: make(chan string, 1),
	}

	fromNetwork := FromNetwork{
		NewAliveNodes: make(chan []string, 1),

		ProtocolRequestInformation:     make(chan bool, 1),
		ProtocolRequestsInterpretation: make(chan map[string]string, 1),
		ResultFromSynchronization:      make(chan string, 1),

		ApprovedBy2PC: make(chan string, 1),
	}

	toElevator := ToElevator{
		UpdateHallRequestLights: make(chan common.HallRequestType, 1),
		ApprovedCabRequests:     make(chan []bool, 1),
		ApprovedHRA:             make(chan common.HallRequestType, 1),
	}

	fromElevator := FromElevator{
		NewHallRequest:   make(chan common.HallRequestType, 1),
		ClearHallRequest: make(chan common.HallRequestType, 1),
		UpdateState:      make(chan common.Elevator, 1),
	}

	fakeElevator := fakeElevator{
		fromSharedState: toElevator,
		toSharedState:   fromElevator,

		getInformation:      make(chan bool),
		closeChannel:        make(chan bool),
		informationLocalHr:  make(chan common.HallRequestType, 1),
		informationGlobalHr: make(chan common.HallRequestType, 1),
		informationGlobalCr: make(chan []bool, 1),
	}

	// Open node
	Node1 := network.NewNode(name1, transferToNetworkChannels(toNetwork, fromNetwork))
	Node1.Connect()
	defer Node1.Close()

	go SharedStatesRoutine(make(chan common.Elevator), toElevator, fromElevator, toNetwork, fromNetwork)
	go fakeElevator.run()
	defer close(fakeElevator.closeChannel)

	elevator := common.Elevator{
		Behaviour:   common.EB_Idle,
		Dirn:        common.D_Stop,
		Floor:       2,
		CabRequests: make([]bool, 4),
	}

	fakeElevator.toSharedState.UpdateState <- elevator

	floor1up := common.HallRequestType{{true, false}, {false, false}, {false, false}, {false, false}}
	floor1dn := common.HallRequestType{{false, true}, {false, false}, {false, false}, {false, false}}
	floor2up := common.HallRequestType{{false, false}, {true, false}, {false, false}, {false, false}}
	//floor2dn := common.HallRequestType{{false, false}, {false, true}, {false, false}, {false, false}}
	//floor3up := common.HallRequestType{{false, false}, {false, false}, {true, false}, {false, false}}
	//floor3dn := common.HallRequestType{{false, false}, {false, false}, {false, true}, {false, false}}
	//floor4up := common.HallRequestType{{false, false}, {false, false}, {false, false}, {true, false}}
	//floor4dn := common.HallRequestType{{false, false}, {false, false}, {false, false}, {false, true}}

	fakeElevator.toSharedState.NewHallRequest <- floor1up
	fakeElevator.toSharedState.NewHallRequest <- floor1dn
	fakeElevator.toSharedState.NewHallRequest <- floor2up
	fakeElevator.toSharedState.ClearHallRequest <- floor1dn

	time.Sleep(time.Millisecond * 150)

	fakeElevator.getInformation <- true

	expectedhrresult := common.HallRequestType{{true, false}, {true, false}, {false, false}, {false, false}}
	expectedcrresult := make([]bool, 4)

	expectedhrresultAsString := fmt.Sprintf("%+v", expectedhrresult)
	expectedcrresultAsString := fmt.Sprintf("%+v", expectedcrresult)

	localhr := <-fakeElevator.informationLocalHr
	localhrAsString := fmt.Sprintf("%+v", localhr)

	globalhr := <-fakeElevator.informationGlobalHr
	globalhrAsString := fmt.Sprintf("%+v", globalhr)

	globalcr := <-fakeElevator.informationGlobalCr
	globalcrAsString := fmt.Sprintf("%+v", globalcr)

	if expectedhrresultAsString != localhrAsString {
		t.Errorf("Error Local Hall Requests. %s != %s", expectedhrresultAsString, localhrAsString)
	}

	if expectedhrresultAsString != globalhrAsString {
		t.Errorf("Error Global Hall Requests. %s != %s", expectedhrresultAsString, globalhrAsString)
	}

	if expectedcrresultAsString != globalcrAsString {
		t.Errorf("Error Global Cab Requests. %s != %s", expectedcrresultAsString, globalcrAsString)
	}
}
