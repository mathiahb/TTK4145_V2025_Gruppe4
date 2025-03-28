package peerToPeer

import (
	"elevator_project/common"
	"fmt"
	"strings"
)

type Dependency struct {
	Dependency_Owner string
	Dependency_Clock LamportClock
}

func NewDependency(owner string, clock LamportClock) Dependency {
	return Dependency{
		Dependency_Owner: owner,
		Dependency_Clock: clock,
	}
}

// Format Dependency:
// DEPENDENCYOWNER/LAMPORTCLOCK
func (dependency Dependency) ToString() string {
	return dependency.Dependency_Owner + common.P2P_DEP_DELIMINATOR + dependency.Dependency_Clock.String()
}

func Dependency_From_String(str string) Dependency {
	fields := strings.Split(str, common.P2P_DEP_DELIMINATOR)

	if len(fields) != 2 {
		fmt.Printf("Error parsing dependency! %s\n", str)
		return Dependency{}
	}

	owner := fields[0]
	time := NewLamportClockFromString(fields[1])

	return Dependency{
		Dependency_Owner: owner,
		Dependency_Clock: time,
	}
}

func (lesser Dependency) Is_Less_Than(greater Dependency) bool {
	return lesser.Dependency_Clock.Is_Less_Than(greater.Dependency_Clock)
}
