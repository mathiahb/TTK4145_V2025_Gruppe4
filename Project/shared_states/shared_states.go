package shared_states

// hvordan koble sammen elevator og nettverk mtp importering etc.?

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

/*
Hvordan shared state skal kommunisere med heis og nettverk

-> upsis, ikke oppdatert! mangler clearHalRequests og clearCabRequests...

a) heisen får en ny hallRequest (knapp har blitt trykket)
	1. si ifra til nettverk: notifyNewHallRequestChannel // fra shared_state til nettverk

b) Nettverket godkjenner ny HallRequest
	1. oppdatere HRAInputVariable
	2. sende HRAInput til getHallRequestAssignments
	3. sende hallRequestAssignments til heisen (elevator)

c) heisen har en ny tilstand
	1. si ifra til nettverk: informNewStateChannel

d) Nettverket informerer om ny tilstand
	1. oppdatere HRAInputVariable
	2. sende HRAInput til getHallRequestAssignments
	3. sende hallRequestAssignments til heisen (elevator)

e) Dersom nettverket ønsker å starte synkronisering
	1. Alle noder som er koblet til nettverket deler shared state
	2. Nettverket sender oppdatert state tilbake
	3. Hver shared state kaller på HRA

Obs! med denne implementasjonen operere en heis som er disconnected fra resten av nettverket som en enkelt enhet og bruker HRA alene helt fram til den kobles til.
Spm: hva skjer dersom heisen kobles fra heis, altså at den mister informasjon om sin egen tilstand. Håndteres det av nettverket? Hvordan bør shared state funke i så tilfelle?

*/



type HRAType struct { 
	HallRequests HallRequestsType    `json:"hallRequests"`
	States       map[string]Elevator `json:"states"`
}

type HallRequestType struct {
	floor int
	button elevio.ButtonType
}

// ===================== REQUEST ASSIGNER ===================== //

func getHallRequestAssignments(HRAInputVariable HRAType) map[string][][2]bool {

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
// Bridge between network and the elevator. The shared states communicates also with the HRA.

func sharedState(elevatorStateChannel chan Elevator, newHallRequestChannel chan HallRequestType, approvedHallRequestChannel  chan HallRequestType, startSynchChannel chan struct{}, updatedSharedStateForSynchChannel chan HRAType, sendSharedStateForSynchChannel chan HRAType, notifyNewHallRequestChannel  chan HallRequestType, approvedNewHallRequestChannel chan HallRequestType, informNewStateChannel chan Elevator, informedNewStateChannel chan Elevator){

	var HRAInputVariable HRAType 
	var localID = getElevatorID() //denne funksjonen eksisterer i FSM


	for{
		select{

			case newHallRequest := <- newHallRequestChannel: // knapp trykket på lokal heis
				notifyNewHallRequestChannel  <- newHallRequest // be om godkjennelse fra nettverk

			case approvedRequest := <- approvedNewHallRequestChannel: // endring godkjent av nettverk
				HRAInputVariable.HallRequests[newHallRequest.floor][newHallRequest.button] = true // oppdaterer hall requests basert på lokal heis
				approvedHallRequestChannel  <- getHallRequestAssignments(HRAInputVariable) // Be om ny oppdragsfordeling og sende til lokal heis

			case newElevatorState := <- elevatorStateChannel: // tilstand endret på lokal heis, samme logikk som over
				informNewStateChannel <- newElevatorState // kan formatere selv, men må være en streng!!
				
			case approvedElevatorState := <- approvedNewelevatorStateChannel : 
				HRAInputVariable.States[localID] = approvedElevatorState 
				approvedHallRequestChannel  <- getHallRequestAssignments(HRAInputVariable)

			case clearCabRequest := <- clearCabRequestChannel:  // hmm er denne nødvendig da? slik den står nå er den litt misvisende
				informNewStateChannel <- clearCabRequest // 	
			

			case clearHallRequest := <- clearHallRequestChannel:
				approveClearHallRequest <- clearHallRequest

			case <- approvedClearHallRequestsChannel:
				HRAInputVariable.HallRequests[newHallRequest.floor][newHallRequest.button] = false // oppdaterer hall requests basert på lokal heis
				approvedHallRequestChannel  <- getHallRequestAssignments(HRAInputVariable) 

			case <- startSynchChannel: //nettverket ønsker å starte synkronisering
				sendStateForSynchChannel <- HRAInputVariable  // dette må være en json-marshall-streng
			
			case updatedSharedStates := <- updatedSharedStateForSynchChannel //shared states får oppdaterte states fra heiser som er koblet på nettverket
				HRAInputVariable = updatedSharedStates //ups denne kommer som en streng
				approvedHallRequestChannel <- getHallRequestAssignments(HRAInputVariable)
		}
	}
}

