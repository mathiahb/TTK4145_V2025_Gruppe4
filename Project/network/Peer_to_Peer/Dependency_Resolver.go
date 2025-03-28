package peerToPeer

import (
	"elevator_project/common"
	"fmt"
	"sync"
)

type DependencyResolver struct {
	mu sync.Mutex

	queueList []Dependency
	queueHead int

	SavedMessages map[Dependency]P2PMessage
}

func NewDependencyResolver() *DependencyResolver {
	return &DependencyResolver{
		queueList: make([]Dependency, common.P2P_MSG_TIME_HORIZON),
		queueHead: 0,

		SavedMessages: make(map[Dependency]P2PMessage, common.P2P_MSG_TIME_HORIZON),
	}
}

func (controller *DependencyResolver) advanceHead() {
	controller.queueHead = (controller.queueHead + 1) % common.P2P_MSG_TIME_HORIZON
}

func (controller *DependencyResolver) EmplaceNewMessage(message P2PMessage) {
	if message.Type != MESSAGE {
		return
	}

	controller.mu.Lock()
	defer controller.mu.Unlock()

	controller.advanceHead()
	key := NewDependency(message.Sender, message.Time)

	delete(controller.SavedMessages, controller.queueList[controller.queueHead])

	controller.queueList[controller.queueHead] = key
	controller.SavedMessages[key] = message
}

func (controller *DependencyResolver) GetMessage(dependency Dependency) (P2PMessage, bool) {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	message, ok := controller.SavedMessages[dependency]
	return message, ok
}

func (network *P2P_Network) handleSpecialCase(message P2PMessage) {
	switch message.Type {
	case REQUEST_MISSING_DEPENDENCY:
		requestedDependency := DependencyFromString(message.Message)
		response, ok := network.dependencyResolver.GetMessage(requestedDependency)
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
