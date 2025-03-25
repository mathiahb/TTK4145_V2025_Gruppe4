package network

import (
	Constants "elevator_project/constants"
	"fmt"
	"time"
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

func (node *Node) coordinate_Synchronization() bool {
	node.mu_voting_resource.Lock()
	defer node.mu_voting_resource.Unlock()

	//begin_synchronization_message := node.create_Message(Constants.SYNC_AFTER_DISCOVERY, begin_discovery_message.id, "")
	//node.Broadcast(begin_synchronization_message)

	begin_synchronization_message := node.create_Vote_Message(Constants.SYNC_REQUEST, "")

	comm := node.create_communication_channel(begin_synchronization_message)
	defer node.delete_communication_channel(begin_synchronization_message)

	node.Broadcast(begin_synchronization_message)
	node.send_Own_Information_For_Synchronization(begin_synchronization_message)

	combined_information := make(map[string]string)
	timeout := time.After(time.Second)
	for {
		select {
		case response := <-comm:
			if !node.alive_nodes_manager.Is_Node_Alive(response.sender) {
				node.abort_Synchronization(begin_synchronization_message)

				node.Connect()
				return false
			}

			if response.message_type == Constants.SYNC_RESPONSE && response.id == begin_synchronization_message.id {
				combined_information[response.sender] = response.payload

				if len(combined_information) == len(node.alive_nodes_manager.Get_Alive_Nodes()) {
					result := node.interpret_Synchronization_Responses(combined_information)

					go node.broadcast_Synchronization_Result(begin_synchronization_message, result)
					go node.send_Synchronization_Result(result)

					return true
				}
			}
			if response.message_type == Constants.ABORT_COMMIT && response.id == begin_synchronization_message.id {
				node.abort_Synchronization(begin_synchronization_message)

				return false
			}
		case <-timeout:
			node.abort_Synchronization(begin_synchronization_message)
			node.protocol_timed_out()

			return false
		}
	}
}

func (node *Node) broadcast_Synchronization_Result(begin_synchronization_message Message, result string) {
	message := node.create_Message(Constants.SYNC_RESULT, begin_synchronization_message.id, result)
	node.Broadcast_Response(message, begin_synchronization_message)
}

func (node *Node) abort_Synchronization(begin_synchronization_message Message) {
	message := node.create_Message(Constants.ABORT_SYNCHRONIZATION, begin_synchronization_message.id, "")
	node.Broadcast_Response(message, begin_synchronization_message)
}

func (node *Node) send_Own_Information_For_Synchronization(synchronization_message Message) {
	information := node.get_Synchronization_Information()

	response := node.create_Message(Constants.SYNC_RESPONSE, synchronization_message.id, information)
	node.Broadcast(response)
}

func (node *Node) participate_In_Synchronization(begin_message Message) {
	if node.isTxIDFromUs(begin_message.id) {
		return
	}

	//node.mu_voting_resource.Lock()
	//defer node.mu_voting_resource.Unlock()

	if !node.alive_nodes_manager.Is_Node_Alive(node.name) {
		// Node is dead, so we should reconnect
		node.Connect()
		return
	}

	comm := node.create_communication_channel(begin_message)
	defer node.delete_communication_channel(begin_message)

	node.send_Own_Information_For_Synchronization(begin_message)

	timeout := time.After(time.Second)

	for {
		select {
		case result := <-comm:
			if result.message_type == Constants.SYNC_RESULT && result.id == begin_message.id {
				node.send_Synchronization_Result(result.payload)
				return
			}

			if result.message_type == Constants.ABORT_COMMIT && result.id == begin_message.id {
				return
			}
		case <-timeout:
			fmt.Printf("[ERROR %s]: Synchronization %s halted in progress!\n", node.name, begin_message.id)
			return
		}
	}
}
