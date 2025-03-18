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

	node.coordinating = true

	message := node.create_Vote_Message(Constants.DISCOVERY_BEGIN, "")
	node.Broadcast(message)

	time_to_complete := time.After(time.Millisecond * 100)
	result := make([]string, 1, 10)
	result[0] = node.name

	for {
		select {
		case response := <-node.comm:
			// Hello!
			fmt.Printf("[Debug %s]: Received forwarding during coordination: %s\n", node.name, response.String())

			if response.message_type == Constants.DISCOVERY_HELLO && response.id == message.id {
				result = append(result, response.payload)
			}
			// Aborted
			if response.message_type == Constants.ABORT_COMMIT && response.id == message.id {
				node.abort_Discovery(message.id)
				return
			}
		case <-time_to_complete:
			node.broadcast_Discovery_Result(message.id, result)
			node.alive_nodes_manager.Set_Alive_Nodes(result)
			node.new_alive_nodes <- result

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

func (node *Node) say_Hello_To_Discovery(p2p_message peer_to_peer.P2P_Message, id_discovery TxID) {
	message := node.create_Message(Constants.DISCOVERY_HELLO, id_discovery, node.name)
	node.Broadcast_Response(message, p2p_message)
}

func (node *Node) participate_In_Discovery(p2p_message peer_to_peer.P2P_Message, id_discovery TxID) {
	if node.isTxIDFromUs(id_discovery) {
		fmt.Printf("[%s] Detected discovery started by us!\n", node.name)
		return
	}

	ok := node.mu_voting_resource.TryLock()
	if !ok {
		node.abort_Discovery(id_discovery)
		return
	}
	defer node.mu_voting_resource.Unlock()

	fmt.Printf("[%s] Participating in discovery %s!\n", node.name, id_discovery)

	node.say_Hello_To_Discovery(p2p_message, id_discovery)

	timeout := time.After(time.Second)
	for {
		select {
		case result_message := <-node.comm:
			fmt.Printf("[Debug %s] Received forwarding during participation: %s\n", node.name, result_message.String())

			if result_message.message_type == Constants.DISCOVERY_COMPLETE && result_message.id == id_discovery {
				result := strings.Split(result_message.payload, ":")
				node.alive_nodes_manager.Set_Alive_Nodes(result)

				node.new_alive_nodes <- result
				fmt.Printf("[Debug %s] Received result %s\n", node.name, node.Get_Alive_Nodes())
				return
			}
		case <-timeout:
			fmt.Printf("ERROR: Discovery %s halted in progress!\n", id_discovery)
			return
		}
	}
}
