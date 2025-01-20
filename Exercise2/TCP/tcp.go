package main

import (
	"fmt"
	"net"
	"time"
)

func client() {
	time.Sleep(time.Second)
	connection, _ := net.Dial("tcp", "10.100.23.204:33546")
	defer connection.Close() // Close the connection when the function returns
	tcpConn := connection.(*net.TCPConn)
	tcpConn.SetNoDelay(true)

	buffer := make([]byte, 1024)

	connection.Write(([]byte("Connect to:10.100.23.15:20005\000")))
	for {
		connection.Write([]byte("Hello, world 1!\000"))
		connection.Read(buffer)
		fmt.Println(string(buffer))

		time.Sleep(time.Second)

	}

}

func server() {

	listener, err := net.Listen("tcp", "10.100.23.15:20005")

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	defer listener.Close()             // Close the listener when the function returns
	connection, _ := listener.Accept() // Accept a connection
	defer connection.Close()           // Close the connection when the function returns

	buffer := make([]byte, 1024)

	for {
		connection.Write([]byte("Hello, world 2!\000"))
		connection.Read(buffer)
		fmt.Println(string(buffer))

		time.Sleep(time.Second)

	}
}

func main() {

	go server()
	go client()

	select {} // Block forever

}
