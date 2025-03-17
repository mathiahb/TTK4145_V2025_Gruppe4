package network

func (node *Node) begin_Discovery() {
	node.busy_voting.TryLock()

}
