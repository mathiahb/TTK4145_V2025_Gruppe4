module tests

require Network-Protocol v0.0.0
require Constants v0.0.0

replace Network-Protocol => ../shared_state/Network_Protocol
replace Constants => ../Constants/

go 1.16