package main

import (
	"fmt"
	"os"
	"tests"

	"Constants"
)

func main() {
	var is_testing bool = false
	var id string = ""
	for i, arg := range os.Args {
		if i == 0 {
			continue
		}

		switch arg {
		case Constants.ARGV_TEST:
			is_testing = true
		case Constants.ARGV_LISTENER_ONLY:
		case Constants.ARGV_BACKUP:
		case Constants.ARGV_ELEVATOR_ID:
			id = os.Args[i+1]
		default:
			fmt.Printf("Unknown Arg: %s", arg)
		}
	}

	if id == "" {
		fmt.Println("Error. No id!")
		return
	}

	if is_testing {
		tests.Test_Creating_Connection(id)
		return
	}
}
