package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type PubSub struct {
	mut    sync.RWMutex
	subs   map[string][]chan string
	closed bool
}

func NewPubSub() *PubSub {
	ps := &PubSub{}
	ps.subs = make(map[string][]chan string)
	return ps
}

// Method to subscribe to a specific topic
func (ps *PubSub) Subscribe(topic string) <-chan string {
	ps.mut.Lock()
	defer ps.mut.Unlock()

	ch := make(chan string, 1)
	ps.subs[topic] = append(ps.subs[topic], ch)
	return ch
}

// Method to publish a message to all subscribers of a specific topic.
func (ps *PubSub) Publish(topic string, msg string) {
	ps.mut.RLock()
	defer ps.mut.RUnlock()

	if ps.closed {
		return
	}

	for _, ch := range ps.subs[topic] {
		go func(ch chan string) {
			ch <- msg
		}(ch)
	}
}

// Mthod to prevent further publications and signaling subscribers to close their channels.
func (ps *PubSub) Close() {
	ps.mut.Lock()
	defer ps.mut.Unlock()

	if !ps.closed {
		ps.closed = true
		for _, subs := range ps.subs {
			for _, ch := range subs {
				close(ch)
			}
		}
	}
}

func main() {
	shutdownContext, cancelShutdown := context.WithCancel(context.Background())
	defer cancelShutdown()

	pubsub := NewPubSub()

	var wg sync.WaitGroup

	// Subscribe to topics
	subscriber1 := pubsub.Subscribe("topic1")
	subscriber2 := pubsub.Subscribe("topic2")

	// Goroutine to publish messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 5; i++ {
			select {
			case <-shutdownContext.Done():
				return
			default:
				pubsub.Publish("topic1", fmt.Sprintf("Message %d for topic1", i))
				time.Sleep(time.Second)
			}
		}
		for i := 1; i <= 3; i++ {
			select {
			case <-shutdownContext.Done():
				return
			default:
				pubsub.Publish("topic2", fmt.Sprintf("Message %d for topic2", i))
				time.Sleep(time.Second)
			}
		}
		pubsub.Close()
	}()

	// Goroutine for subscriber 1
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range subscriber1 {
			fmt.Println("Subscriber 1:", msg)
		}
		fmt.Println("Subscriber 1 closed")
	}()

	// Goroutine for subscriber 2
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range subscriber2 {
			fmt.Println("Subscriber 2:", msg)
		}
		fmt.Println("Subscriber 2 closed")
	}()

	go func() {
		// Wait for an interrupt signal or a specified duration
		<-shutdownContext.Done()
		fmt.Println("Received signal to shutdown gracefully")

		cancelShutdown()

		wg.Wait()
		fmt.Println("Main function exiting")
	}()

	wg.Wait()
}
