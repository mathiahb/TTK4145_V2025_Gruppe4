package elevator_logic

import (
	"Driver-go/elevio"
	"fmt"
)

type Order struct {
	UP     [4]int
	DOWN   [4]int
	INSIDE [4]int
}

const (
	STATE_MOVING_UP   = iota
	STATE_MOVING_DOWN = iota
	STATE_IDLE        = iota
	STATE_UNDEFINED   = iota

	//STATE_STANDING_STILL_MOVING_UP   = iota
	//STATE_STANDING_STILL_MOVING_DOWN = iota
)

func addOrder(bEvent elevio.ButtonEvent, order *Order) {
	switch bEvent.Button {
	case elevio.BT_HallUp:
		order.UP[bEvent.Floor] = 1
	case elevio.BT_HallDown:
		order.DOWN[bEvent.Floor] = 1
	case elevio.BT_Cab:
		order.INSIDE[bEvent.Floor] = 1
	}
	elevio.SetButtonLamp(bEvent.Button, bEvent.Floor, true)
}

func removeOrder(bEvent elevio.ButtonEvent, order *Order) {
	switch bEvent.Button {
	case elevio.BT_HallUp:
		order.UP[bEvent.Floor] = 0
	case elevio.BT_HallDown:
		order.DOWN[bEvent.Floor] = 0
	case elevio.BT_Cab:
		order.INSIDE[bEvent.Floor] = 0
	}
	elevio.SetButtonLamp(bEvent.Button, bEvent.Floor, false)

}

func getMotorDirectionToFloor(currentFloor int, destinationFloor int) elevio.MotorDirection {
	if currentFloor < destinationFloor {
		return elevio.MD_Up
	} else if currentFloor > destinationFloor {
		return elevio.MD_Down
	} else {
		return elevio.MD_Stop
	}
}

func chooseDirection(state *int, order *Order, currentFloor int) elevio.MotorDirection {
	if *state == STATE_MOVING_UP {
		for f := currentFloor + 1; f < 4; f++ {
			if order.UP[f] == 1 || order.INSIDE[f] == 1 {
				*state = STATE_MOVING_UP
				return elevio.MD_Up
			}
		}
		for f := 3; f >= 0; f-- {
			if order.DOWN[f] == 1 || order.INSIDE[f] == 1 {
				*state = STATE_MOVING_DOWN
				return getMotorDirectionToFloor(currentFloor, f)
			}
		}
	} else if *state == STATE_MOVING_DOWN {
		for f := currentFloor - 1; f >= 0; f-- {
			if order.DOWN[f] == 1 || order.INSIDE[f] == 1 {
				*state = STATE_MOVING_DOWN
				return elevio.MD_Down
			}
		}
		for f := 0; f < 4; f++ {
			if order.UP[f] == 1 || order.INSIDE[f] == 1 {
				*state = STATE_MOVING_UP
				return getMotorDirectionToFloor(currentFloor, f)
			}
		}
	}

	if *state != STATE_UNDEFINED{
		for f := 0; f < 4; f++ {
			if order.UP[f] == 1 || order.DOWN[f] == 1 || order.INSIDE[f] == 1 {
				direction := getMotorDirectionToFloor(currentFloor, f)
				if direction == elevio.MD_Up {
					*state = STATE_MOVING_UP
				} else if direction == elevio.MD_Down {
					*state = STATE_MOVING_DOWN
				}
				return direction
			}
		}
	}

	*state = STATE_IDLE
	return elevio.MD_Stop

}

func Init_elevator_logic(numFloors int, d elevio.MotorDirection) {

	var state int = STATE_UNDEFINED
	var order Order = Order{[4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0}}
	var currentFloor int

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	elevio.SetMotorDirection(elevio.MD_Down)

	currentFloor = <-drv_floors // Wait for the elevator to reach a floor before starting the elevator logic

	state = STATE_IDLE

	for {
		select {
		case bEvent := <-drv_buttons:
			addOrder(bEvent, &order)
			fmt.Printf("%+v\n", order)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
			elevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}

	select {}

}
