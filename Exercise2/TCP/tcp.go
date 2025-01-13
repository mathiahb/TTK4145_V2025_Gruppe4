package main

import (
	"fmt"
	"net"
)

func sender() {

}

func receiver() {

	listener, error := net.Listen("tcp", "localhost:30000")

	fmt.Print(error)

	connection, _ := listener.Accept()

	buffer := make([]byte, 1025) // Max 1024 bytes, last byte used for null termination

	var ourAddress net.Addr = connection.LocalAddr()
	var fromWho net.Addr

	for {
		numBytesReceived, _ := connection.Read(buffer)
		fromWho = connection.RemoteAddr()

		if ourAddress != fromWho {
			if numBytesReceived >= 1025 {
				fmt.Print("Buffer overflow!")
			} else {
				buffer[numBytesReceived] = 0 // Null byte for string termination
				fmt.Print(buffer)
				fmt.Print(fromWho)
			}
		}
	}
}

func main() {

}
