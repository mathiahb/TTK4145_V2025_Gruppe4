package peerToPeer

import (
	"elevator_project/common"
	"strconv"
	"testing"
	"time"
)

func Test_Dependency_Horizon(t *testing.T) {
	handler := NewDependencyHandler()
	clock := NewLamportClock()

	owner := "OWNER"

	clock.Event()

	first_dependency := Dependency{owner, clock}

	// 1 more than dependency horizon, first dependency on list should be 1, but should be gone after crossing horizon.
	for i := 0; i < common.P2P_DEP_TIME_HORIZON+1; i++ {
		dependency := Dependency{owner, clock}

		handler.Add_Dependency(dependency)

		if !handler.HasDependency(dependency) {
			t.Fatal("Handler did not add dependency to list!\n")
		}

		clock.Event()
	}

	if handler.HasDependency(first_dependency) {
		t.Fatalf("First dependency was not removed! Map length: %d, Heap length: %d\n", len(handler.lookup_map), handler.min_heap.Len())
	}

	second_dependency := Dependency{owner, NewLamportClockFromString("2")}

	if !handler.HasDependency(second_dependency) {
		t.Fatalf("Second dependency was removed! Map length %d, Heap length %d\n", len(handler.lookup_map), handler.min_heap.Len())
	}

}

func Test_Lamport_Clock_Wraparound(t *testing.T) {
	clock_high := LamportClock{common.LAMPORT_CLOCK_WRAPAROUND_UPPER_EDGE + 1}
	clock_low := LamportClock{common.LAMPORT_CLOCK_WRAPAROUND_LOWER_EDGE - 1}

	if !clock_high.Is_Less_Than(clock_low) {
		t.Fatal("Wraparound clock is not returning true on wraparound!")
	}
}

func Test_Dependency_Wraparound(t *testing.T) {
	handler := NewDependencyHandler()
	clock := NewLamportClock()

	clock.time = common.LAMPORT_CLOCK_WRAPAROUND_UPPER_EDGE + 1

	low_time := common.LAMPORT_CLOCK_WRAPAROUND_LOWER_EDGE - common.P2P_DEP_TIME_HORIZON // Avoid cyclical dependency

	for i := 0; i < common.P2P_DEP_TIME_HORIZON+1; i++ {
		dependency := Dependency{strconv.Itoa(clock.time), clock}

		handler.Add_Dependency(dependency)

		if !handler.HasDependency(dependency) {
			t.Fatal("Handler did not add dependency to list!\n")
		}

		clock.Event()
	}

	clock.time = low_time

	for i := 0; i < common.P2P_DEP_TIME_HORIZON+1; i++ {
		dependency := Dependency{strconv.Itoa(clock.time), clock}

		handler.Add_Dependency(dependency)
		old_dependency := Dependency{strconv.Itoa(clock.time - 1), LamportClock{clock.time - 1}}

		if !handler.HasDependency(dependency) {
			t.Fatal("Handler did not add dependency to list!\n")
		}

		// Old dependency doesn't exist for i = 0.
		if i != 0 && !handler.HasDependency(old_dependency) {
			t.Fatalf("Handler did not keep the new dependency: %s!\n", old_dependency.ToString())
		}

		clock.Event()
	}
}

func Test_P2P_Message_String(t *testing.T) {
	sender_field := "SENDER"
	type_field := MESSAGE
	time_field := NewLamportClockFromString("6")
	dependency_field := NewDependency("OTHER", NewLamportClockFromString("3"))
	body_field := "Hello from body!"

	test_tcp_message :=
		sender_field + common.P2P_FIELD_DELIMINATOR +
			string(type_field) + common.P2P_FIELD_DELIMINATOR +
			time_field.String() + common.P2P_FIELD_DELIMINATOR +
			dependency_field.ToString() + common.P2P_FIELD_DELIMINATOR +
			body_field

	p2pMessage := P2PMessageFromString(test_tcp_message)

	t.Logf("P2P_message generated: %s\n", p2pMessage.ToString())

	if p2pMessage.Sender != sender_field {
		t.Fatalf("Sender field mismatch!\n%s != %s\n", p2pMessage.Sender, sender_field)
	}

	if string(p2pMessage.Type) != string(type_field) {
		t.Fatalf("Type field mismatch!\n%s != %s\n", p2pMessage.Type, type_field)
	}

	if p2pMessage.Time.String() != time_field.String() {
		t.Fatalf("Time field mismatch!\n%s != %s\n", p2pMessage.Time.String(), time_field.String())
	}

	if p2pMessage.dependency.ToString() != dependency_field.ToString() {
		t.Fatalf("Dependency field mismatch!\n%s != %s\n", p2pMessage.dependency.ToString(), dependency_field.ToString())
	}

	if p2pMessage.Message != body_field {
		t.Fatalf("Body field mismatch!\n%s != %s\n", p2pMessage.Message, body_field)
	}

	stringed_p2p_message := p2pMessage.ToString()

	if stringed_p2p_message != test_tcp_message {
		t.Fatalf("Stringed message:\n%s\n---\nDid not match origin string %s\n", stringed_p2p_message, test_tcp_message)
	}
}

func Test_Message_Horizon(t *testing.T) {
	resolver := NewDependencyResolver()
	clock := NewLamportClock()

	// 1 more than horizon
	for i := 0; i < common.P2P_MSG_TIME_HORIZON+1; i++ {
		clock.Event()

		p2pMessage := NewP2PMessage("SENDER", MESSAGE, clock, "BODY")
		resolver.EmplaceNewMessage(p2pMessage)

		dependency := NewDependency("SENDER", clock)

		resolved_p2p_message, ok := resolver.Get_Message(dependency)

		if !ok {
			t.Fatalf("Failed to get an ok on sending %s:\n", p2pMessage.ToString())
		}

		if p2pMessage.ToString() != resolved_p2p_message.ToString() {
			t.Fatalf("Returned string was ok, but not correct! %s != %s\n",
				p2pMessage.ToString(), resolved_p2p_message.ToString())
		}
	}

	first_dependency := NewDependency("SENDER", NewLamportClockFromString("1"))
	second_dependency := NewDependency("SENDER", NewLamportClockFromString("2"))

	message, ok := resolver.Get_Message(first_dependency)

	if ok {
		t.Fatalf("Received a message that was supposed to be out of horizon!\n Resolver returned: %s\n",
			message.ToString())
	}

	message, ok = resolver.Get_Message(second_dependency)

	if !ok {
		t.Fatal("Did not find the second message that was supposed to be in horizon!")
	}

	p2pMessage := NewP2PMessage("SENDER", MESSAGE, NewLamportClockFromString("2"), "BODY")
	if message.ToString() != p2pMessage.ToString() {
		t.Fatalf("Returned second dependency was ok, but not correct! %s != %s\n",
			message.ToString(), p2pMessage.ToString())
	}
}

func Test_Network(t *testing.T) {
	network_1 := NewP2PNetwork()
	network_2 := NewP2PNetwork()
	defer network_1.Close()
	defer network_2.Close()

	time.Sleep(time.Second)

	p2pMessage := network_1.CreateMessage("Hello!")

	network_1.Broadcast(p2pMessage)

	time.Sleep(time.Second)

	select {
	case received_message := <-network_2.ReadChannel:
		if p2pMessage.ToString() != received_message.ToString() {
			t.Fatalf("Error Peer2Peer! Message received did not match sent!\n%s != %s\n",
				p2pMessage.ToString(), received_message.ToString())
		}
	default:
		t.Fatal("Error Peer2Peer! Message was never received!")
	}
}

func Test_Dependency_Resend(t *testing.T) {
	network_1 := NewP2PNetwork()
	network_2 := NewP2PNetwork()
	defer network_1.Close()
	defer network_2.Close()

	time.Sleep(time.Second) // Let the networks connect to each other. Takes about 1.2x UDP SERVER LIFETIME (~300ms)

	p2pMessage := network_1.CreateMessage("Hello!")

	network_1.Broadcast(p2pMessage)

	network_3 := NewP2PNetwork()
	defer network_3.Close()

	time.Sleep(time.Second)

	select {
	case received_message := <-network_2.ReadChannel:
		if p2pMessage.ToString() != received_message.ToString() {
			t.Fatalf("Error Peer2Peer! Message received did not match sent!\n%s != %s\n",
				p2pMessage.ToString(), received_message.ToString())
		}

		p2p_depended_message := network_2.CreateMessage("Hello 2!")
		p2p_depended_message.DependOn(received_message)

		network_2.Broadcast(p2p_depended_message)

		time.Sleep(time.Second)

		select {
		case received_message_network_3 := <-network_3.ReadChannel:
			if p2pMessage.ToString() != received_message_network_3.ToString() {
				t.Fatalf("Error Peer2Peer! Depended message received did not match sent!\n%s != %s\n",
					p2pMessage.ToString(), received_message.ToString())
			}

			select {
			case received_depended_message := <-network_3.ReadChannel:
				if p2p_depended_message.ToString() != received_depended_message.ToString() {
					t.Fatalf("Error Peer2Peer! Depended message received did not match sent!\n%s != %s\n",
						p2pMessage.ToString(), received_message.ToString())
				}
			default:
				t.Fatal("Error Peer2Peer Network 3! Hello 2 Message was never received!")
			}

		default:
			t.Fatal("Error Peer2Peer Network 3! Hello Message was never received!")
		}

	default:
		t.Fatal("Error Peer2Peer Network 2! Message was never received!")
	}
}

func Test_Double_Send(t *testing.T) {
	network_1 := NewP2PNetwork()
	network_2 := NewP2PNetwork()
	defer network_1.Close()
	defer network_2.Close()

	time.Sleep(time.Second)

	p2pMessage := network_1.CreateMessage("Hello!")

	network_1.Broadcast(p2pMessage)
	network_1.Broadcast(p2pMessage)

	time.Sleep(time.Second)

	select {
	case received_message := <-network_2.ReadChannel:
		if p2pMessage.ToString() != received_message.ToString() {
			t.Fatalf("Error Peer2Peer! Message received did not match sent!\n%s != %s\n",
				p2pMessage.ToString(), received_message.ToString())
		}

		select {
		case spurious_message := <-network_2.ReadChannel:
			t.Fatalf("Error Peer2Peer, a message got sent twice! %s", spurious_message.ToString())
		default:
		}
	default:
		t.Fatal("Error Peer2Peer! Message was never received!")
	}
}
