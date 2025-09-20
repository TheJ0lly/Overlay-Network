package queue

import (
	"log/slog"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
)

type MessageQueue struct {
	messages []message.MessageEnvelope
	capacity uint16
}

func CreateQueue(cap uint16) *MessageQueue {
	return &MessageQueue{messages: make([]message.MessageEnvelope, 0, cap), capacity: cap}
}

func (mq *MessageQueue) AddToQueue(msg message.MessageEnvelope) {
	if len(mq.messages) < int(mq.capacity) {
		mq.messages = append(mq.messages, msg)
		return
	}
	slog.Info("message queue is full, cannot add new message to queue")
}

func (mq *MessageQueue) GetNext() message.MessageEnvelope {
	nextMsg := mq.messages[0]
	mq.messages = mq.messages[1:]
	return nextMsg
}

func (mq *MessageQueue) IsEmpty() bool {
	return len(mq.messages) == 0
}
