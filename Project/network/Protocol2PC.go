package network

import (
	"elevator_project/common"
	"fmt"
	"time"
)

// PROTOCOL - 2PC
/*
	Message				:	CODE	PAYLOAD
	---
	PREPARE				:	PREP
	PREPARE_ACK			:	PREA
	COMMIT				:	COMT
	Abort Commit		:	ERRC
	---

	Expected procedure
	---
	Coordinator: 2PC dispatched -> go node.coordinate2PC(success_channel)
	Coordinator: Broadcast PREPARE
	Coordinator: Waits for responses

	Participants: Other nodes receives PREPARE -> go node.participate2PC()
	Participants: Every node responds PREPARE_ACK OR Abort Commit

	Coordinator: Receives responses
	If Abort Commit:
		Coordinator: Broadcast Abort Commit
		Coordinator: Returns false on the success_channel
	If all PREPARE_ACK received:
		Coordinator: Broadcast COMMIT
		Coordinator: Commits the change locally
		Coordinator: Returns true on the success_channel

	Timeout passes:
		Coordinator: Compiles results
		Coordinator: Broadcast Abort Commit
		Coordinator: Returns false on the success_channel


	Participants: Receives COMMIT
	Participants: Commits the change locally



*/

func (node *Node) createCommunicationChannel(prepareMessage Message) chan Message {
	node.muCommunicationChannels.Lock()
	defer node.muCommunicationChannels.Unlock()

	comm := make(chan Message, 32)
	node.communicationChannels[prepareMessage.id] = comm
	return comm
}

func (node *Node) deleteCommunicationChannel(prepareMessage Message) {
	node.muCommunicationChannels.Lock()
	defer node.muCommunicationChannels.Unlock()

	delete(node.communicationChannels, prepareMessage.id)
}

func (node *Node) coordinate2PC(cmd string) bool {
	if !node.connectedToNetwork {
		commit := node.createMessage("", TxID(""), cmd)
		node.doLocalCommit(commit)
		return true
	}

	node.muVotingResource.Lock()
	defer node.muVotingResource.Unlock()

	// Build a message with type = PREPARE, and payload = "Field=New_Value"
	prepareMsg := node.createVoteMessage(common.PREPARE, cmd)

	comm := node.createCommunicationChannel(prepareMsg)
	defer node.deleteCommunicationChannel(prepareMsg)

	// Broadcast the PREPARE to all nodes
	node.Broadcast(prepareMsg)
	fmt.Printf("[%s] 2PC coordinator: broadcast PREPARE %s\n", node.name, prepareMsg.String())

	// Wait for PREPARE_ACK from all alive nodes to proceed
	ackCount := 0

	node.Broadcast(node.createMessage(common.PREPARE_ACK, prepareMsg.id, ""))

	timeToComplete := time.After(time.Millisecond * 1000)

	for {
		if ackCount == len(node.GetAliveNodes()) {
			// Everyone acknowledged, so let's COMMIT
			go node.commit2PC(prepareMsg, cmd)
			return true
		}

		select {
		case msg := <-comm:
			// We only care about messages with our TxID
			if msg.id != prepareMsg.id {
				continue
			}

			if !node.aliveNodesManager.IsNodeAlive(msg.sender) {
				node.abort2PC(prepareMsg)

				node.Connect()
				return false
			}

			switch msg.messageType {
			case common.PREPARE_ACK:
				// Successfully received an PREPARE_ACK
				ackCount++
				fmt.Printf("[%s] 2PC coordinator got PREPARE_ACK from %s\n", node.name, msg.sender)

			case common.ABORT_COMMIT:
				// Some participant aborted -> we must abort
				fmt.Printf("[%s] 2PC coordinator sees ABORT from %s => ABORT.\n", node.name, msg.sender)
				node.abort2PC(prepareMsg)
				return false
			}

		case <-timeToComplete:
			if ackCount == len(node.GetAliveNodes()) {
				// Everyone acknowledged, so let's COMMIT
				go node.commit2PC(prepareMsg, cmd)
				return true
			}

			// Timed out waiting for all ACKs => ABORT
			fmt.Printf("[%s] 2PC coordinator timed out waiting for ACKs => ABORT.\n", node.name)
			go node.abort2PC(prepareMsg)

			node.protocolTimedOut()
			return false
		}
	}
}

func (node *Node) participate2PC(prepareMsg Message) {
	if node.isTxIDFromUs(prepareMsg.id) {
		// We don't want to vote for ourselves
		return
	}

	// Attempt to lock so no other protocols run concurrently
	//ok := node.muVotingResource.TryLock()
	//if !ok {
	//	node.abort2PC(prepareMsg.id)
	//	return
	//}
	//defer node.muVotingResource.Unlock()

	// Parse the payload to get the command
	// Decide if we can do this command

	comm := node.createCommunicationChannel(prepareMsg)
	defer node.deleteCommunicationChannel(prepareMsg)

	canCommit := true
	if canCommit {
		// Send PREPARE_ACK back to coordinator
		ackMsg := node.createMessage(common.PREPARE_ACK, prepareMsg.id, "")
		node.BroadcastResponse(ackMsg, prepareMsg)
		fmt.Printf("[%s] 2PC participant => PREPARE_ACK. Waiting for COMMIT/ABORT.\n", node.name)
	} else {
		// If we can't commit, broadcast ABORT and return
		fmt.Printf("[%s] 2PC participant can't commit => ABORT.\n", node.name)
		node.abort2PC(prepareMsg)
		return
	}

	// After acknowledging the prepare, we must wait for COMMIT or ABORT from the coordinator
	// This while keeping the lock so no other protocols run concurrently
	timeout := time.After(2 * time.Second) // TODO: How long?
	for {
		select {
		case msg := <-comm:
			if msg.id != prepareMsg.id {
				continue
			}
			switch msg.messageType {
			case common.COMMIT:
				go node.doLocalCommit(msg)
				return
			case common.ABORT_COMMIT:
				return
			}
		case <-timeout:
			fmt.Printf("[%s] 2PC participant timed out waiting => ABORT.\n", node.name)
			node.protocolTimedOut()
			return
		}
	}
}

func (node *Node) commit2PC(prepareMessage Message, payload string) {
	commitMsg := node.createMessage(common.COMMIT, prepareMessage.id, payload)
	node.BroadcastResponse(commitMsg, prepareMessage)
	node.doLocalCommit(commitMsg)
}

func (node *Node) abort2PC(prepareMessage Message) {
	abortMsg := node.createMessage(common.ABORT_COMMIT, prepareMessage.id, "")
	node.BroadcastResponse(abortMsg, prepareMessage)
}

func (node *Node) doLocalCommit(msg Message) {

	// Parse the payload to get the command
	fmt.Printf("[%s] Doing commit %s.\n", node.name, msg.payload)

	node.sharedStateCommunication.FromNetwork.TwoPhaseCommit.ProtocolCommited <- msg.payload
}
