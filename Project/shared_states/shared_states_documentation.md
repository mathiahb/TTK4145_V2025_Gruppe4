Shared States
----
This module acts as a bridge between the elevator logic and the network. It stores:

* All hall requests,
* Elevator states, 
* Information about which nodes are currently connected.


In a peer-to-peer network it is critical that all nodes shares the same worldview. Otherwise, the decisions they make are not coherent. Therefore, whenever the local elevator has a new hall request, an updated state, or wants to clear an existing request, it sends updated information through channels to the shared states modules. The shared states module then translates and sends the information to the network (2 phase commit) so that all the elevators remain synchronized.

Meanwhile, the network also sends updates back to the shared states module with information about alive nodes. This ensures that only connected and non-stuck nodes are included in the HRA (Hall request assigner)-input. Since all connected elevators calls upon the HRA with the same input, the ouput will also be identical. The hall requests are then translated and sent to the respective elevators. 

Shared state is also central to network syncronization. When the network initiates a synchronization protocol (sending a "protocol request information" signal), the shared states module responds by sending its entire local state to the network. In return it receives a final and consistent shared state.



Conflict Resolver
----
Used during synchronization

The conflict resolver takes in a map of node names to their shared states and produces a single unified state to be used by all nodes. The main purpose of this function is to solve conflicts between the shared states. The following procedure is used:
States - The owner of the state has priority, if none have been supplied by the owner the first entry is picked. (This ensures that the state will stay updated whenever a node is connected, and also keeps it around should the node crash.)
Hall Requests - Assuming the worst possible case, we assume that if any node says a hall request exists, it exists. (The no dropping orders after turning on the light.)