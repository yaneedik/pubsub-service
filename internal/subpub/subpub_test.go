package subpub

import (
	"context"
	"testing"
	"time"
)

func TestPubSub(t *testing.T) {
	bus := NewSubPub()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	defer bus.Close(ctx)

	topic := "test_topic"
	received := make(chan string, 1)

	sub, err := bus.Subscribe(topic, func(msg interface{}) {
		if s, ok := msg.(string); ok {
			received <- s
		}
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	testMsg := "hello"
	if err := bus.Publish(topic, testMsg); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	
	select {
	case msg := <-received:
		if msg != testMsg {
			t.Errorf("Expected %q, got %q", testMsg, msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Message not received")
	}

	
	sub.Unsubscribe()
	if err := bus.Publish(topic, "should not receive"); err != nil {
		t.Fatalf("Publish after unsubscribe failed: %v", err)
	}

	select {
	case <-received:
		t.Error("Received message after unsubscribe")
	case <-time.After(100 * time.Millisecond):
		
	}
}