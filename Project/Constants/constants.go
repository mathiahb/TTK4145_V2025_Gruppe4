package Constants

import "time"

// ----------------- ARGV ----------------------
const ARGV_TEST string = "--test"
const ARGV_BACKUP string = "--backup"
const ARGV_LISTENER_ONLY string = "--listener"
const ARGV_ELEVATOR_ID string = "--id"

const NO_ID int = -1

// --------------- ELEVATOR --------------------

// ---------------- NETWORK --------------------
const NO_DATA string = ""
const NULL string = "\000"
const TCP_BUFFER_SIZE int = 4096
const P2P_BUFFER_SIZE int = 4096

// Constants for P2P communication
const P2P_FIELD_DELIMINATOR string = "\r\n"
const P2P_DEP_DELIMINATOR string = "/"
const P2P_DEP_TIME_HORIZON int = 100 // Amount of dependencies stored in memory
const P2P_MSG_TIME_HORIZON int = 10  // Amount of messages sent stored in memory

const LAMPORT_CLOCK_WRAPAROUND_LOWER_EDGE int = -1e10
const LAMPORT_CLOCK_WRAPAROUND_UPPER_EDGE int = 1e10

// 3-phase commit messages
const ACKNOWLEDGE string = "SYN/ACK"
const ALL_ACKNOWLEDGE string = "ACK"
const ABORT_COMMIT string = "ERR"

// Heartbeat
const HEARTBEAT_MESSAGE string = "HB"

// Constants for UDP connection
const UDP_PORT string = ":10005"
const TCP_PORT string = ":20005"
const UDP_BROADCAST_IP_PORT string = "255.255.255.255" + UDP_PORT

// Deadlines
const UDP_READ_DEADLINE time.Duration = time.Millisecond
const UDP_WAIT_BEFORE_TRANSMITTING_AGAIN time.Duration = 50 * time.Millisecond // Writing: 20 Hz
const UDP_WAIT_BEFORE_READING_AGAIN time.Duration = 50 * time.Millisecond      // Reading: 20 Hz
const UDP_SERVER_LIFETIME time.Duration = 200 * time.Millisecond               // Lifetime: 5 Hz
const UDP_UNTIL_SERVER_BOOT time.Duration = 200 * time.Millisecond             // Reboot: 5 Hz

const TCP_SERVER_BOOTUP_TIME time.Duration = time.Second // Used during bootup of P2P network.

const TCP_READ_DEADLNE time.Duration = time.Millisecond              // 1 Millisecond read deadline.
const TCP_WAIT_BEFORE_READING_AGAIN time.Duration = time.Millisecond // Checks TCP ~1000 times a second
