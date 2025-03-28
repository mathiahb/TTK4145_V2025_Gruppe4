package peerToPeer

import (
	"elevator_project/common"
	"elevator_project/network/Peer_to_Peer/TCP"
	"elevator_project/network/Peer_to_Peer/UDP"
	"fmt"
	"time"
)

// Package peerToPeer
//
// Handles Peer Detection over UDP and Communication over TCP
// Does not handle elevator detection (not all peers must be elevators, some may be listeners!)

type P2P_Network struct {
	ReadChannel chan P2PMessage

	TCP *TCP.TCPConnectionManager
	UDP UDP.UDPChannel

	closeChannel chan bool

	tcpServerAddress string

	dependencyResolver *DependencyResolver
	dependencyHandler  DependencyHandler
	clock              LamportClock
}

func NewP2PNetwork() *P2P_Network {
	readChannel := make(chan P2PMessage, common.P2P_BUFFER_SIZE)

	tcpManager := TCP.NewTCPConnectionManager()
	udpChannel := UDP.NewUDPChannel()

	serverPort := tcpManager.OpenServer()
	serverAddress := UDP.GetLocalIP().String() + serverPort
	fmt.Printf("Opened server at: %s\n", serverAddress)

	network := P2P_Network{
		ReadChannel: readChannel,

		TCP: tcpManager,
		UDP: udpChannel,

		closeChannel: make(chan bool),

		tcpServerAddress: serverAddress,

		dependencyResolver: NewDependencyResolver(),
		dependencyHandler:  NewDependencyHandler(),
		clock:              NewLamportClock(),
	}

	go network.peerDetection()
	go network.reader()

	time.Sleep(common.P2P_TIME_UNTIL_EXPECTED_ALL_CONNECTED)

	return &network
}

func (network *P2P_Network) Close() {
	close(network.closeChannel)
	network.TCP.CloseAll()
	network.UDP.Close()
}

func (network *P2P_Network) Broadcast(message P2PMessage) {
	network.dependencyResolver.EmplaceNewMessage(message)
	network.TCP.Broadcast(message.ToString())
}

func (network *P2P_Network) Send(message P2PMessage, recipient string) {
	network.TCP.Send(message.ToString(), recipient)
}

func (network *P2P_Network) CreateMessage(message string) P2PMessage {
	return network.createMessage(message, MESSAGE)
}

func (network *P2P_Network) requestDependency(dependency Dependency) {
	message := network.createMessage(dependency.ToString(), REQUEST_MISSING_DEPENDENCY)
	network.Send(message, dependency.DependencyOwner)
}

func (network *P2P_Network) createMessage(message string, messageType P2PMessageType) P2PMessage {
	network.clock.Event()

	return NewP2PMessage(network.tcpServerAddress, messageType, network.clock, message)
}
func (network *P2P_Network) reader() {
	for {
		select {
		case tcpMessage := <-network.TCP.GlobalReadChannel:
			p2pMessage := P2PMessageFromString(tcpMessage)

			network.clock.Update(p2pMessage.Time)
			go network.publisher(p2pMessage)

		case <-network.closeChannel:
			return

		default:
			time.Sleep(time.Microsecond)
		}
	}
}

func (network *P2P_Network) publisher(message P2PMessage) {
	new_dependency := NewDependency(message.Sender, message.Time)

	if network.dependencyHandler.HaveSeenDependencyBefore(new_dependency) {
		return
	}

	timeout := time.NewTimer(time.Second)

	// Wait for message dependency, then publish onto read channel.
	for {
		select {
		case <-timeout.C:
			return // Timed out.
		case <-network.closeChannel:
			return // Network closed.
		default:
		}

		if network.dependencyHandler.HasDependency(message.dependency) {
			if message.Type == MESSAGE {
				network.ReadChannel <- message
			} else {
				network.handleSpecialCase(message)
			}

			return
		}

		// Request the missing dependency.
		network.requestDependency(message.dependency)

		// Wait until more data is processed
		time.Sleep(20 * time.Millisecond)
	}
}

func (network *P2P_Network) peerDetection() {
	renewPresenceTicker := time.NewTicker(common.UDP_WAIT_BEFORE_TRANSMITTING_AGAIN)

	for {
		select {
		case <-network.closeChannel:
			return // P2P Connection closed

		case <-renewPresenceTicker.C:
			network.announcePresence()

		case address := <-network.UDP.ReadChannel:
			if !network.TCP.DoesConnectionExist(address) {
				network.TCP.ConnectClient(address)
			}
		}
	}
}

func (network *P2P_Network) announcePresence() {
	network.UDP.Broadcast(network.tcpServerAddress)
}
