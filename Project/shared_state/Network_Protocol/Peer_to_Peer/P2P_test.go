package peer_to_peer

import (
	"Constants"
	"testing"
	"time"
)

func Test_Dependency_Horizon(t *testing.T) {
	handler := New_Dependency_Handler()
	clock := New_Lamport_Clock()

	owner := "OWNER"

	clock.Event()

	first_dependency := Dependency{owner, clock}

	// 1 more than dependency horizon, first dependency on list should be 1, but should be gone after crossing horizon.
	for i := 0; i < Constants.P2P_DEP_TIME_HORIZON+1; i++ {
		dependency := Dependency{owner, clock}

		handler.Add_Dependency(dependency)

		if !handler.Has_Dependency(dependency) {
			t.Fatal("Handler did not add dependency to list!\n")
		}

		clock.Event()
	}

	if handler.Has_Dependency(first_dependency) {
		t.Fatalf("First dependency was not removed! Map length: %d, Heap length: %d\n", len(handler.lookup_map), handler.min_heap.Len())
	}

	second_dependency := Dependency{owner, New_Lamport_Clock_From_String("2")}

	if !handler.Has_Dependency(second_dependency) {
		t.Fatalf("Second dependency was removed! Map length %d, Heap length %d\n", len(handler.lookup_map), handler.min_heap.Len())
	}

}

func Test_P2P_Message_String(t *testing.T) {
	sender_field := "SENDER"
	type_field := MESSAGE
	time_field := New_Lamport_Clock_From_String("6")
	dependency_field := New_Dependency("OTHER", New_Lamport_Clock_From_String("3"))
	body_field := "Hello from body!"

	test_tcp_message :=
		sender_field + Constants.P2P_FIELD_DELIMINATOR +
			string(type_field) + Constants.P2P_FIELD_DELIMINATOR +
			time_field.String() + Constants.P2P_FIELD_DELIMINATOR +
			dependency_field.To_String() + Constants.P2P_FIELD_DELIMINATOR +
			body_field

	p2p_message := P2P_Message_From_String(test_tcp_message)

	t.Logf("P2P_message generated: %s\n", p2p_message.To_String())

	if p2p_message.Sender != sender_field {
		t.Fatalf("Sender field mismatch!\n%s != %s\n", p2p_message.Sender, sender_field)
	}

	if string(p2p_message.Type) != string(type_field) {
		t.Fatalf("Type field mismatch!\n%s != %s\n", p2p_message.Type, type_field)
	}

	if p2p_message.Time.String() != time_field.String() {
		t.Fatalf("Time field mismatch!\n%s != %s\n", p2p_message.Time.String(), time_field.String())
	}

	if p2p_message.dependency.To_String() != dependency_field.To_String() {
		t.Fatalf("Dependency field mismatch!\n%s != %s\n", p2p_message.dependency.To_String(), dependency_field.To_String())
	}

	if p2p_message.Message != body_field {
		t.Fatalf("Body field mismatch!\n%s != %s\n", p2p_message.Message, body_field)
	}

	stringed_p2p_message := p2p_message.To_String()

	if stringed_p2p_message != test_tcp_message {
		t.Fatalf("Stringed message:\n%s\n---\nDid not match origin string %s\n", stringed_p2p_message, test_tcp_message)
	}
}

func Test_Network(t *testing.T) {
	defer time.Sleep(time.Second) // Let the servers shut down before doing anything else...

	network_1 := New_P2P_Network("20005")
	network_2 := New_P2P_Network("20006")
	defer network_1.Close()
	defer network_2.Close()

	time.Sleep(time.Second)

	p2p_message := network_1.Create_Message("Hello!", MESSAGE)

	network_1.Broadcast(p2p_message)

	time.Sleep(time.Second)

	select {
	case received_message := <-network_2.Read_Channel:
		if p2p_message.To_String() != received_message.To_String() {
			t.Fatalf("Error Peer2Peer! Message received did not match sent!\n%s != %s\n",
				p2p_message.To_String(), received_message.To_String())
		}
	default:
		t.Fatal("Error Peer2Peer! Message was never received!")
	}
}

func Test_Dependency_Resend(t *testing.T) {
	defer time.Sleep(time.Second)

	network_1 := New_P2P_Network("20007")
	network_2 := New_P2P_Network("20008")
	defer network_1.Close()
	defer network_2.Close()

	time.Sleep(time.Second)

	p2p_message := network_1.Create_Message("Hello!", MESSAGE)

	network_1.Broadcast(p2p_message)

	time.Sleep(time.Second)

	network_3 := New_P2P_Network("20009")
	defer network_3.Close()

	time.Sleep(time.Second)

	select {
	case received_message := <-network_2.Read_Channel:
		if p2p_message.To_String() != received_message.To_String() {
			t.Fatalf("Error Peer2Peer! Message received did not match sent!\n%s != %s\n",
				p2p_message.To_String(), received_message.To_String())
		}

		p2p_depended_message := network_2.Create_Message("Hello 2!", MESSAGE)
		p2p_depended_message.Depend_On(received_message)

		network_2.Broadcast(p2p_depended_message)

		time.Sleep(time.Second)

		select {
		case received_message_network_3 := <-network_3.Read_Channel:
			if p2p_message.To_String() != received_message_network_3.To_String() {
				t.Fatalf("Error Peer2Peer! Depended message received did not match sent!\n%s != %s\n",
					p2p_message.To_String(), received_message.To_String())
			}

			select {
			case received_depended_message := <-network_3.Read_Channel:
				if p2p_depended_message.To_String() != received_depended_message.To_String() {
					t.Fatalf("Error Peer2Peer! Depended message received did not match sent!\n%s != %s\n",
						p2p_message.To_String(), received_message.To_String())
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
	defer time.Sleep(time.Second) // Let the servers shut down before doing anything else...

	network_1 := New_P2P_Network("20005")
	network_2 := New_P2P_Network("20006")
	defer network_1.Close()
	defer network_2.Close()

	time.Sleep(time.Second)

	p2p_message := network_1.Create_Message("Hello!", MESSAGE)

	network_1.Broadcast(p2p_message)
	network_1.Broadcast(p2p_message)

	time.Sleep(time.Second)

	select {
	case received_message := <-network_2.Read_Channel:
		if p2p_message.To_String() != received_message.To_String() {
			t.Fatalf("Error Peer2Peer! Message received did not match sent!\n%s != %s\n",
				p2p_message.To_String(), received_message.To_String())
		}

		select {
		case spurious_message := <-network_2.Read_Channel:
			t.Fatalf("Error Peer2Peer, a message got sent twice! %s", spurious_message.To_String())
		default:
		}
	default:
		t.Fatal("Error Peer2Peer! Message was never received!")
	}
}
