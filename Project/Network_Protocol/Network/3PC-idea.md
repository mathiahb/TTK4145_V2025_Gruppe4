
   Coordinator                           Participants
     |                                       |
     | -- PREPARE  ----------------------->  | 
     | <-- PREPARE_ACK or ABORT_COMMIT ----  |  
     |
     | -- PRE_COMMIT --------------------->  |  
     | <-- PRE_COMMIT_ACK ----------------   |
     |
     | -- COMMIT ------------------------->  | 
     | <-- ACK ---------------------------   |  



Skal vi implementere TxID som en del av Command? 
- Når en kordinator starter en PREPARE/SYNC så lages en unik ID (transaction ID). 
- Denne IDen kan da brukes til å sørge for at man adresserer riktig
- Se koden under for en ish implementasjon  man da kunne hatt i feks PREPARE(cmd Command).
- En lignende implementasjon måtte da blitt gjort i samtlige ledd i 3PC.
- På denne måten vil vi unngå å godta en acknowledgement for en annen prosess for denne prosessen.


prepareMsg := fmt.Sprintf("%s tx=%s %s=%s",
        Constants.PREPARE,        
        cmd.TxID,                 
        cmd.Field,                
        cmd.New_Value,            
    )
node.p2p.Broadcast(node.p2p.Create_Message(prepareMsg, peer_to_peer.MESSAGE))

msgType, txID := parseMessageTypeAndTxID(response)
            if txID != cmd.TxID {
                // Ignore or handle if message belongs to a different transaction
                continue
            }

