module github.com/mathiahb/TTK4145_V2025_Gruppe4

require tests v0.0.0
require Constants v0.0.0

replace tests => ./tests/
replace Constants => ./Constants/
replace Network-Protocol => ./shared_state/Network_Protocol/

go 1.16