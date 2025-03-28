package peerToPeer

import (
	"elevator_project/common"
	"fmt"
	"strconv"
)

type LamportClock struct {
	time int
}

func NewLamportClock() LamportClock {
	return LamportClock{
		time: 0,
	}
}

func NewLamportClockFromString(str string) LamportClock {
	time, err := strconv.Atoi(str)

	if err != nil {
		fmt.Printf("Error in P2P Lamport Clock creation: %s\n", err)
		return NewLamportClock()
	}

	return LamportClock{
		time: time,
	}
}

func (clock LamportClock) String() string {
	return strconv.Itoa(clock.time)
}

// UNUSED?
func (clock LamportClock) GetTime() int {
	return clock.time
}

// Advances the clock after an event.
func (clock *LamportClock) Event() {
	clock.time = clock.time + 1
}

func (receiverClock *LamportClock) Update(senderClock LamportClock) {
	if receiverClock.IsLessThan(senderClock) {
		receiverClock.time = senderClock.time
	}

	receiverClock.Event()
}

// If a -> b, then b.time < a.time
func (lesser LamportClock) IsLessThan(greater LamportClock) bool {
	// Check if the clocks have wrapped around
	const lowerEdge int = common.LAMPORT_CLOCK_WRAPAROUND_LOWER_EDGE
	const upperEdge int = common.LAMPORT_CLOCK_WRAPAROUND_UPPER_EDGE

	if greater.time < lowerEdge && lesser.time > upperEdge {
		return true
	}
	if lesser.time < lowerEdge && greater.time > upperEdge {
		return false
	}

	// Otherwise just check the timers directly.
	return lesser.time < greater.time
}
