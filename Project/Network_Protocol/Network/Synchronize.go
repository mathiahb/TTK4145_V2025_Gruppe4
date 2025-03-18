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

func (node *Node) coordinate_Synchronization() {
	node.mu_voting_resource.Lock()
	defer node.mu_voting_resource.Unlock()

	payload := ""

	message := node.create_Vote_Message(Constants.SYNC_AFTER_DISCOVERY, payload)
	node.Broadcast(message)

	time_to_complete := time.After(time.Millisecond * 100)
	result := payload

	for {
		select {
		case <-time_to_complete:
			node.broadcast_Synchronization_Result(message.id, result)
			return
		}
	}
}

func (node *Node) broadcast_Synchronization_Result(id TxID, result string) {
	message := node.create_Message(Constants.DISCOVERY_COMPLETE, id, result)
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

	node.voting_on = id_discovery

	message := node.create_Message(Constants.SYNC_RESPONSE, id_discovery, node.name)
	node.Broadcast_Response(message, p2p_message)

	timeout := time.After(time.Second)

	for {
		select {
		case result := <-node.comm:
			if result.message_type == Constants.SYNC_RESULT && result.id == id_discovery {
				return
			}
		case <-timeout:
			fmt.Printf("ERROR: Discovery %s halted in progress!\n", id_discovery)
			return
		}
	}
}
