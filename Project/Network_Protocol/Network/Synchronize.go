package network

import (
	"Constants"
	"time"

	peer_to_peer "github.com/mathiahb/TTK4145_V2025_Gruppe4/Network_Protocol/Network/Peer_to_Peer"
)

func (node Node) Synchronize() {
	node.start_Synchronization()

	node.handle_Synchronization()
}

func (node Node) handle_Synchronization() {
	vote := protocols.New_Synchronization_Vote()
	node.active_synchronization = vote
	node.has_active_synchronization = true

	synchronization_time := time.Millisecond * 10
	time.Sleep(synchronization_time)

	node.finalize_Synchronization()

	node.has_active_synchronization = false
}

func (node Node) start_Synchronization() {
	message := node.p2p.Create_Message(Constants.SYNC_MESSAGE, peer_to_peer.MESSAGE)

	node.p2p.Broadcast(message)
}

func (node Node) finalize_Synchronization() {
	result := node.active_synchronization.Get_Result()
	node.active_voter_names = result.Voters

}

func (node Node) respond_to_sync(sync_message peer_to_peer.P2P_Message) {

	node.p2p.Broadcast(message)
}
