package network

import (
	"math/rand"
	"time"
)

type ProtocolDispatcher struct {
	command_queue       chan string
	repeat_command      chan string
	should_do_discovery chan bool
}

func New_Protocol_Dispatcher() *ProtocolDispatcher {
	return &ProtocolDispatcher{
		command_queue:       make(chan string, 32),
		repeat_command:      make(chan string, 1),
		should_do_discovery: make(chan bool, 32),
	}
}

func (dispatcher ProtocolDispatcher) Do_Discovery() {
	dispatcher.should_do_discovery <- true
}

func (dispatcher ProtocolDispatcher) Do_Command(command string) {
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
		// First check if we should do Discovery -> Synchronize
		select {
		case <-node.protocol_dispatcher.should_do_discovery:
			node.protocol_dispatcher.Flush_Discovery_Channel()
			go node.coordinate_Discovery(success_channel)
			success := <-success_channel
			if !success {
				go node.protocol_dispatcher.Do_Discovery()
				Random_Wait()
			}
			continue
		default:
		}

		// Then check if we aborted a command
		select {
		case command := <-node.protocol_dispatcher.repeat_command:
			go node.coordinate_2PC(command, success_channel)
			success := <-success_channel
			if !success {
				go func() { node.protocol_dispatcher.repeat_command <- command }()
				Random_Wait()
			}
			continue
		default:
		}

		// Then wait for new commands/discovery
		select {
		case <-node.protocol_dispatcher.should_do_discovery:
			node.protocol_dispatcher.Flush_Discovery_Channel()
			go node.coordinate_Discovery(success_channel)
			success := <-success_channel
			if !success {
				go node.protocol_dispatcher.Do_Discovery()
				Random_Wait()
			}

		case command := <-node.protocol_dispatcher.command_queue:
			go node.coordinate_2PC(command, success_channel)
			success := <-success_channel
			if !success {
				go func() { node.protocol_dispatcher.repeat_command <- command }()
				Random_Wait()
			}
		}
		Wait_After_Protocol()
	}
}
