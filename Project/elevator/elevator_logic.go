package elevator

import (
	elevio "Driver-Elevio"
	"shared_states" //hvordan koble ting sammen?
	"fmt"
)

// getElevatorID returns a unique identifier for this elevator instance.
func getElevatorID() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return "unknown_elevator"
	}
	return hostname
}


func ElevatorUninitialized() Elevator { // Initialize a new elevator with default values
	turnOffAllLights()
	return Elevator{
		Floor:       -1,
		Dirn:        D_Stop,
		Behaviour:   EB_Idle,
		CabRequests: make([]bool, N_FLOORS),
	}
}

func turnOffAllLights(){
	//starter med alle lys avslått
	for button := 0; button < N_BUTTONS; button++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, false)
		}
	}
	elevio.SetDoorOpenLamp(false)
}


// må legge til alle kanalene
func elevator(elevatorStateChannel chan Elevator, newHallRequestChannel chan HallRequestsType,  approvedHallRequestChannel chan HallRequestsType, clearHallRequestChannel chan HallRequestsType, clearCabRequestChannel chan Elevator) {
	
	//føler at det er litt initialisering/konfigurering som mangler

	var localElevator = ElevatorUninitialized() // lager et lokalt heisobjekt
	var hallRequests = HallRequestsUninitialized() // lager et tomt request-objekt
	var isObstructed = false

	localElevator = InitFSM(elevatorStateChannel, localElevator) // shared state får vite at en heis eksisterer, kjenner ikke helt til poenget med resten av funksjonen
	if (inputDevice.FloorSensor() == -1) { // fase ut inputDevice?
		localElevator = FSMOnInitBetweenFloors(localElevator, elevatorStateChannel)
	}
	
	for {
		select{
		
		case buttonEvent := <-buttonChannel: 
			localElevator = FSMButtonPress(buttonEvent.btnFloor,  buttonEvent.btnType, localElevator, elevatorStateChannel, newHallRequestChannel)
		
		case newFloor := <-floorsChannel:
			localElevator, hallRequests = FSMOnFloorArrival(newFloor, localElevator, hallRequests, clearHallRequestChannel, clearCabRequestChannel, elevatorStateChannel)

		case isObstructed = <- obstructionChannel: // dette føles rart, men er nødt til å vite om noen prøver å obstructe døren
			
			if(isObstructed){

			}
		
		case <- doorTimerIsUp:

			if(!isObstructed){
				localElevator = FSMCloseDoors(localElevator, hallRequests, elevatorStateChannel)
			}

		case hallRequests <- approvedHallRequestChannel: // fordi alle ordre kommer fra shared states
			localElevator = FSMStartMoving(localElevator, hallRequests, elevatorStateChannel)
			
		}
	}
}

