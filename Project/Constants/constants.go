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

// Connection names
const CLIENT_CREATED_NAME string = "client_" // + address
const SERVER_CREATED_NAME string = "server_" // + address

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
const UDP_READ_DEADLINE time.Duration = time.Millisecond                   // Shouldn't need more time to check if anything has been sent.
const UDP_WAIT_BEFORE_READING_AGAIN time.Duration = 500 * time.Millisecond // Checks UDP ~2 times a second
const TCP_READ_DEADLNE time.Duration = time.Millisecond                    // 1 Millisecond read deadline.
const TCP_WAIT_BEFORE_READING_AGAIN time.Duration = time.Millisecond       // Checks TCP ~1000 times a second
