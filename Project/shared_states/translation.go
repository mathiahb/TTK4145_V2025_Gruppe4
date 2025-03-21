package shared_states

import (
	"encoding/json"
	"fmt"
	"os/exec"
	. "elevator_project/constants"
)
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

// ===================== TRANSLATION TO NETWORK ===================== //
func translateHallRequestToNetwork(hallRequest HallRequestType)  string { 

	translatedHallRequest, err := json.Marshal(hallRequest)
	
	if err != nil {
		fmt.Println("json.Marshal error:", err)
		return ""
	}
	return string(translatedHallRequest)
}


func translateElevatorStateToNetwork(elevator Elevator) string{
	
	translatedElevator, err := json.Marshal(elevator)

	if err != nil {
		fmt.Println("json.Marshal error:", err)
		return ""
	}

	return string(translatedElevator)
}

func translateHRAToNetwork(HRAInputVariable HRAType) string{
	
	translatedHRA, err := json.Marshal(HRAInputVariable)

	if err != nil {
		fmt.Println("json.Marshal error:", err)
		return ""
	}

	return string(translatedHRA)
}

func translateToNetwork(variable any) string{

	translatedVariable, err := json.Marshal(variable)

	if err != nil {
		fmt.Println("json.Marshal error:", err)
		return ""
	}

	return string(translatedVariable)
}


// ===================== TRANSLATION FROM NETWORK ===================== //
func translateFromNetwork(){

}