package UDP

import (
	"net"
	"time"

	"Constants"
)

// Package UDP
//-----------------------------------------------------------------
//
// Implements a struct UDP_Channel, connected to Broadcast (255.255.255.255), made by New_UDP_Channel().
//
// WRITING:
// By default on
// Use Channel.Write_Channel <- message, to write onto the channel
//
// READING:
// Use Channel.Start_Reading() to open the server and claim the port.
// Use Channel.Stop_Reading() to close the server and free the port.
// Use message <- Channel.Read_Channel, to read from the channel after Start_Reading() has been called
//
// Note: There may be messages read in Read_Channel after Stop_Reading has been called.
//
// ----------------------------------------------------------------

type UDP_Channel struct {
	// Public
	Write_Channel chan string
	Read_Channel  chan string

	// Protected
	stop_server_channel chan bool
	quit_channel        chan bool
}

func (channel *UDP_Channel) Start_Reading() error {
	channel.stop_server_channel = make(chan bool)
	return channel.create_UDP_server()
}

func (channel UDP_Channel) Stop_Reading() {
	close(channel.stop_server_channel)
}

func (channel UDP_Channel) Close() {
	close(channel.quit_channel)
}

func (channel UDP_Channel) Broadcast(message string) {
	channel.Write_Channel <- message
}

// --------------------------------------------------------------------

func Get_local_IP() net.IP {
	connection, err := net.Dial("udp", "255.255.255.255:1")
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	localaddr := connection.LocalAddr().(*net.UDPAddr)

	return localaddr.IP
}

func (Channel UDP_Channel) udp_client(connection *net.UDPConn) {
	defer connection.Close()

	// Handle incoming write requests
	for {
		select {
		case message := <-Channel.Write_Channel:
			data := []byte(message)
			connection.Write(data)
		case <-Channel.quit_channel:
			return
		}

		time.Sleep(time.Microsecond)
	}
}

func (Channel UDP_Channel) udp_server(connection *net.UDPConn) {
	defer connection.Close()

	ticker_read_UDP := time.NewTicker(Constants.UDP_WAIT_BEFORE_READING_AGAIN)

	for {
		select {
		case <-ticker_read_UDP.C:
			deadline := time.Now().Add(Constants.UDP_READ_DEADLINE)
			connection.SetReadDeadline(deadline)

			data := make([]byte, 1024)
			bytes_received, _, err := connection.ReadFromUDP(data)

			if err == nil {
				message := string(data[0:bytes_received])
				Channel.Read_Channel <- message
			}
		case <-Channel.stop_server_channel:
			return
		case <-Channel.quit_channel:
			return
		}

		time.Sleep(time.Microsecond) // Needed to make time advance in the VM so that the channels can close.
	}
}

func (channel *UDP_Channel) create_UDP_client() {
	addr, err := net.ResolveUDPAddr("udp", Constants.UDP_BROADCAST_IP_PORT)
	if err != nil {
		panic(err)
	}

	connection, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	go channel.udp_client(connection)
}

// func create_UDP_server, to be called by a function that creates an UDP_Channel
func (channel *UDP_Channel) create_UDP_server() error {
	addr, err := net.ResolveUDPAddr("udp", Constants.UDP_PORT) // Listen to the broadcast port.
	if err != nil {
		panic(err)
	}

	connection, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	go channel.udp_server(connection)

	return nil
}

func New_UDP_Channel() UDP_Channel {
	channel_write := make(chan string, 1024)
	channel_read := make(chan string, 1024)
	channel_stop := make(chan bool)
	channel_quit := make(chan bool)

	channel := UDP_Channel{
		Write_Channel:       channel_write,
		Read_Channel:        channel_read,
		stop_server_channel: channel_stop,
		quit_channel:        channel_quit,
	}

	channel.create_UDP_client()

	return channel
}
