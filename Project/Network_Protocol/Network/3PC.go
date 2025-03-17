package network

import (
	"Constants"
	"fmt"
	"time"

	peer_to_peer "github.com/mathiahb/TTK4145_V2025_Gruppe4/Network_Protocol/Network/Peer_to_Peer"
)

type Command struct {
	Field     string
	New_Value string
	// TxID      string
}

// Jeg har en endring jeg har lyst til at alle skal gjøre, kan dere gjøre den?
// Burde starte en eller annen prosess i reader, som gjør at vi kaller på commit/abort ved behov.

//--------------------------------------------------------------------------------------------

func (node *Node) SYN(cmd Command) { // Prepare command - First message in 3PC
	aliveNodes := node.Get_Alive_Nodes()
	synMessage := fmt.Sprintf("%s %s=%s", Constants.SYNC_MESSAGE, cmd.Field, cmd.New_Value) // SYN Floor=3
	node.p2p.Broadcast(node.p2p.Create_Message(synMessage, peer_to_peer.MESSAGE))

	acks := 0
	total := len(aliveNodes)
	for {
		timeout := time.After(2 * time.Second) // Adjust timeout as needed
		select {
		case response := <-node.comm:
			if response == Constants.PREPARE_ACK { // Acknowledge the sync request
				acks++
			} else if response == Constants.ABORT_COMMIT {
				node.ABORT()
				return
			}

		case <-timeout:
			node.ABORT() // Abort if timeout is reached
			return
		default:
		}
		if acks == total {
			// Instead of committing we should send a prepare message (in order to have a real 3PC)
			node.PREPARE_COMMIT(cmd)
			return
		}
	}
}

// Jeg kan gjøre endringen
func (node *Node) PREPARE_ACK() { // Say that you acknowledge the sync request and agree to the change
	message := node.p2p.Create_Message(Constants.PREPARE_ACK, peer_to_peer.MESSAGE)
	node.p2p.Broadcast(message)

}

// Prepare commit - låser ressurser

func (node *Node) PREPARE_COMMIT(cmd Command) {
	aliveNodes := node.Get_Alive_Nodes()
	prepareMessage := fmt.Sprintf("%s %s=%s", Constants.PRE_COMMIT, cmd.Field, cmd.New_Value)
	node.p2p.Broadcast(node.p2p.Create_Message(prepareMessage, peer_to_peer.MESSAGE))

	acks := 0
	total := len(aliveNodes)
	for {
		timeout := time.After(2 * time.Second)
		select {
		case response := <-node.comm:
			if response == Constants.PRE_COMMIT_ACK {
				acks++
			} else if response == Constants.ABORT_COMMIT {
				node.ABORT()
				return
			}

		case <-timeout:
			node.ABORT()
			return
		}
		if acks == total {
			node.COMMIT(cmd)
			break
		}
	}
}

// Jeg låser meg til endringen
func (node *Node) PREPARE_COMMIT_ACKNOWLEDGE() {

	//Logikk for å se
	message := node.p2p.Create_Message(Constants.PRE_COMMIT_ACK, peer_to_peer.MESSAGE)
	node.p2p.Broadcast(message)
}

// Endringen var ok for alle, gjør endringen.
func (node *Node) COMMIT(cmd Command) {
	node.doLocalCommit(cmd)
	commitMessageStr := fmt.Sprintf("%s %s=%s", Constants.COMMIT, cmd.Field, cmd.New_Value)
	commitMessage := node.p2p.Create_Message(commitMessageStr, peer_to_peer.MESSAGE)
	node.p2p.Broadcast(commitMessage)
}

// Brukes til å si at noe gikk galt, prøv igjen om litt.
func (node *Node) ABORT() { // If aborted, wait a random amount of time before trying again.
	node.doLocalAbort()

	// 2) Send ABORT til alle noder
	abortMessage := node.p2p.Create_Message(Constants.ABORT_COMMIT, peer_to_peer.MESSAGE)
	node.p2p.Broadcast(abortMessage)
}

// Kun brukt til å si at de har hørt commit/abort. Ingenting mer.
func (node *Node) ACK() { // All-Ack
	ackMessage := node.p2p.Create_Message(Constants.ACK, peer_to_peer.MESSAGE)
	node.p2p.Broadcast(ackMessage)
}
