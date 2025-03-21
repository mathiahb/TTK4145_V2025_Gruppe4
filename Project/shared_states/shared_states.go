package shared_states

import (
	"encoding/json"
	"fmt"
	. "elevator_project/constants"
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

// ===================== SHARED STATE ===================== //
// Bridge between network and the elevator. The shared states communicates also with the HRA.

func SharedStateThread(betweenElevatorAndSharedStatesChannels BetweenElevatorAndSharedStatesChannels){

	var HRAInputVariable HRAType 
	var localID = getElevatorID() // denne funksjonen eksisterer i FSM
	var aliveNodes = []string // ser ikke poenget med denne


	for{
		select{

			case newHallRequest := <- newHallRequestChannel: // knapp trykket på lokal heis
				translatedNewHallRequest := translateHallRequestToNetwork(newHallRequest)
				notifyNewHallRequestChannel  <- translatedNewHallRequest // be om godkjennelse fra nettverk

			case approvedRequest := <- approvedNewHallRequestChannel: // endring godkjent av nettverk
				// translate from network
				HRAInputVariable.HallRequests[newHallRequest.floor][newHallRequest.button] = true // oppdaterer hall requests basert på lokal heis
				approvedHallRequestChannel  <- getHallRequestAssignments(HRAInputVariable) // Be om ny oppdragsfordeling og sende til lokal heis

			case newElevatorState := <- elevatorStateChannel: // tilstand endret på lokal heis, samme logikk som over
				translatedNewElevatorState := translateElevatorStateToNetwork(newElevatorState)
				informNewStateChannel <- translatedNewElevatorState 
				
			case approvedElevatorState := <- approvedNewElevatorStateChannel: 
				HRAInputVariable.States[localID] = approvedElevatorState // localID???
				approvedHallRequestChannel  <- getHallRequestAssignments(HRAInputVariable)

			case clearCabRequest := <- clearCabRequestChannel: // må dette være en egen kanal?
				translatedClearCabRequest := translateElevatorStateToNetwork(clearCabRequest)
				informNewStateChannel <- translatedClearCabRequest   	
			
			case clearHallRequest := <- clearHallRequestChannel: // fra elevator til shared state
				translatedClearHallRequest := translateHallRequestToNetwork(clearHallRequest)
				approveClearHallRequest <- translatedClearHallRequest

			case <- approvedClearHallRequestsChannel:
				// må oversette fra nettverket
				HRAInputVariable.HallRequests[newHallRequest.floor][newHallRequest.button] = false // oppdaterer hall requests basert på lokal heis
				approvedHallRequestChannel  <- getHallRequestAssignments(HRAInputVariable) 

			case <- startSynchChannel: //nettverket ønsker å starte synkronisering
				translatedHRAInputVariable := translateHRAToNetwork(HRAInputVariable)
				sendStateForSynchChannel <- translatedHRAInputVariable  // dette må være en json-marshall-streng
			
			case updatedSharedStates := <- updatedSharedStateForSynchChannel //shared states får oppdaterte states fra heiser som er koblet på nettverket
				HRAInputVariable = updatedSharedStates //ups denne kommer som en streng
				approvedHallRequestChannel <- getHallRequestAssignments(HRAInputVariable)
		}
	}
}

