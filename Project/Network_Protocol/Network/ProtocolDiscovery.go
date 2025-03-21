package network

import (
	Constants "elevator_project/constants"
	"fmt"
	"strings"
	"time"

	peer_to_peer "elevator_project/Network_Protocol/Network/Peer_to_Peer"
)

// PROTOCOL - Discovery
/*
	Message				:	CODE	PAYLOAD
	---
	Discovery Begin		: 	NDSC
	Discovery Hello		:	HELO	NodeName
	Discovery Complete	:	DSCC	NodeName1:NodeName2:NodeName3:...
	Abort Commit		:	ERRC
	---

	Expected procedure
	---
	Coordinator: Discovery dispatched -> go coordinate_Discovery(success_channel)
	Coordinator: Broadcast Discovery Begin

	Participants: Other nodes receives Discovery Begin -> go participate_Discovery()
	Participants: Every node responds Discovery Hello OR Abort Commit

	Coordinator: Receives responses
	If Abort Commit:
		Coordinator: Broadcast Abort Commit
		Coordinator: Returns false on the success_channel

	Timeout passes:
		Coordinator: Compiles results
		Coordinator: Broadcast Discovery Complete
		Coordinator: Passes the new node list to new_node_channel
		Coordinator: Returns true on the success_channel
		Participants: Receives Discovery Complete and recompiles result
		Participants: Passes the new node list to new_node_channel
*/

func (node *Node) coordinate_Discovery(success_channel chan bool) {
	node.mu_voting_resource.Lock()
	defer node.mu_voting_resource.Unlock()

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

				success_channel <- false
				return
			}
		case <-time_to_complete:
			node.broadcast_Discovery_Result(message.id, result)
			node.alive_nodes_manager.Set_Alive_Nodes(result)
			node.new_alive_nodes <- result

			success_channel <- true
			return
		}
	}
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
			fmt.Printf("[ERROR %s] Discovery %s halted in progress!\n", node.name, id_discovery)
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
