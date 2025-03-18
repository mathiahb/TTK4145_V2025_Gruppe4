package shared_states

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

/*
Hvordan shared state skal kommunisere med heis og nettverk

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

-spm: høres dette forløpet logisk ut? Er dette slik nettverket bør kobles på?
-spm: hva synes dere om navngigningen til kanalene. Brukte lang tid på å reflektere rundt hva som er mest logisk, og dette er mitt forlag. Si ifra hvis noe er utydelig.

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

			case newElevatorState := <- elevatorStateChannel : // tilstand endret på lokal heis, samme logikk som over
				informNewStateChannel <- newElevatorState 
				
			case approvedElevatorState := <- approvedNewelevatorStateChannel : 
				HRAInputVariable.States[localID] = approvedElevatorState 
				approvedHallRequestChannel  <- getHallRequestAssignments(HRAInputVariable)
		
			case <- startSynchChannel: //nettverket ønsker å starte synkronisering
				sendStateForSynchChannel <- HRAInputVariable 
			
			case updatedSharedStates := <- updatedSharedStateForSynchChannel //shared states får oppdaterte states fra heiser som er koblet på nettverket
				HRAInputVariable = updatedSharedStates
				approvedHallRequestChannel <- getHallRequestAssignments(HRAInputVariable)
		}
	}
}

