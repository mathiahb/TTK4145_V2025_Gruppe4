package UDP

import (
	"testing"
	"time"
)

func TestMulticastMultipleListeners(t *testing.T) {

	channel1 := NewUDPChannel()
	channel2 := NewUDPChannel()

	defer channel1.Close()
	defer channel2.Close()

	time.Sleep(time.Second)

	sendMessage := "TEST"

	channel1.Broadcast(sendMessage)

	time.Sleep(time.Millisecond)

	select {
	case message := <-channel2.ReadChannel:
		if message != sendMessage {
			t.Errorf("Message 1 received did not match message sent!\n%s != %s\n", message, sendMessage)
		}
	default:
		t.Error("Did not receive message from channel 1!")
	}

	channel2.Broadcast(sendMessage)

	time.Sleep(time.Millisecond)

	select {
	case message := <-channel1.ReadChannel:
		if message != sendMessage {
			t.Errorf("Message 2 received did not match message sent!\n%s != %s\n", message, sendMessage)
		}
	default:
		t.Error("Did not receive message from channel 2!")
	}
}
