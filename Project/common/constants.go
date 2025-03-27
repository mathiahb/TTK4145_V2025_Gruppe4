package common

import (
	"time"
)

// ----------------- ARGV ----------------------
const ARGV_TEST string = "--test"
const ARGV_BACKUP string = "--backup"
const ARGV_LISTENER_ONLY string = "--listener"
const ARGV_ELEVATOR_ID string = "--id"

const NO_ID int = -1

// --------------- ELEVATOR -------------------- //

const (
	N_FLOORS          = 4
	N_BUTTONS         = 3
	DoorOpenDurationS = 3.0
	IsStuckDurationS  = 4.0
)

// Motor direction (stop, up, down)
const (
	D_Stop Dirn = "stop"
	D_Up   Dirn = "up"
	D_Down Dirn = "down"
)

// Elevator behaviour:
// (idle, door open, moving,
// stuck moving i.e. stuck when the elevator is suppeosed to be moving,
// stuck door open i.e. stuck when door is open because of obstruction)
const (
	EB_Idle           ElevatorBehaviour = "idle"
	EB_DoorOpen       ElevatorBehaviour = "doorOpen"
	EB_Moving         ElevatorBehaviour = "moving"
	EB_Stuck_Moving   ElevatorBehaviour = "stuckMoving"
	EB_Stuck_DoorOpen ElevatorBehaviour = "stuckDoorOpen"
)

// Button types (Hall Up, Hall Down, Cab)
const (
	B_HallUp = iota
	B_HallDown
	B_Cab
)

// --------------- SHARED STATES -------------------- //
// Defines message types sent from the network to shared states after completing the two-phase commit protocol
const (
	ADD          = "ADD"
	REMOVE       = "REMOVE"
	UPDATE_STATE = "UPDATE_STATE"
)

// ---------------- NETWORK --------------------
const NO_DATA string = ""
const NULL string = "\000"
const TCP_BUFFER_SIZE int = 4096
const P2P_BUFFER_SIZE int = 4096

// Constants for P2P communication
const P2P_FIELD_DELIMINATOR string = "\\\r\n"
const P2P_DEP_DELIMINATOR string = "/"
const P2P_DEP_TIME_HORIZON int = 1000                                              // Amount of dependencies stored in memory
const P2P_MSG_TIME_HORIZON int = 1000                                              // Amount of messages sent stored in memory
const P2P_TIME_UNTIL_EXPECTED_ALL_CONNECTED time.Duration = time.Millisecond * 100 // Expected time for P2P to find other nodes

const LAMPORT_CLOCK_WRAPAROUND_LOWER_EDGE int = -1e10
const LAMPORT_CLOCK_WRAPAROUND_UPPER_EDGE int = 1e10

// 2-phase commit messages - 4 bytes long
const SIZE_TYPE_FIELD int = 4               // 4 Bytes per message type
const PREPARE string = "PREP"               // SYNchronize
const PREPARE_ACK string = "PREA"           // SYNchronize ACKnowledge
const COMMIT string = "COMT"                // COMmiT
const ABORT_COMMIT string = "ERRC"          // Error commit
const ABORT_SYNCHRONIZATION string = "ABRT" // ABORT DiSCovery

const ACK string = "ACKS" // ACKnowledgeS

// Synchronization messages [4 bytes long]
const SYNC_REQUEST string = "SYNC"  // SYNChronize
const SYNC_RESPONSE string = "SRSP" // Synchronize ReSPonse
const SYNC_RESULT string = "SRST"   // Synchronize ReSulT

const NETWORK_FIELD_DELIMITER = "\\\n"

// Constants for UDP connection
const UDP_PORT int = 10005
const PEERS_PORT int = 10006
const UDP_BROADCAST_IP_PORT string = "239.255.255.255:10005"

// Constants for TCP connection
const TCP_PORT int = 20005
const TCP_MAX_FAIL_WRITES int = 10
const TCP_HEARTBEAT = "H"

// Deadlines
const UDP_READ_DEADLINE time.Duration = time.Millisecond
const UDP_WAIT_BEFORE_TRANSMITTING_AGAIN time.Duration = time.Millisecond // Writing: 1000 Hz
const UDP_WAIT_BEFORE_READING_AGAIN time.Duration = time.Microsecond      // Reading: 1e6 Hz

const TCP_READ_DEADLNE time.Duration = time.Millisecond              // 1 Millisecond read deadline.
const TCP_WAIT_BEFORE_READING_AGAIN time.Duration = time.Millisecond // Checks TCP 500~1000 times a second
const TCP_HEARTBEAT_TIME time.Duration = time.Second
const TCP_RESEND_HEARTBEAT time.Duration = time.Millisecond * 10
