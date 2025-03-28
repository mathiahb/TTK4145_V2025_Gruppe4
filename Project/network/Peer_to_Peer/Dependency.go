package peerToPeer

import (
	"elevator_project/common"
	"fmt"
	"strings"
)

type Dependency struct {
	DependencyOwner string
	DependencyClock LamportClock
}

func NewDependency(owner string, clock LamportClock) Dependency {
	return Dependency{
		DependencyOwner: owner,
		DependencyClock: clock,
	}
}

// Format Dependency:
// DEPENDENCYOWNER/LAMPORTCLOCK
func (dependency Dependency) ToString() string {
	return dependency.DependencyOwner + common.P2P_DEP_DELIMINATOR + dependency.DependencyClock.String()
}

func DependencyFromString(str string) Dependency {
	fields := strings.Split(str, common.P2P_DEP_DELIMINATOR)

	if len(fields) != 2 {
		fmt.Printf("Error parsing dependency! %s\n", str)
		return Dependency{}
	}

	owner := fields[0]
	time := NewLamportClockFromString(fields[1])

	return Dependency{
		DependencyOwner: owner,
		DependencyClock: time,
	}
}

func (lesser Dependency) IsLessThan(greater Dependency) bool {
	return lesser.DependencyClock.IsLessThan(greater.DependencyClock)
}
