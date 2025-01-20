package main

import (
	"Driver-go/elevator_logic"
	"Driver-go/elevio"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	elevator_logic.Init_elevator_logic(numFloors, d)

	select {}

}
