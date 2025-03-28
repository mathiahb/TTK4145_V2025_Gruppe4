Modules
---------------------------
common: Shared constants and functions for all other modules

elevator: The elevator, communicates with shared states to perform updates.
elevio: The code given to us to communicate with the physical elevator. Used by elevator.

shared_states: The place we store our information about the states of all elevators and the hall requests. Communicates with the elevator and network.
hall_request_assigner: Code given to us to assign hall requests to elevators. Used by shared state.

network: The module that handles connecting to the other nodes and handles performing the synchronization and 2PC protocols.
Network-go: Code given to us, uses peers module to detect alive nodes.


See doc.md for a visual representation of the connections.

Compile
---------------------------

In a terminal in the Project folder:
go build

How to start an elevator
---------------------------

./elevatorProject
./elevatorProject --elevatorPort port --name #

Name is by default computer name + 0. If 2 elevators share computer name, use --name # to switch the 0 to a different number.