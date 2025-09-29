package queue

import (
	"fmt"
	"log"
)

type Queue[T any] struct {
	Notify chan struct{}
	items  []T
}

func Create[T any](cap uint16) *Queue[T] {
	return &Queue[T]{Notify: make(chan struct{}, 1), items: make([]T, 0, cap)}
}

// Add function adds an incoming message envelope into the queue if there is space for it, otherwise it will discard it.
func (mq *Queue[T]) Add(item T) error {
	if len(mq.items)+1 > cap(mq.items) {
		return fmt.Errorf("queue is full")
	}
	mq.items = append(mq.items, item)

	// If all comms cases/ops are blocked, the runtime will run the default case, that's why this thing below works and it is non-blocking.
	// Seeing how the Notify channel is buffered at 1, if we send 2 notifications, the first will go, the second will block the routine, until it can send it.
	// I don't know if this is the most efficient way, both in memory and performance, but I'll take it for now XD.
	select {
	case mq.Notify <- struct{}{}:
		log.Printf("new item queued, NMN send")
	default:
	}
	return nil
}

// Empty will return true if the queue is empty, otherwise false.
func (mq *Queue[T]) Empty() bool {
	return len(mq.items) == 0
}

// GetNext returns the next message to be processed in the queue and a bool that signals if the queue was empty and returned zero-valued envelope.
func (mq *Queue[T]) GetNext() (T, bool) {
	var zero T
	if mq.Empty() {
		return zero, false
	}
	toret := mq.items[0]
	mq.items = mq.items[1:]
	return toret, true
}
