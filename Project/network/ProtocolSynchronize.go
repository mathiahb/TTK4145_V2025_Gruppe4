package network

import (
	Constants "elevator_project/constants"
	"fmt"
	"time"

	peer_to_peer "elevator_project/network/Peer_to_Peer"
)

// PROTOCOL - Synchronize
/*
	Message				:	CODE	PAYLOAD
	---
	Synchronize Begin	: 	SYNC
	Synchronize Response:	SRSP	INFO_TO_SYNCHRONIZE
	Synchronize Result	:	SRST	INFO_UPDATED
	Abort Commit		:	ERRC
	---

	Expected procedure
	---
	Coordinator: Synchronize dispatched -> go coordinate_Synchronization(success_channel)
	Coordinator: Broadcast Synchronize Begin
	Coordinator: Acquires it's own information via the synchronization channels.

	Participants: Other nodes receives Synchronize Begin -> go participate_Synchronization()
	Participants: Acquires their information via the synchronization channels.
	Participants: Every node responds Synchronize Response OR Abort Commit

	Coordinator: Receives responses
	If Abort Commit OR Timeout passes:
		Coordinator: Broadcast Abort Commit
		Coordinator: Returns false on the success_channel

	If All nodes deliver information:
		Coordinator: Compiles results and requests result from the synchronization channels
		Coordinator: Broadcast Synchronize Result
		Coordinator: Passes the result to synchronizechannels.ResultFromSynchronization
		Coordinator: Returns true on the success_channel
		Participants: Receives Synchronize Result
		Participants: Passes the result to synchronizechannels.ResultFromSynchronization
*/

func (node *Node) get_Synchronization_Information() string {
	node.shared_state_communication.FromNetwork.Synchronization.ProtocolRequestInformation <- true
	return <-node.shared_state_communication.ToNetwork.Synchronization.RespondToInformationRequest
}

func (node *Node) interpret_Synchronization_Responses(responses map[string]string) string {
	node.shared_state_communication.FromNetwork.Synchronization.ProtocolRequestsInterpretation <- responses
	return <-node.shared_state_communication.ToNetwork.Synchronization.RespondWithInterpretation
}

func (node *Node) send_Synchronization_Result(result string) {
	node.shared_state_communication.FromNetwork.Synchronization.ResultFromSynchronization <- result
}

func (node *Node) coordinate_Synchronization(success_channel chan bool, begin_discovery_message Message) {
	//.mu_voting_resource.Lock()
	//defer node.mu_voting_resource.Unlock()

	//begin_synchronization_message := node.create_Message(Constants.SYNC_AFTER_DISCOVERY, begin_discovery_message.id, "")
	//node.Broadcast(begin_synchronization_message)

	own_information := node.get_Synchronization_Information()
	node.Broadcast(node.create_Message(Constants.SYNC_RESPONSE, begin_discovery_message.id, own_information))

	combined_information := make(map[string]string)
	amount_of_info_needed := len(node.Get_Alive_Nodes())
	timeout := time.After(time.Second)
	for {
		select {
		case response := <-node.comm:
			if response.message_type == Constants.SYNC_RESPONSE && response.id == begin_discovery_message.id {
				combined_information[response.sender] = response.payload

				if len(combined_information) == amount_of_info_needed {
					result := node.interpret_Synchronization_Responses(combined_information)

					node.broadcast_Synchronization_Result(begin_discovery_message.id, result)
					node.send_Synchronization_Result(result)

					success_channel <- true
					return
				}
			}
			if response.message_type == Constants.ABORT_COMMIT && response.id == begin_discovery_message.id {
				node.abort_Synchronization(begin_discovery_message.id)

				success_channel <- false
				return
			}
		case <-timeout:
			node.abort_Synchronization(begin_discovery_message.id)
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

	//ok := node.mu_voting_resource.TryLock()
	//if !ok {
	//	node.abort_Discovery(id_discovery)
	//	return
	//}
	//defer node.mu_voting_resource.Unlock()

	information := node.get_Synchronization_Information()

	response := node.create_Message(Constants.SYNC_RESPONSE, id_discovery, information)
	node.Broadcast_Response(response, p2p_message)

	timeout := time.After(time.Second)

	for {
		select {
		case result := <-node.comm:
			if result.message_type == Constants.SYNC_RESULT && result.id == id_discovery {
				node.send_Synchronization_Result(result.payload)
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
