package main

import (
	"fmt"
	"net"
	"net/netip"
	"time"
)

func sender() {
	//Address, _ := netip.ParseAddrPort("10.100.23.204:20005")
	//UDPAddress := net.UDPAddrFromAddrPort(Address)
	Address, _ := netip.ParseAddrPort("10.100.23.255:20005")
	UDPAddress := net.UDPAddrFromAddrPort(Address)

	connection, error := net.DialUDP("udp4", nil, UDPAddress) //Broadcast address

	if error != nil {
		panic(error)
	}

	for {
		connection.Write([]byte("Hello, world!"))

		time.Sleep(time.Second)
	}
}

func receiver() {

	UDPAddress, _ := net.ResolveUDPAddr("udp4", ":20005")

	connection, error := net.ListenUDP("udp4", UDPAddress)

	if error != nil {
		panic(error)
	}

	buffer := make([]byte, 1024) // Max 1024 bytes, last byte used for null termination

	var ourAddress net.Addr = connection.LocalAddr()

	for {
		numBytesReceived, fromWho, _ := connection.ReadFromUDP(buffer)

		if ourAddress != fromWho {
			if numBytesReceived >= 1024 {
				fmt.Print("Buffer overflow!")
			} else {
				fmt.Print(string(buffer), "\n")
				fmt.Print(fromWho, "\n")
			}
		}
		fmt.Print(string(buffer), "\n")
		fmt.Print(fromWho, "\n")
	}
}

func main() {
	go receiver()
	go sender()

	select {}
}
