package UDP

import (
	"testing"
	"time"
)

func Test_Multicast_Multiple_Listeners(t *testing.T) {

	channel1 := NewUDPChannel()
	channel2 := NewUDPChannel()

	defer channel1.Close()
	defer channel2.Close()

	time.Sleep(time.Second)

	send_message := "TEST"

	channel1.Broadcast(send_message)

	time.Sleep(time.Millisecond)

	select {
	case message := <-channel2.ReadChannel:
		if message != send_message {
			t.Errorf("Message 1 received did not match message sent!\n%s != %s\n", message, send_message)
		}
	default:
		t.Error("Did not receive message from channel 1!")
	}

	channel2.Broadcast(send_message)

	time.Sleep(time.Millisecond)

	select {
	case message := <-channel1.ReadChannel:
		if message != send_message {
			t.Errorf("Message 2 received did not match message sent!\n%s != %s\n", message, send_message)
		}
	default:
		t.Error("Did not receive message from channel 2!")
	}
}
