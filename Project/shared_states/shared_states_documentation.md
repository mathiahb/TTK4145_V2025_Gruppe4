Shared States
----

The shared states module is the bridge between the elevator and the network. Shared states holds all hall requests, elevator states and which nodes are connected to the network. 


A peer-to-peer network is dependent on all the nodes sharing the same worldview at all times. Therefore the local elevator sends updated information through channels to the shared states modules every time there is a new hall request, a hall request that should be cleared or an updated state on the elevator. The shared states module then translates and sends the information to the network (2 phase commit) so that all the elevators will be updated and have identical shared states. 

The network updates the shared states module with information about which nodes are connected to the network. 

The shared states are then used to call upon the hall request assignment (HRA). Since all elevators calls upon the HRA with the same input, the ouput will also be identical. The hall requests are then translated and send to the elevator. 


Shared state is also used when the network is syncronizing, as it hold all information about .... Conflict resolver







