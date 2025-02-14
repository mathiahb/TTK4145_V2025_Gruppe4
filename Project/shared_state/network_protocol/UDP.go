package network_protocol

// Implements a UDP access point to the broadcast channel
// Interface:
// initialize() - Returns [send, receive] channels. Warning: Receive blocks?
//			Sets up the access point, should be called by protocol intending to use UDP on the LAN
