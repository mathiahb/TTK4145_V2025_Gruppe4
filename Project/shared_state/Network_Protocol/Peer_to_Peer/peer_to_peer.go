package peer_to_peer

import (
	"Network-Protocol/TCP"
	"Network-Protocol/UDP"
)

// Package peer_to_peer
//
// Handles Peer Detection over UDP and Communication over TCP
// Does not handle elevator detection (not all peers must be elevators, some may be listeners!)

type P2P_Network struct {
	TCP TCP.TCP_Connection_Manager
	UDP UDP.UDP_Channel
}
