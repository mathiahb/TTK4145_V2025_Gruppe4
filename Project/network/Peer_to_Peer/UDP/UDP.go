package UDP

import (
	"elevator_project/common"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/ipv4"
)

// Package UDP
//-----------------------------------------------------------------
//
// Implements a struct UDP_Channel, connected to Multicast (239.255.255.255), made by NewUDPChannel().
//
// WRITING:
// By default on
// Use Channel.Write_Channel <- message, to write onto the channel
//
// READING:
// Use message <- Channel.ReadChannel, to read from the channel after Start_Reading() has been called
//
// Note: There may be messages read in ReadChannel after Stop_Reading has been called.
//
// ----------------------------------------------------------------

type UDPChannel struct {
	// Public
	WriteChannel chan string
	ReadChannel  chan string

	// Protected
	quitChannel chan bool
}

func (channel UDPChannel) Close() {
	close(channel.quitChannel)
}

func (channel UDPChannel) Broadcast(message string) {
	channel.WriteChannel <- message
}

// --------------------------------------------------------------------

func GetLocalIP() net.IP {
	connection, err := net.Dial("udp", "255.255.255.255:1")
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	localaddr := connection.LocalAddr().(*net.UDPAddr)

	return localaddr.IP
}

func (Channel UDPChannel) udpClient(connection *net.UDPConn) {
	defer connection.Close()

	// Handle incoming write requests
	for {
		select {
		case message := <-Channel.WriteChannel:
			data := []byte(message)
			connection.Write(data)
		case <-Channel.quitChannel:
			return
		}

		time.Sleep(time.Microsecond)
	}
}

func (Channel UDPChannel) udpServer(connection *ipv4.PacketConn) {
	defer connection.Close()

	tickerReadUDP := time.NewTicker(common.UDP_WAIT_BEFORE_READING_AGAIN)

	for {
		select {
		case <-tickerReadUDP.C:
			deadline := time.Now().Add(common.UDP_READ_DEADLINE)
			connection.SetReadDeadline(deadline)

			data := make([]byte, 1024)
			bytesReceived, _, _, err := connection.ReadFrom(data)

			if err == nil {
				message := string(data[0:bytesReceived])
				Channel.ReadChannel <- message
			}
		case <-Channel.quitChannel:
			return
		}

		time.Sleep(time.Microsecond) // Needed to make time advance in the VM so that the channels can close.
	}
}

func (channel *UDPChannel) createUDPClient() {
	addr, err := net.ResolveUDPAddr("udp", common.UDP_BROADCAST_IP_PORT)
	if err != nil {
		panic(err)
	}

	connection, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	go channel.udpClient(connection)
}

// func createUDPServer, to be called by a function that creates an UDP_Channel
func (channel *UDPChannel) createUDPServer() error {
	addr, err := net.ResolveUDPAddr("udp", common.UDP_BROADCAST_IP_PORT)
	if err != nil {
		panic(err)
	}

	connection, err := net.ListenPacket("udp4", common.UDP_BROADCAST_IP_PORT)
	if err != nil {
		return err
	}

	packetConn := ipv4.NewPacketConn(connection)

	nifi, err := net.Interfaces()
	if err != nil {
		return err
	}

	err = packetConn.SetMulticastLoopback(true)
	if err != nil {
		return err
	}

	for _, ifi := range nifi {
		packetConn.JoinGroup(&ifi, addr)
	}

	go channel.udpServer(packetConn)

	return nil
}

func NewUDPChannel() UDPChannel {
	channelWrite := make(chan string, 1024)
	channelRead := make(chan string, 1024)
	channelQuit := make(chan bool)

	channel := UDPChannel{
		WriteChannel: channelWrite,
		ReadChannel:  channelRead,
		quitChannel:  channelQuit,
	}

	channel.createUDPClient()
	err := channel.createUDPServer()
	if err != nil {
		fmt.Printf("Error when starting UDP server: %s\n", err.Error())
	}

	return channel
}
