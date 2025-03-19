package network

import (
	"math/rand"
	"time"
)

type ProtocolDispatcher struct {
	command_queue         chan Command
	should_do_discovery   chan bool
	should_do_synchronize chan bool
}

func New_Protocol_Dispatcher() *ProtocolDispatcher {
	return &ProtocolDispatcher{
		command_queue:         make(chan Command, 32),
		should_do_discovery:   make(chan bool, 32),
		should_do_synchronize: make(chan bool, 32),
	}
}

func (dispatcher ProtocolDispatcher) Do_Discovery() {
	dispatcher.should_do_discovery <- true
}

func (dispatcher ProtocolDispatcher) Do_Synchronization() {
	dispatcher.should_do_synchronize <- true
}

func (dispatcher ProtocolDispatcher) Do_Command(command Command) {
	dispatcher.command_queue <- command
}

func (dispatcher ProtocolDispatcher) Flush_Discovery_Channel() {
	for {
		select {
		case <-dispatcher.should_do_discovery:
		default:
			return
		}
	}
}

func (dispatcher ProtocolDispatcher) Flush_Synchronization_Channel() {
	for {
		select {
		case <-dispatcher.should_do_synchronize:
		default:
			return
		}
	}
}

// Waits a random amount of time between no waiting and a millisecond. This is to manage multi-master conflict.
func Random_Wait() {
	time.Sleep(time.Duration(rand.Intn(int(time.Millisecond))))
}

func Wait_After_Protocol() {
	time.Sleep(time.Millisecond)
}

func (node *Node) start_dispatcher() {
	go node.dispatcher()
}

// Dispatcher awaits calls to perform a protocol on the network, then queues the protocol for dispatch.
// Priority: Discovery > Synchronize > Commands
func (node *Node) dispatcher() {
	success_channel := make(chan bool)

	for {
		select {
		case <-node.protocol_dispatcher.should_do_discovery:
			go node.coordinate_Discovery(success_channel)
			success := <-success_channel
			if !success {
				go node.protocol_dispatcher.Do_Discovery()
				Random_Wait()
			}

		case <-node.protocol_dispatcher.should_do_synchronize:
			go node.coordinate_Synchronization(success_channel)
			success := <-success_channel
			if !success {
				go node.protocol_dispatcher.Do_Synchronization()
				Random_Wait()
			}

		case command := <-node.protocol_dispatcher.command_queue:
			command.Field = "" // Dummy task
		}

		Wait_After_Protocol()
	}
}
