package subpub

import (
	"context"
	"errors"
	"sync"
)

var ErrBusClosed = errors.New("subpub: bus is closed")

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}

type subscription struct {
	subject string
	handler MessageHandler
	bus     *subPub
}

func (s *subscription) Unsubscribe() {
	s.bus.unsubscribe(s)
}

type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

type subPub struct {
	mu         sync.RWMutex
	subs       map[string][]*subscription
	wg         sync.WaitGroup
	closeChan  chan struct{}
	closed     bool
}

func NewSubPub() SubPub {
	return &subPub{
		subs:      make(map[string][]*subscription),
		closeChan: make(chan struct{}),
	}
}

func (b *subPub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, ErrBusClosed
	}

	sub := &subscription{
		subject: subject,
		handler: cb,
		bus:     b,
	}

	b.subs[subject] = append(b.subs[subject], sub)
	return sub, nil
}

func (b *subPub) unsubscribe(sub *subscription) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs, ok := b.subs[sub.subject]
	if !ok {
		return
	}

	for i, s := range subs {
		if s == sub {
			b.subs[sub.subject] = append(subs[:i], subs[i+1:]...)
			break
		}
	}

	if len(b.subs[sub.subject]) == 0 {
		delete(b.subs, sub.subject)
	}
}

func (b *subPub) Publish(subject string, msg interface{}) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return ErrBusClosed
	}

	subs, ok := b.subs[subject]
	if !ok {
		return nil
	}

	for _, sub := range subs {
		b.wg.Add(1)
		go func(s *subscription) {
			defer b.wg.Done()
			s.handler(msg)
		}(sub)
	}
	return nil
}

func (b *subPub) Close(ctx context.Context) error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	close(b.closeChan)
	b.mu.Unlock()

	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}