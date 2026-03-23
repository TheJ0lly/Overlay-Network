package queue

import (
	"fmt"
	"slices"
)

type MessageQueue[T any] struct {
	q []T
}

func Create[T any](cap uint16) MessageQueue[T] {
	return MessageQueue[T]{
		q: make([]T, 0, cap),
	}
}

func (mq *MessageQueue[T]) PopFront() T {
	toret := mq.q[0]
	mq.q = mq.q[1:]
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

func (mq *MessageQueue[T]) RemoveByFunc(del func(T) bool) {
	mq.q = slices.DeleteFunc(mq.q, del)
}

func (mq *MessageQueue[T]) Length() int {
	return len(mq.q)
}
