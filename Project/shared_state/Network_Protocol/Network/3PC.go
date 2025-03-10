package network

import peer_to_peer "Network-Protocol/Peer_to_Peer"

type Command struct{
	Field string
	New_Value string
}

// Jeg har en endring jeg har lyst til at alle skal gjøre, kan dere gjøre den?
// Burde starte en eller annen prosess i reader, som gjør at vi kaller på commit/abort ved behov.
func (node Node) SYN(Command) // Prepare command - First message

// Jeg kan gjøre endringen
func (node Node) SYN_ACK(){ // Say that you can commit, if you cannot - use abort.
	message := node.p2p.Create_Message("SYN/ACK", peer_to_peer.MESSAGE)
	node.p2p.Broadcast(message)
}
// Endringen var ok for alle, gjør endringen.
func (node Node) COMMIT() //

// Brukes til å si at noe gikk galt, prøv igjen om litt.
func (node Node) ABORT() // If aborted, wait a random amount of time before trying again.

// Kun brukt til å si at de har hørt commit/abort. Ingenting mer.
func (node Node) ACK() // All-Ack