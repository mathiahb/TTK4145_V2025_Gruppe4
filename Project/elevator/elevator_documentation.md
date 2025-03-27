Elevator
---

The elevator recieves information of the physical elevator from elevio. The logic of the elevator is implemented in the FSM- (finite state machine) and request-file. 

The elevator routine gets information from elevio and sends it to the shared state module to update the other elevators. In return from shared states the elevator gets hall requests. It is worth noting that the elevator module has no information of whether it is connected to other nodes or not.

The elevator is also dependent on shared states to set their cab- and hallrequest-lights. This is because the hall requests lights are supposed to be identical on all the elevators. Cab requests on the other hand are only decided by the local elevator. Still it is crucial that only cab lights that have been confirmed by the network are lit so that the system is robust in case of a network diconnection. 

the code handles two cases where the elevator is connected to the network but still not operable.
1. When the elevator is obstructed.
2. If the motor is disconnected or obstructed.
In both cases the elevator will tell the shared states module that it is stuck.
