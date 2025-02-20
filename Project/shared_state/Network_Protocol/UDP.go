package Network_Protocol

// Implements a UDP access point to the broadcast channel
// Interface:
// initialize() - Returns [send, receive] channels. Warning: Receive blocks?
//			Sets up the access point, should be called by protocol intending to use UDP on the LAN

import (
	"net"
	"time"

	"github.com/mathiahb/TTK4145_V2025_Gruppe4/Constants"
)

// ----------------------------------------------------------------

type UDP_Channel struct {
	Write_Channel chan string
	Read_Channel  chan string
}

// --------------------------------------------------------------------

func Get_local_IP() net.IP {
	connection, err := net.Dial("udp", "0.0.0.0:10")
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	localaddr := connection.LocalAddr().(*net.UDPAddr)

	return localaddr.IP
}

// func create_UDP_client, to be called as a go-routine by a function that creates an UDP_Channel
func create_UDP_client(channel_write chan string) {
	addr, err := net.ResolveUDPAddr("udp", Constants.UDP_BROADCAST_IP_PORT)
	if err != nil {
		panic(err)
	}

	connection, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	// Handle incoming write requests
	for {
		var message string = <-channel_write

		//connection.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
		data := []byte(message)
		connection.Write(data)
	}
}

// func create_UDP_server, to be called by a function that creates an UDP_Channel
func create_UDP_server(channel_read chan string) {
	addr, err := net.ResolveUDPAddr("udp", Constants.UDP_PORT) // Listen to the broadcast port.
	if err != nil {
		panic(err)
	}

	connection, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	for {
		deadline := time.Now().Add(Constants.UDP_READ_DEADLINE)
		connection.SetReadDeadline(deadline)

		data := make([]byte, 1024)
		bytes_received, sender_address, err := connection.ReadFromUDP(data)

		if err == nil && sender_address != connection.LocalAddr() {
			message := string(data[0:bytes_received])
			channel_read <- message
		}

		<-time.NewTimer(Constants.UDP_WAIT_BEFORE_READING_AGAIN).C
	}
}

func Setup_UDP_Broadcast() UDP_Channel {
	channel_write := make(chan string, 1024)
	channel_read := make(chan string, 1024)

	go create_UDP_client(channel_write)
	go create_UDP_server(channel_read)

	return UDP_Channel{
		Write_Channel: channel_write,
		Read_Channel:  channel_read,
	}
}
