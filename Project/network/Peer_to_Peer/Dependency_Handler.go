package peerToPeer

import (
	"container/heap"
	"elevator_project/common"
	"sync"
)

// Example from Golang container documentation for IntHeap
type dependencyHeap []Dependency

func (h dependencyHeap) Len() int               { return len(h) }
func (h dependencyHeap) Less(i int, j int) bool { return h[i].IsLessThan(h[j]) }
func (h dependencyHeap) Swap(i int, j int)      { h[i], h[j] = h[j], h[i] }

func (h *dependencyHeap) Push(dependency any) {
	*h = append(*h, dependency.(Dependency))
}

func (h *dependencyHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type DependencyHandler struct {
	mu sync.Mutex

	minHeap   *dependencyHeap
	lookupMap map[Dependency]struct{}
}

func NewDependencyHandler() DependencyHandler {
	minHeap := make(dependencyHeap, 0)

	heap.Init(&minHeap)

	return DependencyHandler{
		minHeap:   &minHeap,
		lookupMap: make(map[Dependency]struct{}),
	}
}

func (handler *DependencyHandler) addDependency(dependency Dependency) {
	if dependency.DependencyOwner == "" {
		// No dependency.
		return
	}

	if handler.minHeap.Len() == common.P2P_DEP_TIME_HORIZON {
		oldDependency := heap.Pop(handler.minHeap).(Dependency)
		delete(handler.lookupMap, oldDependency)
	}

	heap.Push(handler.minHeap, dependency)
	handler.lookupMap[dependency] = struct{}{} // Creates an instance of an empty struct.
}

func (handler *DependencyHandler) AddDependency(dependency Dependency) {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	handler.addDependency(dependency)
}

func (handler *DependencyHandler) hasDependency(dependency Dependency) bool {
	if dependency.DependencyOwner == "" {
		return true
	}

	_, ok := handler.lookupMap[dependency]
	return ok
}

func (handler *DependencyHandler) HasDependency(dependency Dependency) bool {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	return handler.hasDependency(dependency)
}

// Performs the action of checking for dependency and adding it to the list
func (handler *DependencyHandler) HaveSeenDependencyBefore(dependency Dependency) bool {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	result := handler.hasDependency(dependency)

	if !result {
		handler.addDependency(dependency)
	}

	return result
}
