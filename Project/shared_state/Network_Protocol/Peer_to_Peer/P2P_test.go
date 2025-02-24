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
			type_field + Constants.P2P_FIELD_DELIMINATOR +
			time_field.String() + Constants.P2P_FIELD_DELIMINATOR +
			dependency_field.To_String() + Constants.P2P_FIELD_DELIMINATOR +
			body_field

	p2p_message := P2P_Message_From_String(test_tcp_message)

	t.Logf("P2P_message generated: %s\n", p2p_message.To_String())

	if p2p_message.Sender != sender_field {
		t.Fatalf("Sender field mismatch!\n%s != %s\n", p2p_message.Sender, sender_field)
	}

	if string(p2p_message.Type) != type_field {
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
	network_1 := New_P2P_Network("20005")
	network_2 := New_P2P_Network("20006")

	time.Sleep(time.Second)

	sender_field := "SENDER"
	type_field := MESSAGE
	time_field := New_Lamport_Clock_From_String("6")
	dependency_field := New_Dependency("OTHER", New_Lamport_Clock_From_String("3"))
	body_field := "Hello from body!"

	test_tcp_message :=
		sender_field + Constants.P2P_FIELD_DELIMINATOR +
			type_field + Constants.P2P_FIELD_DELIMINATOR +
			time_field.String() + Constants.P2P_FIELD_DELIMINATOR +
			dependency_field.To_String() + Constants.P2P_FIELD_DELIMINATOR +
			body_field

	p2p_message := P2P_Message_From_String(test_tcp_message)

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
