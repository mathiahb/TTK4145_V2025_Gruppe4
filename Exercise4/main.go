package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"
)

const address string = ":10005"
const done_message string = "DONE"
const done_value int = -1

func backup() int {
	UDP_Address, _ := net.ResolveUDPAddr("udp", address)
	connection, _ := net.ListenUDP("udp", UDP_Address)

	// Value being updated by UDP listening
	last_known_counter := 1

	// Backup logic loop
	for {
		connection.SetDeadline(time.Now().Add(time.Second))

		var buf []byte = make([]byte, 512)
		amount_of_bytes, _, err := connection.ReadFromUDP(buf)

		if err != nil {
			connection.Close()
			return last_known_counter
		}

		if string(buf[0:amount_of_bytes]) == done_message {
			return done_value
		}

		last_known_counter, err = strconv.Atoi(string(buf[0:amount_of_bytes]))
	}
}

func primary(counter int) {
	// Spawn Backup
	backup_command_process := exec.Command("gnome-terminal", "--", "go", "run", "./main.go")
	backup_command_process.Run()

	// Create connection to write to
	UDP_Address, _ := net.ResolveUDPAddr("udp", address)
	connection, _ := net.DialUDP("udp", nil, UDP_Address) // Nil assumes local system

	for {
		// Imitate work time
		<-time.NewTimer(time.Millisecond * 500).C

		// Actual work
		fmt.Printf("Counter: %d\n", counter)
		counter++

		// Heartbeat
		message := fmt.Sprintf("%d", counter)
		connection.Write([]byte(message))

		// Termination
		if counter == 100 {
			connection.Write([]byte(done_message))
			return
		}
	}
}

func main() {

	// Assume backup until backup returns.
	counter := backup()
	if counter == done_value {
		return
	}
	primary(counter)
}
