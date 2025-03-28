package peerToPeer

import (
	"elevator_project/common"
	"fmt"
	"sync"
)

type Dependency_Resolver struct {
	mu sync.Mutex

	queue_list []Dependency
	queue_head int

	Saved_Messages map[Dependency]P2PMessage
}

func NewDependencyResolver() *Dependency_Resolver {
	return &Dependency_Resolver{
		queue_list: make([]Dependency, common.P2P_MSG_TIME_HORIZON),
		queue_head: 0,

		Saved_Messages: make(map[Dependency]P2PMessage, common.P2P_MSG_TIME_HORIZON),
	}
}

func (controller *Dependency_Resolver) advance_head() {
	controller.queue_head = (controller.queue_head + 1) % common.P2P_MSG_TIME_HORIZON
}

func (controller *Dependency_Resolver) EmplaceNewMessage(message P2PMessage) {
	if message.Type != MESSAGE {
		return
	}

	controller.mu.Lock()
	defer controller.mu.Unlock()

	controller.advance_head()
	key := NewDependency(message.Sender, message.Time)

	delete(controller.Saved_Messages, controller.queue_list[controller.queue_head])

	controller.queue_list[controller.queue_head] = key
	controller.Saved_Messages[key] = message
}

func (controller *Dependency_Resolver) Get_Message(dependency Dependency) (P2PMessage, bool) {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	message, ok := controller.Saved_Messages[dependency]
	return message, ok
}

func (network *P2P_Network) handleSpecialCase(message P2PMessage) {
	switch message.Type {
	case REQUEST_MISSING_DEPENDENCY:
		requested_dependency := Dependency_From_String(message.Message)
		response, ok := network.dependencyResolver.Get_Message(requested_dependency)
		if ok {
			network.Send(response, message.Sender)
		} else {
			fmt.Printf("Error in special cases, a requested dependency was not found: %s", message.ToString())
		}
		// Other special cases?
	case MESSAGE:
		fmt.Printf("Error in special cases, a message got handled as special case: %s", message.ToString())
		return
	}
}
