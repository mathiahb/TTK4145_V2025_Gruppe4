package network

import (
	"Constants"
	"fmt"
	"strings"
	"time"

	peer_to_peer "github.com/mathiahb/TTK4145_V2025_Gruppe4/Network_Protocol/Network/Peer_to_Peer"
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

type Command struct {
	Field     string
	New_Value string
}

func (node *Node) coordinate_2PC(cmd Command, success_channel chan bool) {
	node.mu_voting_resource.Lock()
	defer node.mu_voting_resource.Unlock()

	// Build a message with type = PREPARE, and payload = "Field=New_Value"
	txid := node.generateTxID()
	payload := fmt.Sprintf("%s=%s", cmd.Field, cmd.New_Value)
	prepareMsg := node.create_Message(Constants.PREPARE, txid, payload)

	// Broadcast the PREPARE to all nodes
	node.Broadcast(prepareMsg)
	fmt.Printf("[%s] 2PC coordinator: broadcast PREPARE %s\n", node.name, prepareMsg.String())

	// Wait for PREPARE_ACK from all alive nodes to proceed
	neededAcks := len(node.Get_Alive_Nodes())
	ackCount := 0

	time_to_complete := time.After(time.Millisecond * 100)

	for {
		if ackCount == neededAcks {
			// Everyone acknowledged, so let's COMMIT
			node.commit2PC(txid, payload)
			success_channel <- true

			// TODO: Her forlater vi funksjonen, uten å låse den igjen i commit2PC.
			// SKal vi heller bare dra commit2PC hit slik at låsen forblir?
			return
		}

		select {
		case msg := <-node.comm:
			// We only care about messages with our TxID
			if msg.id != txid {
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
				node.abort2PC(txid)
				success_channel <- false
				return
			}

		case <-time_to_complete:
			// Timed out waiting for all ACKs => ABORT
			fmt.Printf("[%s] 2PC coordinator timed out waiting for ACKs => ABORT.\n", node.name)
			node.abort2PC(txid)
			success_channel <- false
			// TOOD: initiate a discovery?
			return
		}
	}
}

func (node *Node) participate_2PC(p2p_message peer_to_peer.P2P_Message, prepareMsg Message) {
	// Attempt to lock so no other protocols run concurrently
	ok := node.mu_voting_resource.TryLock()
	if !ok {
		node.abort2PC(prepareMsg.id)
		return
	}
	defer node.mu_voting_resource.Unlock()

	// Parse the payload to get the command
	field, newVal := parseCommandPayload(prepareMsg.payload)

	// Decide if we can do this command
	canCommit := checkIfWeCanCommit(field, newVal)
	if canCommit {
		// Send PREPARE_ACK back to coordinator
		ackMsg := node.create_Message(Constants.PREPARE_ACK, prepareMsg.id, "")
		node.Broadcast_Response(ackMsg, p2p_message)
		fmt.Printf("[%s] 2PC participant => PREPARE_ACK. Waiting for COMMIT/ABORT.\n", node.name)
	} else {
		// If we can't commit, broadcast ABORT and return
		fmt.Printf("[%s] 2PC participant can't commit => ABORT.\n", node.name)
		node.abort2PC(prepareMsg.id)
		return
	}

	// After acknowledging the prepare, we must wait for COMMIT or ABORT from the coordinator
	// This while keeping the lock so no other protocols run concurrently
	timeout := time.After(1 * time.Second) // TODO: How long?
	for {
		select {
		case msg := <-node.comm:
			if msg.id != prepareMsg.id {
				continue
			}
			switch msg.message_type {
			case Constants.COMMIT:
				node.doLocalCommit(msg)
				// TODO: send ACK back to coordinator?
				return
			case Constants.ABORT_COMMIT:
				// node.doLocalAbort(msg)
				return
			}
		case <-timeout:
			// If the coordinator never finalized => local abort
			fmt.Printf("[%s] 2PC participant timed out waiting => ABORT.\n", node.name)
			// node.doLocalAbort(prepareMsg)
			return
		}
	}
}

func (node *Node) commit2PC(txid TxID, payload string) {
	commitMsg := node.create_Message(Constants.COMMIT, txid, payload)
	node.Broadcast(commitMsg)

	//TODO: Somewhere we need to wait for the ACKs from the participants?

	node.doLocalCommit(commitMsg)
}

func (node *Node) abort2PC(txid TxID) {
	abortMsg := node.create_Message(Constants.ABORT_COMMIT, txid, "")
	node.Broadcast(abortMsg)

	// node.doLocalAbort(abortMsg)
	// Is this necessary? We're not really doing anything with the abort locally
}

func parseCommandPayload(payload string) (string, string) {
	parts := strings.SplitN(payload, "=", 2)
	if len(parts) < 2 {
		return payload, ""
	}
	return parts[0], parts[1]
}

func checkIfWeCanCommit(field, newVal string) bool {
	// TODO: Implement checks on the field and new value (if needed)
	return true
}

func (node *Node) doLocalCommit(msg Message) {

	// Parse the payload to get the command
	field, newVal := parseCommandPayload(msg.payload)
	// TODO:  Do changes locally based on the command
	fmt.Printf("set Node [%s]  %f=%v\n", node.name, field, newVal)
}

// func (node *Node) doLocalAbort(Message) {
// 	fmt.Printf("[Local %s] Doing abort.\n", node.name)
// }
