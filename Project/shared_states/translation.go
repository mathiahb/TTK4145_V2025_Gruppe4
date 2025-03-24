package shared_states

import (
	. "elevator_project/constants"
	"encoding/json"
	"fmt"
	"os/exec"
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
	ret, err := exec.Command("./hall_request_assigner/hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
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

func translateToNetwork(variable any) string {

	translatedVariable, err := json.Marshal(variable)

	if err != nil {
		fmt.Println("json.Marshal error:", err)
		return ""
	}

	return string(translatedVariable)
}

// ===================== TRANSLATION FROM NETWORK ===================== //

func translateFromNetwork[T any](variable string) T {
	var translatedVariable T
	err := json.Unmarshal([]byte(variable), &translatedVariable)
	if err != nil {
		fmt.Println("json.Unmarshal error:", err)
		//returnere hva?
	}
	return translatedVariable
}
