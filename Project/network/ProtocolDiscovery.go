package network

import (
	Constants "elevator_project/constants"
	"fmt"
	"strings"
	"time"
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

func (id TxID) isOlderThan(other TxID) bool {
	split_self := strings.Split(string(id), ":")
	split_other := strings.Split(string(other), ":")

	name_self := split_self[0]
	name_other := split_other[0]

	if name_self != name_other {
		return false
	}

	id_number := split_self[1]
	other_number := split_other[1]

	return id_number < other_number
}

func (node *Node) send_Discovery_Result(result []string) {
	node.shared_state_communication.FromNetwork.Discovery.Updated_Alive_Nodes <- result
}

func (node *Node) coordinate_Discovery(success_channel chan bool) {
	node.mu_voting_resource.Lock()
	defer node.mu_voting_resource.Unlock()

	begin_discovery_message := node.create_Vote_Message(Constants.DISCOVERY_BEGIN, "")
	node.Broadcast(begin_discovery_message)

	time_to_complete := time.After(time.Millisecond * 100)
	result := make([]string, 1, 10)
	result[0] = node.name

	for {
		select {
		case response := <-node.comm:

			// Hello!
			fmt.Printf("[Debug %s]: Received forwarding during coordination: %s\n", node.name, response.String())

			if response.message_type == Constants.DISCOVERY_HELLO && response.id == begin_discovery_message.id {
				result = append(result, response.payload)
			}
			// Aborted
			if response.message_type == Constants.ABORT_COMMIT && response.id == begin_discovery_message.id {
				node.abort_Discovery(begin_discovery_message)

				success_channel <- false
				return
			}
		case <-time_to_complete:
			node.broadcast_Discovery_Result(begin_discovery_message, result)
			node.alive_nodes_manager.Set_Alive_Nodes(result)
			go node.send_Discovery_Result(result)

			node.coordinate_Synchronization(success_channel, begin_discovery_message)
			return
		}
	}
}

func (node *Node) participate_In_Discovery(discovery_message Message) {
	if node.isTxIDFromUs(discovery_message.id) {
		fmt.Printf("[%s] Detected discovery started by us!\n", node.name)
		return
	}

	node.mu_voting_resource.Lock()
	defer node.mu_voting_resource.Unlock()

	fmt.Printf("[%s] Participating in discovery %s!\n", node.name, discovery_message.id)

	node.say_Hello_To_Discovery(discovery_message)

	timeout := time.After(time.Second)
	for {
		select {
		case result_message := <-node.comm:
			fmt.Printf("[Debug %s] Received forwarding during participation: %s\n", node.name, result_message.String())

			if discovery_message.id.isOlderThan(result_message.id) {
				fmt.Printf("[Debug %s] Received newer message %s, giving up on old message %s\n", node.name, result_message.String(), discovery_message.id)
				return
			}

			if result_message.message_type == Constants.DISCOVERY_COMPLETE && result_message.id == discovery_message.id {
				result := strings.Split(result_message.payload, ":")
				node.alive_nodes_manager.Set_Alive_Nodes(result)
				go node.send_Discovery_Result(result)

				fmt.Printf("[Debug %s] Received result %s\n", node.name, node.Get_Alive_Nodes())
				node.participate_In_Synchronization(discovery_message) // Successful discovery -> Synchronization
				return
			}
		case <-timeout:
			fmt.Printf("[ERROR %s] Discovery %s halted in progress!\n", node.name, discovery_message.id)
			return
		}
	}
}

func (node *Node) broadcast_Discovery_Result(discovery_message Message, result []string) {
	result_string := ""

	for i, name := range result {
		if i != 0 {
			result_string = result_string + ":"
		}
		result_string = result_string + name
	}

	message := node.create_Message(Constants.DISCOVERY_COMPLETE, discovery_message.id, result_string)
	node.Broadcast_Response(message, discovery_message)
}

func (node *Node) abort_Discovery(discovery_message Message) {
	message := node.create_Message(Constants.ABORT_DISCOVERY, discovery_message.id, "")
	node.Broadcast_Response(message, discovery_message)
}

func (node *Node) say_Hello_To_Discovery(discovery_message Message) {
	message := node.create_Message(Constants.DISCOVERY_HELLO, discovery_message.id, node.name)
	node.Broadcast_Response(message, discovery_message)
}
