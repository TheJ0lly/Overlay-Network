package queue

import (
	"log"

	"github.com/TheJ0lly/Overlay-Network/internal/networkmessage"
)

type MessageQueue struct {
	Notify   chan struct{}
	messages []networkmessage.MessageEnvelope
}

func Create(cap uint16) *MessageQueue {
	return &MessageQueue{Notify: make(chan struct{}, 1), messages: make([]networkmessage.MessageEnvelope, 0, cap)}
}

// Add function adds an incoming message envelope into the queue if there is space for it, otherwise it will discard it.
func (mq *MessageQueue) Add(msg networkmessage.MessageEnvelope) {
	if len(mq.messages)+1 > cap(mq.messages) {
		log.Printf("message queue is full, cannot queue message with type: %s", msg.Type.String())
		return
	}
	mq.messages = append(mq.messages, msg)

	// If all comms cases/ops are blocked, the runtime will run the default case, that's why this thing below works and it is non-blocking.
	// Seeing how the Notify channel is buffered at 1, if we send 2 notifications, the first will go, the second will block the routine, until it can send it.
	// I don't know if this is the most efficient way, both in memory and performance, but I'll take it for now XD.
	select {
	case mq.Notify <- struct{}{}:
		log.Printf("new message queued, NMN send")
	default:
	}
}

// Empty will return true if the queue is empty, otherwise false.
func (mq *MessageQueue) Empty() bool {
	return len(mq.messages) == 0
}

// GetNext returns the next message to be processed in the queue and a bool that signals if the queue was empty and returned zero-valued envelope.
func (mq *MessageQueue) GetNext() (networkmessage.MessageEnvelope, bool) {
	if mq.Empty() {
		return networkmessage.MessageEnvelope{}, false
	}
	toret := mq.messages[0]
	mq.messages = mq.messages[1:]
	return toret, true
}
