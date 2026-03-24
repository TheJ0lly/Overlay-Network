package queue

import (
	"fmt"
	"slices"

	"github.com/TheJ0lly/Overlay-Network/internal/logging"
)

type MessageQueue[T any] struct {
	q []T
}

func Create[T any](cap uint16) MessageQueue[T] {
	logging.LogDebug("creating queue with capacity: %d", cap)
	return MessageQueue[T]{
		q: make([]T, 0, cap),
	}
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
