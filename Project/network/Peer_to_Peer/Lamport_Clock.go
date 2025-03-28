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

func (receiver_clock *LamportClock) Update(sender_clock LamportClock) {
	if receiver_clock.Is_Less_Than(sender_clock) {
		receiver_clock.time = sender_clock.time
	}

	receiver_clock.Event()
}

// If a -> b, then b.time < a.time
func (lesser LamportClock) Is_Less_Than(greater LamportClock) bool {
	// Check if the clocks have wrapped around
	const lower_edge int = common.LAMPORT_CLOCK_WRAPAROUND_LOWER_EDGE
	const upper_edge int = common.LAMPORT_CLOCK_WRAPAROUND_UPPER_EDGE

	if greater.time < lower_edge && lesser.time > upper_edge {
		return true
	}
	if lesser.time < lower_edge && greater.time > upper_edge {
		return false
	}

	// Otherwise just check the timers directly.
	return lesser.time < greater.time
}
