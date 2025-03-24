package network

import "sync"

type AliveNodeManager struct {
	mu          sync.Mutex
	alive_nodes []string
}

func (manager *AliveNodeManager) Get_Alive_Nodes() []string {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.alive_nodes
}

func (manager *AliveNodeManager) Set_Alive_Nodes(new_nodes []string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.alive_nodes = new_nodes
}

func (manager *AliveNodeManager) Is_Node_Alive(node string) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for _, alive_node := range manager.alive_nodes {
		if alive_node == node {
			return true
		}
	}
	return false
}
