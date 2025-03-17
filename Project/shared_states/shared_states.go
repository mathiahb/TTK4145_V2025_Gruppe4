package shared_states

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

/*

struct Shared_State
 - Contains the shared state
 - Contains alive nodes

 - Accepts updates & synchronization result from Network via channel made by main
 - Informs HRA about new shared state	   (State information about alive nodes)
 - Informs elevator about new shared state (lights)

 - Defers updates from elevator to network (which will start an update protocol sequence which should end in an update to shared state.)
*/

// ===================== REQUEST ASSIGNER ===================== //


type HRAInput struct { 
	HallRequests HallRequestsType    `json:"hallRequests"`
	States       map[string]Elevator `json:"states"`
}

type NewHallRequest struct {
	floor int
	button elevio.ButtonType
}

// Kanaler for å sende og motta oppdrag
var AssignRequestChannel = make(chan struct{})            // Signaler for ny oppdragsfordeling
var AssignResultChannel = make(chan map[string][][2]bool) // Resultat fra assigner



// Calls `hall_request_assigner` to get updated assignments for all elevators
func getHallRequestAssignments(HRAInputVariable HRAInput) map[string][][2]bool {
	// Fetch latest state
	

	// Convert to JSON
	jsonBytes, err := json.Marshal(HRAInputVariable) 
	if err != nil {
		fmt.Println("json.Marshal error:", err)
		return nil
	}

	// Call `hall_request_assigner`
	ret, err := exec.Command("../hall_request_assigner/hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error:", err)
		fmt.Println(string(ret))
		return nil
	}

	// Debugging: Print raw response
	fmt.Println("Raw response from hall_request_assigner:", string(ret))

	// Parse JSON response
	output := make(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error:", err)
		return nil
	}

	return output
}




// ===================== SHARED STATE ===================== //
// SharedState keeps track of all elevator states and hall requests.

// Handles Elevator and Hall request assigner communication
func sharedState(newHallRequestChannel chan NewHallRequest, elevatorStateChannel chan Elevator, hallRequestChannel chan HRAOutput){

	var HRAInputVariable HRAInput
	
	for{
		select{

		case newHallRequest := <- newHallRequestChannel: // knapp trykket
			
			// oppdaterer hall requests basert på lokal heis
			HRAInputVariable.HallRequests[newHallRequest.floor][newHallRequest.button] = true 
			
			// Be om ny oppdragsfordeling og sende til lokal heis
			hallRequestChannel <- getHallRequestAssignments(HRAInputVariable)
			


		case elevatorState := <- elevatorStateChannel:
			
			HRAInputVariable.States[getElevatorID()] = elevatorState // feilhåndtering, finnes vår egen heis?

			// Be om ny oppdragsfordeling og sende til lokal heis
			hallRequestChannel <- getHallRequestAssignments(HRAInputVariable) // i Atle sin versjon ble det satt opp en egen kanal for å kalle på getHallRew.... Hvorfor?

		}

	}
}




// Variabler som tilhører sentralen
/*
- Liste over alle hall requests (knapper som er trykket - ikke cab)
- Liste over alle heistilstander (liste av Elevatir)
-
*/



// type SharedState struct {
// 	// alive nodes on network
// 	Elevator
//}





