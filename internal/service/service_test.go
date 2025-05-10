package service

import (
	"context"
	"testing"
	"time"

	"github.com/yaneedik/pubsub-service/internal/subpub"
	"google.golang.org/grpc"
)

func TestService(t *testing.T) {
	bus := subpub.NewSubPub()
	svc := NewPubSubService(bus)

	topic := "test_topic"
	testMsg := "test_message"

	stream := &mockStream{
		ctx:      context.Background(),
		messages: make(chan *Event, 1),
	}

	go func() {
		err := svc.Subscribe(&SubscribeRequest{Key: topic}, stream)
		if err != nil {
			t.Errorf("Subscribe failed: %v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	_, err := svc.Publish(context.Background(), &PublishRequest{
		Key:  topic,
		Data: testMsg,
	})
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case msg := <-stream.messages:
		if msg.Data != testMsg {
			t.Errorf("Expected %q, got %q", testMsg, msg.Data)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Message not received")
	}
}

type mockStream struct {
	grpc.ServerStream
	ctx      context.Context
	messages chan *Event
}

func (m *mockStream) Context() context.Context {
	return m.ctx
}

func (m *mockStream) Send(msg *Event) error {
	m.messages <- msg
	return nil
}