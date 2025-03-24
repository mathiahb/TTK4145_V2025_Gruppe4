package network

import (
	Constants "elevator_project/constants"
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
	Coordinator: 2PC dispatched -> go node.coordinate_2PC(success_channel)
	Coordinator: Broadcast PREPARE
	Coordinator: Waits for responses

	Participants: Other nodes receives PREPARE -> go node.participate_2PC()
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

func (node *Node) create_2PC_comm(prepare_message Message) chan Message {
	node.mu_twopc_comm.Lock()
	defer node.mu_twopc_comm.Unlock()

	comm := make(chan Message, 32)
	node.twopc_comm[prepare_message.id] = comm
	return comm
}

func (node *Node) delete_2PC_comm(prepare_message Message) {
	node.mu_twopc_comm.Lock()
	defer node.mu_twopc_comm.Unlock()

	delete(node.twopc_comm, prepare_message.id)
}

func (node *Node) coordinate_2PC(cmd string, success_channel chan bool) {
	node.mu_voting_resource.Lock()
	defer node.mu_voting_resource.Unlock()

	// Build a message with type = PREPARE, and payload = "Field=New_Value"
	prepareMsg := node.create_Vote_Message(Constants.PREPARE, cmd)

	comm := node.create_2PC_comm(prepareMsg)
	defer node.delete_2PC_comm(prepareMsg)

	// Broadcast the PREPARE to all nodes
	node.Broadcast(prepareMsg)
	fmt.Printf("[%s] 2PC coordinator: broadcast PREPARE %s\n", node.name, prepareMsg.String())

	// Wait for PREPARE_ACK from all alive nodes to proceed
	neededAcks := len(node.Get_Alive_Nodes())
	ackCount := 0

	node.Broadcast(node.create_Message(Constants.PREPARE_ACK, prepareMsg.id, ""))

	time_to_complete := time.After(time.Millisecond * 800)

	for {
		if ackCount == neededAcks {
			// Everyone acknowledged, so let's COMMIT
			go node.commit2PC(prepareMsg, cmd)
			success_channel <- true
			return
		}

		select {
		case msg := <-comm:
			// We only care about messages with our TxID
			if msg.id != prepareMsg.id {
				continue
			}

			if !node.alive_nodes_manager.Is_Node_Alive(msg.sender) {
				node.abort2PC(prepareMsg)

				node.Connect()
				success_channel <- false
				continue
			}

			switch msg.message_type {
			case Constants.PREPARE_ACK:
				// Successfully received an PREPARE_ACK
				ackCount++
				fmt.Printf("[%s] 2PC coordinator got PREPARE_ACK from %s\n", node.name, msg.sender)

			case Constants.ABORT_COMMIT:
				// Some participant aborted -> we must abort
				fmt.Printf("[%s] 2PC coordinator sees ABORT from %s => ABORT.\n", node.name, msg.sender)
				node.abort2PC(prepareMsg)
				success_channel <- false
				return
			}

		case <-time_to_complete:
			// Timed out waiting for all ACKs => ABORT
			fmt.Printf("[%s] 2PC coordinator timed out waiting for ACKs => ABORT.\n", node.name)
			go node.abort2PC(prepareMsg)

			node.protocol_timed_out()
			success_channel <- false
			return
		}
	}
}

func (node *Node) participate_2PC(prepareMsg Message) {
	if node.isTxIDFromUs(prepareMsg.id) {
		// We don't want to vote for ourselves
		return
	}

	// Attempt to lock so no other protocols run concurrently
	//ok := node.mu_voting_resource.TryLock()
	//if !ok {
	//	node.abort2PC(prepareMsg.id)
	//	return
	//}
	//defer node.mu_voting_resource.Unlock()

	// Parse the payload to get the command
	// Decide if we can do this command

	comm := node.create_2PC_comm(prepareMsg)
	defer node.delete_2PC_comm(prepareMsg)

	canCommit := true
	if canCommit {
		// Send PREPARE_ACK back to coordinator
		ackMsg := node.create_Message(Constants.PREPARE_ACK, prepareMsg.id, "")
		node.Broadcast_Response(ackMsg, prepareMsg)
		fmt.Printf("[%s] 2PC participant => PREPARE_ACK. Waiting for COMMIT/ABORT.\n", node.name)
	} else {
		// If we can't commit, broadcast ABORT and return
		fmt.Printf("[%s] 2PC participant can't commit => ABORT.\n", node.name)
		node.abort2PC(prepareMsg)
		return
	}

	// After acknowledging the prepare, we must wait for COMMIT or ABORT from the coordinator
	// This while keeping the lock so no other protocols run concurrently
	timeout := time.After(1 * time.Second) // TODO: How long?
	for {
		select {
		case msg := <-comm:
			if msg.id != prepareMsg.id {
				continue
			}
			switch msg.message_type {
			case Constants.COMMIT:
				go node.doLocalCommit(msg)
				return
			case Constants.ABORT_COMMIT:
				return
			}
		case <-timeout:
			fmt.Printf("[%s] 2PC participant timed out waiting => ABORT.\n", node.name)
			node.protocol_timed_out()
			return
		}
	}
}

func (node *Node) commit2PC(prepare_message Message, payload string) {
	commitMsg := node.create_Message(Constants.COMMIT, prepare_message.id, payload)
	node.Broadcast_Response(commitMsg, prepare_message)
	node.doLocalCommit(commitMsg)
}

func (node *Node) abort2PC(prepare_message Message) {
	abortMsg := node.create_Message(Constants.ABORT_COMMIT, prepare_message.id, "")
	node.Broadcast_Response(abortMsg, prepare_message)
}

func (node *Node) doLocalCommit(msg Message) {

	// Parse the payload to get the command
	fmt.Printf("[%s] Doing commit %s.\n", node.name, msg.payload)

	node.shared_state_communication.FromNetwork.TwoPhaseCommit.ProtocolCommited <- msg.payload
}
