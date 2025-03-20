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
