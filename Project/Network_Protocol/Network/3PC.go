package network

import (
	"Constants"
	"time"

	peer_to_peer "github.com/mathiahb/TTK4145_V2025_Gruppe4/Network_Protocol/Network/Peer_to_Peer"
)

type Command struct {
	Field     string
	New_Value string
}

/*
type Message struct {
	message_type string
	id           TxID
	sender       string
	payload      string
}*/

// Jeg har en endring jeg har lyst til at alle skal gjøre, kan dere gjøre den?
// Burde starte en eller annen prosess i reader, som gjør at vi kaller på commit/abort ved behov.

//--------------------------------------------------------------------------------------------
// func (message Message) String() string {
// 	return fmt.Sprintf("%s txid=%s s=%s r=%s",
// 		message.message_type,
// 		message.id,
// 		message.sender,
// 		message.payload,
// 	)
// }

func (node *Node) PREPARE(message Message) { // Prepare command - First message in 3PC
	aliveNodes := node.Get_Alive_Nodes()
	//synMessage := fmt.Sprintf("%s %s=%s", Constants.PREPARE, cmd.Field, cmd.New_Value) // PREPARE Floor=3
	// message.String() --> "PREP txid=1 s=1 r=Floor=3"
	node.p2p.Broadcast(node.p2p.Create_Message(message.String(), peer_to_peer.MESSAGE))

	acks := 0
	total := len(aliveNodes)
	for acks < total {
		timeout := time.After(2 * time.Second) // Adjust timeout as needed
		select {
		case response := <-node.comm:
			if response.message_type == Constants.PREPARE_ACK { // Acknowledge the sync request
				acks++
			} else if response.message_type == Constants.ABORT_COMMIT {
				node.ABORT(message)
				return
			}
		case <-timeout:
			node.ABORT(message) // Abort if timeout is reached
			return

		}
	}
	// Hvis vi har nådd hit, så har vi fått ACK fra alle noder
	node.COMMIT(message)

}

// Jeg kan gjøre endringen, send meg en commit
func (node *Node) PREPARE_ACK(msg Message, respondingTo peer_to_peer.P2P_Message) {
	msg.message_type = Constants.PREPARE_ACK
	node.Broadcast_Response(msg, respondingTo)
}

/*
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
			if response.message_type == Constants.PRE_COMMIT_ACK {
				acks++
			} else if response.message_type == Constants.ABORT_COMMIT {
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
*/
// Endringen var ok for alle, gjør endringen.
func (node *Node) COMMIT(msg Message) {
	msg.message_type = Constants.COMMIT
	node.doLocalCommit(msg)
	// commitMessageStr := fmt.Sprintf("%s %s=%s", Constants.COMMIT, msg.payload)
	// commitMessage := node.p2p.Create_Message(msg.String(), peer_to_peer.MESSAGE)
	node.p2p.Broadcast(node.p2p.Create_Message(msg.String(), peer_to_peer.MESSAGE))
}

// Brukes til å si at noe gikk galt, prøv igjen om litt.
func (node *Node) ABORT(msg Message) { // If aborted, wait a random amount of time before trying again.
	msg.message_type = Constants.ABORT_COMMIT
	node.doLocalAbort(msg)
	// 2) Send ABORT til alle noder
	node.p2p.Broadcast(node.p2p.Create_Message(msg.String(), peer_to_peer.MESSAGE))
}

// Kun brukt til å si at de har hørt commit/abort. Ingenting mer.
func (node *Node) ACK(msg Message) { // All-Ack
	msg.message_type = Constants.ACK
	node.p2p.Broadcast(node.p2p.Create_Message(msg.String(), peer_to_peer.MESSAGE))
}
