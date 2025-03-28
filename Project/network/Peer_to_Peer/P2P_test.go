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

	firstDependency := Dependency{owner, clock}

	// 1 more than dependency horizon, first dependency on list should be 1, but should be gone after crossing horizon.
	for i := 0; i < common.P2P_DEP_TIME_HORIZON+1; i++ {
		dependency := Dependency{owner, clock}

		handler.AddDependency(dependency)

		if !handler.HasDependency(dependency) {
			t.Fatal("Handler did not add dependency to list!\n")
		}

		clock.Event()
	}

	if handler.HasDependency(firstDependency) {
		t.Fatalf("First dependency was not removed! Map length: %d, Heap length: %d\n", len(handler.lookupMap), handler.minHeap.Len())
	}

	secondDependency := Dependency{owner, NewLamportClockFromString("2")}

	if !handler.HasDependency(secondDependency) {
		t.Fatalf("Second dependency was removed! Map length %d, Heap length %d\n", len(handler.lookupMap), handler.minHeap.Len())
	}

}

func TestLamportClockWraparound(t *testing.T) {
	clockHigh := LamportClock{common.LAMPORT_CLOCK_WRAPAROUND_UPPER_EDGE + 1}
	clockLow := LamportClock{common.LAMPORT_CLOCK_WRAPAROUND_LOWER_EDGE - 1}

	if !clockHigh.IsLessThan(clockLow) {
		t.Fatal("Wraparound clock is not returning true on wraparound!")
	}
}

func TestDependencyWraparound(t *testing.T) {
	handler := NewDependencyHandler()
	clock := NewLamportClock()

	clock.time = common.LAMPORT_CLOCK_WRAPAROUND_UPPER_EDGE + 1

	lowTime := common.LAMPORT_CLOCK_WRAPAROUND_LOWER_EDGE - common.P2P_DEP_TIME_HORIZON // Avoid cyclical dependency

	for i := 0; i < common.P2P_DEP_TIME_HORIZON+1; i++ {
		dependency := Dependency{strconv.Itoa(clock.time), clock}

		handler.AddDependency(dependency)

		if !handler.HasDependency(dependency) {
			t.Fatal("Handler did not add dependency to list!\n")
		}

		clock.Event()
	}

	clock.time = lowTime

	for i := 0; i < common.P2P_DEP_TIME_HORIZON+1; i++ {
		dependency := Dependency{strconv.Itoa(clock.time), clock}

		handler.AddDependency(dependency)
		oldDependency := Dependency{strconv.Itoa(clock.time - 1), LamportClock{clock.time - 1}}

		if !handler.HasDependency(dependency) {
			t.Fatal("Handler did not add dependency to list!\n")
		}

		// Old dependency doesn't exist for i = 0.
		if i != 0 && !handler.HasDependency(oldDependency) {
			t.Fatalf("Handler did not keep the new dependency: %s!\n", oldDependency.ToString())
		}

		clock.Event()
	}
}

func TestP2PMessageString(t *testing.T) {
	senderField := "SENDER"
	typeField := MESSAGE
	timeField := NewLamportClockFromString("6")
	dependencyField := NewDependency("OTHER", NewLamportClockFromString("3"))
	bodyField := "Hello from body!"

	testTCPMessage :=
		senderField + common.P2P_FIELD_DELIMINATOR +
			string(typeField) + common.P2P_FIELD_DELIMINATOR +
			timeField.String() + common.P2P_FIELD_DELIMINATOR +
			dependencyField.ToString() + common.P2P_FIELD_DELIMINATOR +
			bodyField

	p2pMessage := P2PMessageFromString(testTCPMessage)

	t.Logf("P2P_message generated: %s\n", p2pMessage.ToString())

	if p2pMessage.Sender != senderField {
		t.Fatalf("Sender field mismatch!\n%s != %s\n", p2pMessage.Sender, senderField)
	}

	if string(p2pMessage.Type) != string(typeField) {
		t.Fatalf("Type field mismatch!\n%s != %s\n", p2pMessage.Type, typeField)
	}

	if p2pMessage.Time.String() != timeField.String() {
		t.Fatalf("Time field mismatch!\n%s != %s\n", p2pMessage.Time.String(), timeField.String())
	}

	if p2pMessage.dependency.ToString() != dependencyField.ToString() {
		t.Fatalf("Dependency field mismatch!\n%s != %s\n", p2pMessage.dependency.ToString(), dependencyField.ToString())
	}

	if p2pMessage.Message != bodyField {
		t.Fatalf("Body field mismatch!\n%s != %s\n", p2pMessage.Message, bodyField)
	}

	stringedP2PMessage := p2pMessage.ToString()

	if stringedP2PMessage != testTCPMessage {
		t.Fatalf("Stringed message:\n%s\n---\nDid not match origin string %s\n", stringedP2PMessage, testTCPMessage)
	}
}

func TestMessageHorizon(t *testing.T) {
	resolver := NewDependencyResolver()
	clock := NewLamportClock()

	// 1 more than horizon
	for i := 0; i < common.P2P_MSG_TIME_HORIZON+1; i++ {
		clock.Event()

		p2pMessage := NewP2PMessage("SENDER", MESSAGE, clock, "BODY")
		resolver.EmplaceNewMessage(p2pMessage)

		dependency := NewDependency("SENDER", clock)

		resolvedP2PMessage, ok := resolver.GetMessage(dependency)

		if !ok {
			t.Fatalf("Failed to get an ok on sending %s:\n", p2pMessage.ToString())
		}

		if p2pMessage.ToString() != resolvedP2PMessage.ToString() {
			t.Fatalf("Returned string was ok, but not correct! %s != %s\n",
				p2pMessage.ToString(), resolvedP2PMessage.ToString())
		}
	}

	firstDependency := NewDependency("SENDER", NewLamportClockFromString("1"))
	secondDependency := NewDependency("SENDER", NewLamportClockFromString("2"))

	message, ok := resolver.GetMessage(firstDependency)

	if ok {
		t.Fatalf("Received a message that was supposed to be out of horizon!\n Resolver returned: %s\n",
			message.ToString())
	}

	message, ok = resolver.GetMessage(secondDependency)

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
	network1 := NewP2PNetwork()
	network2 := NewP2PNetwork()
	defer network1.Close()
	defer network2.Close()

	time.Sleep(time.Second)

	p2pMessage := network1.CreateMessage("Hello!")

	network1.Broadcast(p2pMessage)

	time.Sleep(time.Second)

	select {
	case receivedMessage := <-network2.ReadChannel:
		if p2pMessage.ToString() != receivedMessage.ToString() {
			t.Fatalf("Error Peer2Peer! Message received did not match sent!\n%s != %s\n",
				p2pMessage.ToString(), receivedMessage.ToString())
		}
	default:
		t.Fatal("Error Peer2Peer! Message was never received!")
	}
}

func TestDependencyResend(t *testing.T) {
	network1 := NewP2PNetwork()
	network2 := NewP2PNetwork()
	defer network1.Close()
	defer network2.Close()

	time.Sleep(time.Second) // Let the networks connect to each other. Takes about 1.2x UDP SERVER LIFETIME (~300ms)

	p2pMessage := network1.CreateMessage("Hello!")

	network1.Broadcast(p2pMessage)

	network3 := NewP2PNetwork()
	defer network3.Close()

	time.Sleep(time.Second)

	select {
	case receivedMessage := <-network2.ReadChannel:
		if p2pMessage.ToString() != receivedMessage.ToString() {
			t.Fatalf("Error Peer2Peer! Message received did not match sent!\n%s != %s\n",
				p2pMessage.ToString(), receivedMessage.ToString())
		}

		p2pDependedMessage := network2.CreateMessage("Hello 2!")
		p2pDependedMessage.DependOn(receivedMessage)

		network2.Broadcast(p2pDependedMessage)

		time.Sleep(time.Second)

		select {
		case receivedMessageNetwork_3 := <-network3.ReadChannel:
			if p2pMessage.ToString() != receivedMessageNetwork_3.ToString() {
				t.Fatalf("Error Peer2Peer! Depended message received did not match sent!\n%s != %s\n",
					p2pMessage.ToString(), receivedMessage.ToString())
			}

			select {
			case receivedDependedMessage := <-network3.ReadChannel:
				if p2pDependedMessage.ToString() != receivedDependedMessage.ToString() {
					t.Fatalf("Error Peer2Peer! Depended message received did not match sent!\n%s != %s\n",
						p2pMessage.ToString(), receivedMessage.ToString())
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

func TestDoubleSend(t *testing.T) {
	network1 := NewP2PNetwork()
	network2 := NewP2PNetwork()
	defer network1.Close()
	defer network2.Close()

	time.Sleep(time.Second)

	p2pMessage := network1.CreateMessage("Hello!")

	network1.Broadcast(p2pMessage)
	network1.Broadcast(p2pMessage)

	time.Sleep(time.Second)

	select {
	case receivedMessage := <-network2.ReadChannel:
		if p2pMessage.ToString() != receivedMessage.ToString() {
			t.Fatalf("Error Peer2Peer! Message received did not match sent!\n%s != %s\n",
				p2pMessage.ToString(), receivedMessage.ToString())
		}

		select {
		case spuriousMessage := <-network2.ReadChannel:
			t.Fatalf("Error Peer2Peer, a message got sent twice! %s", spuriousMessage.ToString())
		default:
		}
	default:
		t.Fatal("Error Peer2Peer! Message was never received!")
	}
}
