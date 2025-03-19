package network

import (
	"Constants"
	"fmt"
	"time"

	peer_to_peer "github.com/mathiahb/TTK4145_V2025_Gruppe4/Network_Protocol/Network/Peer_to_Peer"
)

// Protocol SYNCHRONIZATION
// ---
// Event that triggers Synchronization ->
// Coordinator sends a SYNC_AFTER_DISCOVERY message with it's own payload.
// Every voter responds with

// Channels to speak with whatever manages the Synchronization interpreting and sharing
/*
Expected usage if WE initiated the synchronization:
ProtocolRequestInformation <- true
-- Receive information from the other side via ResponToInformationRequest
Waits for everyone else
ProtocolRequestInterpretation <- responses (In the form of map[InformationOwner]Information)
-- Receive the correct interpretation from the other side via RespondWithInterpretation
Waits for result
ResultFromSynchronization <- result (should be the same as the interpretation.)
///////////////////////////////////////////////////////////////////////////////////////
Expected usage if SOMEONE ELSE initiated the synchronization:
ProtocolRequestInformation <-true
-- Receive information from the other side via RespondToInformationRequest
Wait for protocol result
ResultFromSynchronization <- result
*/
type SynchronizationChannels struct {
	ProtocolRequestInformation  chan bool
	RespondToInformationRequest chan string

	ProtocolRequestsInterpretation chan map[string]string
	RespondWithInterpretation      chan string

	ResultFromSynchronization chan string
}

func New_SynchronizationChannels() SynchronizationChannels {
	return SynchronizationChannels{
		ProtocolRequestInformation:  make(chan bool),
		RespondToInformationRequest: make(chan string),

		ProtocolRequestsInterpretation: make(chan map[string]string),
		RespondWithInterpretation:      make(chan string),

		ResultFromSynchronization: make(chan string),
	}
}

func (node *Node) get_Synchronization_Information() string {
	node.synchronization_channels.ProtocolRequestInformation <- true
	return <-node.synchronization_channels.RespondToInformationRequest
}

func (node *Node) interpret_Synchronization_Responses(responses map[string]string) string {
	node.synchronization_channels.ProtocolRequestsInterpretation <- responses
	return <-node.synchronization_channels.RespondWithInterpretation
}

func (node *Node) coordinate_Synchronization(success_channel chan bool) {
	node.mu_voting_resource.Lock()
	defer node.mu_voting_resource.Unlock()

	begin_synchronization_message := node.create_Vote_Message(Constants.SYNC_AFTER_DISCOVERY, "")
	node.Broadcast(begin_synchronization_message)

	combined_information := make(map[string]string)
	combined_information[node.name] = node.get_Synchronization_Information()

	amount_of_info_needed := len(node.Get_Alive_Nodes())
	timeout := time.After(time.Second)
	for {
		select {
		case response := <-node.comm:
			if response.message_type == Constants.SYNC_RESPONSE && response.id == begin_synchronization_message.id {
				combined_information[response.sender] = response.payload

				if len(combined_information) == amount_of_info_needed {
					result := node.interpret_Synchronization_Responses(combined_information)

					node.broadcast_Synchronization_Result(begin_synchronization_message.id, result)
					node.synchronization_channels.ResultFromSynchronization <- result

					success_channel <- true
					return
				}
			}
			if response.message_type == Constants.ABORT_COMMIT && response.id == begin_synchronization_message.id {
				node.abort_Synchronization(begin_synchronization_message.id)

				success_channel <- false
				return
			}
		case <-timeout:
			node.abort_Synchronization(begin_synchronization_message.id)
			node.protocol_timed_out()

			success_channel <- false
			return
		}
	}
}

func (node *Node) broadcast_Synchronization_Result(id TxID, result string) {
	message := node.create_Message(Constants.SYNC_RESULT, id, result)
	node.Broadcast(message)
}

func (node *Node) abort_Synchronization(id_discovery TxID) {
	message := node.create_Message(Constants.ABORT_COMMIT, id_discovery, "")
	node.Broadcast(message)
}

func (node *Node) participate_In_Synchronization(p2p_message peer_to_peer.P2P_Message, id_discovery TxID) {
	if node.isTxIDFromUs(id_discovery) {
		return
	}

	ok := node.mu_voting_resource.TryLock()
	if !ok {
		node.abort_Discovery(id_discovery)
		return
	}
	defer node.mu_voting_resource.Unlock()

	node.synchronization_channels.ProtocolRequestInformation <- true
	information := <-node.synchronization_channels.RespondToInformationRequest

	response := node.create_Message(Constants.SYNC_RESPONSE, id_discovery, information)
	node.Broadcast_Response(response, p2p_message)

	timeout := time.After(time.Second)

	for {
		select {
		case result := <-node.comm:
			if result.message_type == Constants.SYNC_RESULT && result.id == id_discovery {
				node.synchronization_channels.ResultFromSynchronization <- result.payload
				return
			}

			if result.message_type == Constants.ABORT_COMMIT && result.id == id_discovery {
				return
			}
		case <-timeout:
			fmt.Printf("[ERROR %s]: Synchronization %s halted in progress!\n", node.name, id_discovery)
			return
		}
	}
}
