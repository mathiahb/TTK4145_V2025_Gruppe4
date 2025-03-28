package peerToPeer

import (
	"container/heap"
	"elevator_project/common"
	"sync"
)

// Example from Golang container documentation for IntHeap
type dependency_heap []Dependency

func (h dependency_heap) Len() int               { return len(h) }
func (h dependency_heap) Less(i int, j int) bool { return h[i].Is_Less_Than(h[j]) }
func (h dependency_heap) Swap(i int, j int)      { h[i], h[j] = h[j], h[i] }

func (h *dependency_heap) Push(dependency any) {
	*h = append(*h, dependency.(Dependency))
}

func (h *dependency_heap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type Dependency_Handler struct {
	mu sync.Mutex

	min_heap   *dependency_heap
	lookup_map map[Dependency]struct{}
}

func NewDependencyHandler() Dependency_Handler {
	min_heap := make(dependency_heap, 0)

	heap.Init(&min_heap)

	return Dependency_Handler{
		min_heap:   &min_heap,
		lookup_map: make(map[Dependency]struct{}),
	}
}

func (handler *Dependency_Handler) add_Dependency(dependency Dependency) {
	if dependency.Dependency_Owner == "" {
		// No dependency.
		return
	}

	if handler.min_heap.Len() == common.P2P_DEP_TIME_HORIZON {
		old_dependency := heap.Pop(handler.min_heap).(Dependency)
		delete(handler.lookup_map, old_dependency)
	}

	heap.Push(handler.min_heap, dependency)
	handler.lookup_map[dependency] = struct{}{} // Creates an instance of an empty struct.
}

func (handler *Dependency_Handler) Add_Dependency(dependency Dependency) {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	handler.add_Dependency(dependency)
}

func (handler *Dependency_Handler) has_Dependency(dependency Dependency) bool {
	if dependency.Dependency_Owner == "" {
		return true
	}

	_, ok := handler.lookup_map[dependency]
	return ok
}

func (handler *Dependency_Handler) HasDependency(dependency Dependency) bool {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	return handler.has_Dependency(dependency)
}

// Performs the action of checking for dependency and adding it to the list
func (handler *Dependency_Handler) HaveSeenDependencyBefore(dependency Dependency) bool {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	result := handler.has_Dependency(dependency)

	if !result {
		handler.add_Dependency(dependency)
	}

	return result
}
