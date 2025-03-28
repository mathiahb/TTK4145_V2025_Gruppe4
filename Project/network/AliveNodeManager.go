package network

import "sync"

type AliveNodeManager struct {
	mu         sync.Mutex
	aliveNodes []string
}

func (manager *AliveNodeManager) GetAliveNodes() []string {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.aliveNodes
}

func (manager *AliveNodeManager) SetAliveNodes(newNodes []string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.aliveNodes = newNodes
}

func (manager *AliveNodeManager) IsNodeAlive(node string) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for _, aliveNode := range manager.aliveNodes {
		if aliveNode == node {
			return true
		}
	}
	return false
}
