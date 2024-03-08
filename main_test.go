package main

import "testing"

func TestPubSub(t *testing.T) {
	ps := NewPubSub()
	ch := ps.Subscribe("topic")
	ps.Publish("topic", "hello")
	msg := <-ch
	if msg != "hello" {
		t.Errorf("Expected hello, got %q", msg)
	}
}

func TestPubSubClose(t *testing.T) {
	ps := NewPubSub()
	ch := ps.Subscribe("topic")
	ps.Publish("topic", "hello")
	msg := <-ch
	if msg != "hello" {
		t.Errorf("Expected hello, got %q", msg)
	}
	ps.Close()
	_, ok := <-ch
	if ok {
		t.Errorf("Expected Closed Channel")
	}
}

func TestConcurrentPubSub(t *testing.T) {
	ps := NewPubSub()
	ch := ps.Subscribe("topic")
	go func() {
		for i := 0; i < 100; i++ {
			ps.Publish("topic", "hello")
		}
	}()

	for i := 0; i < 100; i++ {
		msg := <-ch
		if msg != "hello" {
			t.Errorf("Expected hello, got %q", msg)
		}
	}
}
