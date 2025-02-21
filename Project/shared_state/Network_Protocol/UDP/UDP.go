package UDP

// Implements a UDP access point to the broadcast channel
// Interface:
// initialize() - Returns [send, receive] channels. Warning: Receive blocks?
//			Sets up the access point, should be called by protocol intending to use UDP on the LAN

import (
	"net"
	"time"

	"Constants"
)

// ----------------------------------------------------------------

type UDP_Channel struct {
	// Public
	Write_Channel chan string
	Read_Channel  chan string

	// Protected
	stop_channel chan bool
}

func (channel UDP_Channel) Start_Reading() error {
	return create_UDP_server(channel.Read_Channel, channel.stop_channel)
}

func (channel UDP_Channel) Stop_Reading() {
	channel.stop_channel <- true
}

func (channel UDP_Channel) Broadcast(message string) {
	channel.Write_Channel <- message
}

// --------------------------------------------------------------------

func Get_local_IP() net.IP {
	connection, err := net.Dial("udp", "8.8.8.8:10")
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	localaddr := connection.LocalAddr().(*net.UDPAddr)

	return localaddr.IP
}

func udp_client(connection net.UDPConn, channel_write chan string) {
	defer connection.Close()

	// Handle incoming write requests
	for {
		var message string = <-channel_write

		//connection.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
		data := []byte(message)
		connection.Write(data)

		time.Sleep(time.Microsecond)
	}
}

func udp_server(connection *net.UDPConn, channel_read chan string, close_channel chan bool) {
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
				channel_read <- message
			}
		case <-close_channel:
			return
		}

		time.Sleep(time.Microsecond) // Needed to make time advance in the VM so that the channels can close.
	}
}

func create_UDP_client(channel_write chan string) {
	addr, err := net.ResolveUDPAddr("udp", Constants.UDP_BROADCAST_IP_PORT)
	if err != nil {
		panic(err)
	}

	connection, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	go udp_client(*connection, channel_write)
}

// func create_UDP_server, to be called by a function that creates an UDP_Channel
func create_UDP_server(channel_read chan string, close_channel chan bool) error {
	addr, err := net.ResolveUDPAddr("udp", Constants.UDP_PORT) // Listen to the broadcast port.
	if err != nil {
		panic(err)
	}

	connection, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	go udp_server(connection, channel_read, close_channel)

	return nil
}

func New_UDP_Channel() UDP_Channel {
	channel_write := make(chan string, 1024)
	channel_read := make(chan string, 1024)
	channel_stop := make(chan bool)

	create_UDP_client(channel_write)

	return UDP_Channel{
		Write_Channel: channel_write,
		Read_Channel:  channel_read,
		stop_channel:  channel_stop,
	}
}
