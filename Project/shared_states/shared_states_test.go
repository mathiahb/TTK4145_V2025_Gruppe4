package shared_states

import (
	network "elevator_project/network"
	"fmt"
	"testing"
	"time"
	"elevator_project/common"
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

type fakeElevator struct {
	toSharedState   FromElevator
	fromSharedState ToElevator

	get_information chan bool
	close_channel   chan bool

	information_local_hr  chan common.HallRequestType
	information_global_hr chan common.HallRequestType
	information_global_cr chan []bool
}

func (elevator fakeElevator) run() {
	global_cr := make([]bool, 4)
	local_hr := make(common.HallRequestType, 4)
	global_hr := make(common.HallRequestType, 4)

	for {
		select {
		case <-elevator.close_channel:
			return
		case <-elevator.get_information:
			elevator.information_local_hr <- local_hr
			elevator.information_global_hr <- global_hr
			elevator.information_global_cr <- global_cr

		case global_cr = <-elevator.fromSharedState.ApprovedCabRequests:
		case local_hr = <-elevator.fromSharedState.ApprovedHRA:
		case global_hr = <-elevator.fromSharedState.UpdateHallRequestLights:
		}
	}
}

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

		get_information:       make(chan bool),
		close_channel:         make(chan bool),
		information_local_hr:  make(chan common.HallRequestType, 1),
		information_global_hr: make(chan common.HallRequestType, 1),
		information_global_cr: make(chan []bool, 1),
	}

	// Open node
	Node1 := network.New_Node(name1, transferToNetworkChannels(toNetwork, fromNetwork))
	Node1.Connect()
	defer Node1.Close()

	go SharedStatesRoutine(make(chan common.Elevator), toElevator, fromElevator, toNetwork, fromNetwork)
	go fakeElevator.run()
	defer close(fakeElevator.close_channel)

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

	fakeElevator.get_information <- true

	expectedhrresult := common.HallRequestType{{true, false}, {true, false}, {false, false}, {false, false}}
	expectedcrresult := make([]bool, 4)

	expectedhrresultAsString := fmt.Sprintf("%+v", expectedhrresult)
	expectedcrresultAsString := fmt.Sprintf("%+v", expectedcrresult)

	localhr := <-fakeElevator.information_local_hr
	localhrAsString := fmt.Sprintf("%+v", localhr)

	globalhr := <-fakeElevator.information_global_hr
	globalhrAsString := fmt.Sprintf("%+v", globalhr)

	globalcr := <-fakeElevator.information_global_cr
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
