The module is built around three main elements:

Discovery – Keeps track of which nodes are currently connected.
Synchronization – Gathers and shares state information among all active nodes.
Two-Phase Commit (2PC) – Ensures that changes to shared state are either fully committed or aborted in a coordinated manner.

Overview:

Node: Represents a single node in the network, with an identifier and a p2p channel for communication.
The node receives messages and forwards it to the 2PC and Synchronization protocols for communication.

AliveNodeManager: Maintains a list of “alive” nodes currently present in the network.

ProtocolDispatcher: Serves as a control center that receives requests to run Discovery, Synchronization, and 2PC (called “commands”). The dispatcher handles protocols in a queue, making sure that one protocol run finishes – or fails – before the next begins.

NetworkCommunicationChannels: 


   Coordinator                           Participants
     |                                       |
     | -- PREPARE  ----------------------->  | 
     | <-- PREPARE_ACK or ABORT_COMMIT ----  |  
     |
     | -- COMMIT ------------------------->  | 
     | <-- ACK ---------------------------   |  


