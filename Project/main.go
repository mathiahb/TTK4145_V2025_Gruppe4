package main

import (
	"fmt"
	"time"

	"github.com/mathiahb/TTK4145_V2025_Gruppe4/shared_state/Network_Protocol"
)

func listener(connection Network_Protocol.TCP_Connection, kill chan bool) {
	for {
		select {
		case message := <-connection.Read_Channel:
			fmt.Printf("%s received message: %s\n", connection.Get_Name(), message)
		case <-kill:
			return
		}
	}
}

func sender(connection Network_Protocol.TCP_Connection, kill chan bool) {
	ticker_send := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker_send.C:
			connection.Write_Channel <- connection.Get_Name()
		case <-kill:
			return
		}
	}
}

func main() {

	var connection_manager Network_Protocol.TCP_Connection_Manager = *Network_Protocol.New_TCP_Connection_Manager()

	connection_manager.Open_Server()

	<-time.After(time.Second)
	connection_manager.Connect_Client("10.24.51.1")

	<-time.After(time.Second * 5)

	timer_kill := time.NewTimer(time.Minute)
	kill_channel := make(chan bool)

	for _, connection := range connection_manager.Connections {
		go listener(connection, kill_channel)
		go sender(connection, kill_channel)

		fmt.Printf("Added %s\n", connection.Get_Name())
	}

	<-timer_kill.C
	close(kill_channel)
}
