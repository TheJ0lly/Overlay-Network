package queue

import (
	"fmt"
	"slices"
)

type MessageQueue[T any] struct {
	q      []T
	notify chan struct{}
}

func Create[T any](cap uint16) MessageQueue[T] {
	return MessageQueue[T]{
		q:      make([]T, 0, cap),
		notify: make(chan struct{}, 1),
	}
}

// Notify is used to send a notification to the queue, that there is at least one item.
func (mq *MessageQueue[T]) Notify() {
	select {
	case mq.notify <- struct{}{}:
	default:
	}
}

// Wait is used as a blocking operation until there is at least one item in the queue.
func (mq *MessageQueue[T]) Wait() {
	<-mq.notify
}

func (mq *MessageQueue[T]) PopFront() T {
	toret := mq.q[0]
	mq.q = slices.Delete(mq.q, 0, 1)
	return toret
}

func (mq *MessageQueue[T]) LookFront() T {
	toret := mq.q[0]
	return toret
}

func (mq *MessageQueue[T]) Append(item T) error {
	if len(mq.q) >= cap(mq.q) {
		return fmt.Errorf("queue is full (%d)! new message will be discarded", cap(mq.q))
	}

	mq.q = append(mq.q, item)
	return nil
}

func (mq *MessageQueue[T]) Insert(item T, idx int) error {
	if len(mq.q) >= cap(mq.q) {
		return fmt.Errorf("queue is full (%d)! new message will be discarded", cap(mq.q))
	}

	mq.q = slices.Insert(mq.q, idx, item)
	return nil
}

// FindAllByFunc returns a slice of copies of the objects inside the actual queue. The `find` function condition must return `true` for the item to be found.
func (mq *MessageQueue[T]) FindAllByFunc(find func(T) bool) []T {
	var sToRet []T

	for i := range mq.q {
		if find(mq.q[i]) {
			sToRet = append(sToRet, mq.q[i])
		}
	}

	return sToRet
}

func (mq *MessageQueue[T]) ContainsFunc(contains func(T) bool) bool {
	return slices.ContainsFunc(mq.q, contains)
}

func (mq *MessageQueue[T]) RemoveByFunc(del func(T) bool) {
	mq.q = slices.DeleteFunc(mq.q, del)
}

func (mq *MessageQueue[T]) Length() int {
	return len(mq.q)
}
