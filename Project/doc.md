Main
---

Depending on arguments added will do one of the following:

- Enter test mode (currently testing TCP)
- Enter listener mode (not implemented)
- Enter backup mode (not implemented)
- Continue to normal mode (not implemented)

In test mode:
---
Will start some test code.

In listener mode:
---
Will open a listener node and print to terminal the shared state. + Any messages on the network.

In backup mode:
---
Will just run the local heartbeat code in backup.

Normal mode:
---
Firstly spawns a backup from the local heartbeat module.
Will start the Shared State, then the Network, then the Elevator.


Gammel mod-fil:
module github.com/mathiahb/TTK4145_V2025_Gruppe4

require Driver-Elevator v0.0.0

require (
	Driver-Elevio v0.0.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

require (
	Constants v0.0.0
	golang.org/x/net v0.35.0
)

replace Constants => ./Constants/

replace Driver-Elevator => ./elevator/

replace Driver-Elevio => ./elevator/elevio/

go 1.18