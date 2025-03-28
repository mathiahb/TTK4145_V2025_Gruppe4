package network

import (
	"math/rand"
	"time"
)

type ProtocolDispatcher struct {
	commandQueue            chan string
	repeatCommand           chan string
	shouldDoSynchronization chan bool
}

func NewProtocolDispatcher() *ProtocolDispatcher {
	return &ProtocolDispatcher{
		commandQueue:            make(chan string, 32),
		repeatCommand:           make(chan string, 1),
		shouldDoSynchronization: make(chan bool, 32),
	}
}

func (dispatcher ProtocolDispatcher) DoSynchronization() {
	dispatcher.shouldDoSynchronization <- true
}

func (dispatcher ProtocolDispatcher) DoCommand(command string) {
	dispatcher.commandQueue <- command
}

func (dispatcher ProtocolDispatcher) FlushSynchronizationChannel() {
	for {
		select {
		case <-dispatcher.shouldDoSynchronization:
		default:
			return
		}
	}
}

// Waits a random amount of time between no waiting and a millisecond. This is to manage multi-master conflict.
func RandomWait() {
	time.Sleep(time.Duration(rand.Intn(int(100 * time.Millisecond))))
}

func WaitAfterProtocol() {
	time.Sleep(100 * time.Millisecond)
}

func (node *Node) startDispatcher() {
	go node.dispatcher()
}

// Dispatcher awaits calls to perform a protocol on the network, then queues the protocol for dispatch.
// Priority: Discovery > Synchronize > Commands
func (node *Node) dispatcher() {
	for {
		// First check if we should do Discovery -> Synchronize
		select {
		case <-node.protocolDispatcher.shouldDoSynchronization:
			node.protocolDispatcher.FlushSynchronizationChannel()
			success := node.coordinateSynchronization()

			if !success {
				go node.protocolDispatcher.DoSynchronization()
				RandomWait()
			}
			continue
		default:
		}

		// Then check if we aborted a command
		select {
		case command := <-node.protocolDispatcher.repeatCommand:
			success := node.coordinate2PC(command)

			if !success {
				go func() { node.protocolDispatcher.repeatCommand <- command }()
				RandomWait()
			}
			continue
		default:
		}

		// Then wait for new commands/discovery
		select {
		case <-node.protocolDispatcher.shouldDoSynchronization:
			node.protocolDispatcher.FlushSynchronizationChannel()
			success := node.coordinateSynchronization()

			if !success {
				go node.protocolDispatcher.DoSynchronization()
				RandomWait()
			}

		case command := <-node.protocolDispatcher.repeatCommand:
			success := node.coordinate2PC(command)

			if !success {
				go func() { node.protocolDispatcher.repeatCommand <- command }()
				RandomWait()
			}

		case command := <-node.protocolDispatcher.commandQueue:
			success := node.coordinate2PC(command)

			if !success {
				go func() { node.protocolDispatcher.repeatCommand <- command }()
				RandomWait()
			}
		}
		WaitAfterProtocol()
	}
}
