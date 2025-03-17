package network

import (
	"Constants"
	"fmt"
	"strings"
	"time"

	peer_to_peer "github.com/mathiahb/TTK4145_V2025_Gruppe4/Network_Protocol/Network/Peer_to_Peer"
)

func (node *Node) coordinate_Discovery() {
	node.mu_voting_resource.Lock()
	defer node.mu_voting_resource.Unlock()

	message := node.create_Vote_Message(Constants.DISCOVERY_BEGIN, "")
	node.Broadcast(message)

	time_to_complete := time.After(time.Millisecond * 100)
	result := make([]string, 0, 10)

	for {
		select {
		case <-time_to_complete:
			node.broadcast_Discovery_Result(message.id, result)
			return
		}
	}
}

func (node *Node) broadcast_Discovery_Result(id TxID, result []string) {
	result_string := ""

	for i, name := range result {
		if i != 0 {
			result_string = result_string + ":"
		}
		result_string = result_string + name
	}

	message := node.create_Message(Constants.DISCOVERY_COMPLETE, id, result_string)
	node.Broadcast(message)
}

func (node *Node) abort_Discovery(id_discovery TxID) {
	message := node.create_Message(Constants.ABORT_COMMIT, id_discovery, "")
	node.Broadcast(message)
}

func (node *Node) participate_In_Discovery(p2p_message peer_to_peer.P2P_Message, id_discovery TxID) {
	ok := node.mu_voting_resource.TryLock()
	if !ok {
		node.abort_Discovery(id_discovery)
		return
	}
	defer node.mu_voting_resource.Unlock()

	node.voting_on = id_discovery

	message := node.create_Message(Constants.DISCOVERY_HELLO, id_discovery, node.name)
	node.Broadcast_Response(message, p2p_message)

	timeout := time.After(time.Second)

	for {
		select {
		case result := <-node.comm:
			if result.message_type == Constants.DISCOVERY_COMPLETE && result.id == id_discovery {
				node.alive_nodes = strings.Split(result.payload, ":")
				return
			}
		case <-timeout:
			fmt.Printf("ERROR: Discovery %s halted in progress!\n", id_discovery)
			return
		}
	}
}
