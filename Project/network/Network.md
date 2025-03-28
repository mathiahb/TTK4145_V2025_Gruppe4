The module is built around three main elements:

1. Discovery – Keeps track of which nodes are currently connected.
2. Synchronization – Gathers and shares state information among all active nodes. The protocol uses messages like SYNC_REQUEST, SYNC_RESPONSE, and SYNC_RESULT to gather all local states and produce a consistant mutual shared state.
3. Two-Phase Commit (2PC) – Ensures that changes to shared state are either fully committed or aborted in a coordinated manner.

Overview:

Node: Represents a single node in the network, with an identifier and a p2p channel for communication.
The node receives messages and forwards it to the 2PC and Synchronization protocols for communication.

AliveNodeManager: Maintains a list of “alive” nodes currently present in the network.

ProtocolDispatcher: Serves as a control center that receives requests to run Discovery, Synchronization, and 2PC (called “commands”). The dispatcher handles protocols in a queue, making sure that one protocol run finishes – or fails – before the next begins.

NetworkCommunicationChannels: Used by the sharedstate to communicate with the network. 


Our implementation of 2 Phase Commit follows the structure shown below. In classic 2PC, participants also acknowledges the commit. We found that a 1.5 phase commit was sufficient.


   Coordinator                           Participants
     |                                       |
     | -- PREPARE  ----------------------->  | 
     | <-- PREPARE_ACK or ABORT_COMMIT ----  |  
     |
     | -- COMMIT ------------------------->  | 


