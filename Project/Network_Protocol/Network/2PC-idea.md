
   Coordinator                           Participants
     |                                       |
     | -- PREPARE  ----------------------->  | 
     | <-- PREPARE_ACK or ABORT_COMMIT ----  |  
     |
     | -- COMMIT ------------------------->  | 
     | <-- ACK ---------------------------   |  


1. Skal vi implementere ACKs som et siste steg etter COMMIT er sendt? Nå har jeg bare latt COMMIT avslutte protokollen.
2. doLocalAbort() - Trenger vi egentlig å gjøre noe her? Det finnes vel ikke tilfeller hvor en ABORT kommer etter en COMMIT; dermed vil aldri endringen faktisk være gjort når en   ABORT skjer



// En alternativ coordinate_2PC hvor vi venter på ACKs()


func (node *Node) coordinate_2PC(cmd Command, success_channel chan bool) {
    // Ensure we only coordinate one protocol at a time on this node
    node.mu_voting_resource.Lock()
    defer node.mu_voting_resource.Unlock()

    // Build a PREPARE message. We embed "Field=Value" in the payload
    txid := node.generateTxID()
    payload := fmt.Sprintf("%s=%s", cmd.Field, cmd.New_Value)
    prepareMsg := node.create_Message(Constants.PREPARE, txid, payload)

    // Step 1: Broadcast PREPARE
    node.Broadcast(prepareMsg)
    fmt.Printf("[%s] 2PC coordinator: broadcast PREPARE %s\n", node.name, prepareMsg.String())

    neededAcks := len(node.Get_Alive_Nodes()) // # of participants we expect
    ackCount := 0

    // Wait for participants' PREPARE_ACK or for an ABORT
    prepareTimeout := time.After(2 * time.Second)

WAIT_PREPARE_ACKS:
    for {
        if ackCount == neededAcks {
            break WAIT_PREPARE_ACKS
        }
        select {
        case msg := <-node.comm:
            // Only handle messages for our TxID
            if msg.id != txid {
                continue
            }
            switch msg.message_type {
            case Constants.PREPARE_ACK:
                ackCount++
                fmt.Printf("[%s] 2PC coordinator got PREPARE_ACK from %s (%d/%d)\n",
                    node.name, msg.sender, ackCount, neededAcks)

            case Constants.ABORT_COMMIT:
                fmt.Printf("[%s] 2PC coordinator sees ABORT => we ABORT.\n", node.name)
                node.abort2PC(txid)
                success_channel <- false
                return
            }
        case <-prepareTimeout:
            // Timed out => ABORT
            fmt.Printf("[%s] 2PC coordinator timed out waiting for PREPARE_ACK => ABORT.\n", node.name)
            node.abort2PC(txid)
            success_channel <- false
            return
        }
    }

    // Step 2: If we got all PREPARE_ACK, broadcast COMMIT
    commitMsg := node.create_Message(Constants.COMMIT, txid, payload)
    node.Broadcast(commitMsg)
    fmt.Printf("[%s] 2PC coordinator: broadcast COMMIT.\n", node.name)

    // Optionally, the coordinator might do its own local commit right here:
    node.doLocalCommit(commitMsg)

    // Step 3: Wait for final ACK from all participants
    finalAckCount := 0
    ackTimeout := time.After(2 * time.Second)

WAIT_FINAL_ACKS:
    for {
        if finalAckCount == neededAcks {
            // Great! Everyone confirmed they committed.
            fmt.Printf("[%s] 2PC coordinator received final ACK from all participants.\n", node.name)
            success_channel <- true
            return
        }
        select {
        case msg := <-node.comm:
            if msg.id != txid {
                continue
            }
            if msg.message_type == Constants.ACK {
                finalAckCount++
                fmt.Printf("[%s] 2PC coordinator got final ACK from %s (%d/%d)\n",
                    node.name, msg.sender, finalAckCount, neededAcks)
            } else if msg.message_type == Constants.ABORT_COMMIT {
                // If a participant aborts now, it's basically an error:
                fmt.Printf("[%s] 2PC coordinator sees ABORT during final ack => ignoring.\n", node.name)
            }
        case <-ackTimeout:
            // If participants don't finalize, you might log an error or re-try
            fmt.Printf("[%s] 2PC coordinator timed out waiting for final ACK => partial success.\n", node.name)
            success_channel <- false
            return
        }
    }
}
