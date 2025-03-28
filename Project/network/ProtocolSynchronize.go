package network

import (
	"elevator_project/common"
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
	Coordinator: Synchronize dispatched -> go coordinateSynchronization(success_channel)
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
	node.sharedStateCommunication.FromNetwork.Synchronization.ProtocolRequestInformation <- true
	return <-node.sharedStateCommunication.ToNetwork.Synchronization.RespondToInformationRequest
}

func (node *Node) interpret_Synchronization_Responses(responses map[string]string) string {
	node.sharedStateCommunication.FromNetwork.Synchronization.ProtocolRequestsInterpretation <- responses
	return <-node.sharedStateCommunication.ToNetwork.Synchronization.RespondWithInterpretation
}

func (node *Node) sendSynchronizationResult(result string) {
	node.sharedStateCommunication.FromNetwork.Synchronization.ResultFromSynchronization <- result
}

func (node *Node) coordinateSynchronization() bool {
	if !node.connectedToNetwork {
		return true // We're not connected, no need to do anything.
	}

	ok := node.muVotingResource.TryLock()
	if !ok {
		return false
	}
	defer node.muVotingResource.Unlock()

	//beginSynchronizationMessage := node.createMessage(common.SYNC_AFTER_DISCOVERY, begin_discovery_message.id, "")
	//node.Broadcast(beginSynchronizationMessage)

	beginSynchronizationMessage := node.createVoteMessage(common.SYNC_REQUEST, "")

	comm := node.createCommunicationChannel(beginSynchronizationMessage)
	defer node.deleteCommunicationChannel(beginSynchronizationMessage)

	node.Broadcast(beginSynchronizationMessage)
	node.sendOwnInformationForSynchronization(beginSynchronizationMessage)

	combined_information := make(map[string]string)
	timeout := time.After(time.Second)
	for {
		select {
		case response := <-comm:
			if !node.aliveNodesManager.IsNodeAlive(response.sender) {
				node.abortSynchronization(beginSynchronizationMessage)

				node.Connect()
				return false
			}

			if response.messageType == common.SYNC_RESPONSE && response.id == beginSynchronizationMessage.id {
				combined_information[response.sender] = response.payload

				if len(combined_information) == len(node.aliveNodesManager.GetAliveNodes()) {
					result := node.interpret_Synchronization_Responses(combined_information)

					go node.broadcast_Synchronization_Result(beginSynchronizationMessage, result)
					go node.sendSynchronizationResult(result)

					return true
				}
			}
			if response.messageType == common.ABORT_SYNCHRONIZATION && response.id == beginSynchronizationMessage.id {
				node.abortSynchronization(beginSynchronizationMessage)

				return false
			}
		case <-timeout:
			node.abortSynchronization(beginSynchronizationMessage)
			node.protocolTimedOut()

			return false
		}
	}
}

func (node *Node) broadcast_Synchronization_Result(beginSynchronizationMessage Message, result string) {
	message := node.createMessage(common.SYNC_RESULT, beginSynchronizationMessage.id, result)
	node.BroadcastResponse(message, beginSynchronizationMessage)
}

func (node *Node) abortSynchronization(beginSynchronizationMessage Message) {
	message := node.createMessage(common.ABORT_SYNCHRONIZATION, beginSynchronizationMessage.id, "")
	node.BroadcastResponse(message, beginSynchronizationMessage)
}

func (node *Node) sendOwnInformationForSynchronization(synchronization_message Message) {
	information := node.get_Synchronization_Information()

	response := node.createMessage(common.SYNC_RESPONSE, synchronization_message.id, information)
	node.Broadcast(response)
}

func (node *Node) participateInSynchronization(beginMessage Message) {
	if node.isTxIDFromUs(beginMessage.id) {
		return
	}

	ok := node.muVotingResource.TryLock()
	if !ok {
		node.abortSynchronization(beginMessage)
		return
	}
	defer node.muVotingResource.Unlock()

	if !node.aliveNodesManager.IsNodeAlive(node.name) {
		// Node is dead, so we should reconnect
		node.Connect()
		return
	}

	comm := node.createCommunicationChannel(beginMessage)
	defer node.deleteCommunicationChannel(beginMessage)

	node.sendOwnInformationForSynchronization(beginMessage)

	timeout := time.After(time.Second)

	for {
		select {
		case result := <-comm:
			if result.messageType == common.SYNC_RESULT && result.id == beginMessage.id {
				node.sendSynchronizationResult(result.payload)
				return
			}

			if result.messageType == common.ABORT_SYNCHRONIZATION && result.id == beginMessage.id {
				return
			}
		case <-timeout:
			fmt.Printf("[ERROR %s]: Synchronization %s halted in progress!\n", node.name, beginMessage.id)
			return
		}
	}
}
