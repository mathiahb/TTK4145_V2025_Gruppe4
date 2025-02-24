package peer_to_peer

import (
	"Constants"
	"Network-Protocol/TCP"
	"Network-Protocol/UDP"
	"math/rand"
	"time"
)

// Package peer_to_peer
//
// Handles Peer Detection over UDP and Communication over TCP
// Does not handle elevator detection (not all peers must be elevators, some may be listeners!)

type P2P_Network struct {
	Read_Channel chan P2P_Message

	TCP *TCP.TCP_Connection_Manager
	UDP UDP.UDP_Channel

	tcp_server_address string

	message_controller *Message_Controller
	dependency_handler Dependency_Handler
	clock              Lamport_Clock
}

func New_P2P_Network(server_port string) *P2P_Network {
	read_channel := make(chan P2P_Message, Constants.P2P_BUFFER_SIZE)

	tcp_manager := TCP.New_TCP_Connection_Manager()
	udp_channel := UDP.New_UDP_Channel()

	server_address := UDP.Get_local_IP().String() + ":" + server_port

	network := P2P_Network{
		Read_Channel: read_channel,

		TCP: tcp_manager,
		UDP: udp_channel,

		tcp_server_address: server_address,

		message_controller: New_Message_Controller(),
		dependency_handler: New_Dependency_Handler(),
		clock:              New_Lamport_Clock(),
	}

	tcp_manager.Open_Server(server_port)
	time.Sleep(Constants.TCP_SERVER_BOOTUP_TIME) // Let the TCP server start up first.

	go network.peer_detection()
	go network.reader()

	return &network
}

func (network *P2P_Network) Close() {
	network.TCP.Close_All()
	network.UDP.Close()
}

func (network *P2P_Network) Broadcast(message P2P_Message) {
	network.message_controller.Emplace_New_Message(message)
	network.TCP.Broadcast(message.To_String())
}

func (network *P2P_Network) Send(message P2P_Message, recipient string) {
	network.TCP.Send(message.To_String(), recipient)
}

func (network *P2P_Network) Create_Message(message string, message_type P2P_Message_Type) P2P_Message {
	network.clock.Event()

	return New_P2P_Message(network.tcp_server_address, message_type, network.clock, message)
}

func (network *P2P_Network) reader() {
	for {
		tcp_message := <-network.TCP.Global_Read_Channel
		p2p_message := P2P_Message_From_String(tcp_message)

		network.clock.Update(p2p_message.Time)
		go network.publisher(p2p_message)
	}
}

func (network *P2P_Network) publisher(message P2P_Message) {
	new_dependency := New_Dependency(message.Sender, message.Time)
	if network.dependency_handler.Has_Dependency(new_dependency) {
		// Already received this message, discard.
		return
	}
	network.dependency_handler.Add_Dependency(new_dependency)

	timeout := time.NewTimer(time.Second)

	// Wait for message dependency, then publish onto read channel.
	for {
		select {
		case <-timeout.C:
			return // Timed out.
		default:
		}

		if network.dependency_handler.Has_Dependency(message.dependency) {
			if message.Type == MESSAGE {
				network.Read_Channel <- message
			} else {
				network.handle_special_case(message)
			}

			return
		}

		// Request the missing dependency.
		network.request_dependency(message.dependency)

		// Wait until more data is processed
		time.Sleep(20 * time.Millisecond)
	}
}

func (network *P2P_Network) request_dependency(dependency Dependency) {
	message := network.Create_Message(dependency.To_String(), REQUEST_MISSING_DEPENDENCY)
	network.Send(message, dependency.Dependency_Owner)
}

// An UDP Port is limited in availability should the elevators be on the same machine / IP.
// Try to open a server, keep it alive for server lifetime defined in Constants.
// If not opened, try again after a random delay.
//
// TODO: Separate the function into more readable parts
func (network *P2P_Network) peer_detection() {
	renew_presence_ticker := time.NewTicker(Constants.UDP_WAIT_BEFORE_TRANSMITTING_AGAIN)

	timer_until_looking_for_peers := time.NewTimer(Constants.UDP_UNTIL_SERVER_BOOT)
	timer_until_stopping := time.NewTimer(Constants.UDP_SERVER_LIFETIME)
	timer_until_stopping.Stop() // Server starts closed

	server_open := false

	for {
		select {
		case <-renew_presence_ticker.C:
			network.announce_presence()

		case <-timer_until_looking_for_peers.C:
			err := network.UDP.Start_Reading()

			if err == nil {
				server_open = true
				timer_until_stopping.Reset(Constants.UDP_SERVER_LIFETIME)
			} else {
				// Someone else is using this resource, wait a random amount of time!
				timer_until_looking_for_peers.Reset(time.Duration(1 + rand.Intn(int(Constants.UDP_UNTIL_SERVER_BOOT))))
			}

		case <-timer_until_stopping.C:
			network.UDP.Stop_Reading()
			server_open = false
			timer_until_looking_for_peers.Reset(Constants.UDP_UNTIL_SERVER_BOOT)

		default:
			if server_open {
				network.detect_and_connect_to_peers()
			}
		}
	}
}

func (network *P2P_Network) announce_presence() {
	network.UDP.Broadcast(network.tcp_server_address)
}

func (network *P2P_Network) detect_and_connect_to_peers() {
	select {
	case address := <-network.UDP.Read_Channel:
		if !network.TCP.Does_Connection_Exist(address) {
			network.TCP.Connect_Client(address)
		}
	default:
		return
	}
}
