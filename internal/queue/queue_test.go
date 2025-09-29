package queue

import (
	"fmt"
	"testing"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
)

func createTestQueue[T any](capacity uint16) *Queue[T] {
	return Create[T](capacity)
}

func TestCreate(t *testing.T) {
	var queueCap uint16 = 2
	mq := Create[message.MessageEnvelope](queueCap)
	if len(mq.items) != 0 || cap(mq.items) != int(queueCap) {
		t.Fatalf("queue should be empty with a capacity of %d", queueCap)
	}
}

func TestEmptyWithNoMessage(t *testing.T) {
	mq := createTestQueue[message.MessageEnvelope](1)

	if mq.Empty() != true {
		t.Fatal("queue should be empty")
	}
}

func TestEmptyWithOneMessage(t *testing.T) {
	mq := createTestQueue[message.MessageEnvelope](1)
	env := message.MessageEnvelope{}

	mq.Add(env)

	if mq.Empty() != false {
		t.Fatal("queue should not be empty")
	}
}

func TestAddOneMessageWithCapacityOne(t *testing.T) {
	mq := createTestQueue[message.MessageEnvelope](1)
	env := message.MessageEnvelope{}
	mq.Add(env)

	if mq.Empty() {
		t.Fatal("queue should have one message")
	}
}

func TestAddTwoMessagesWithCapacityOne(t *testing.T) {
	mq := createTestQueue[message.MessageEnvelope](1)
	env1 := message.MessageEnvelope{}
	env2 := message.MessageEnvelope{}
	mq.Add(env1)
	err := mq.Add(env2)

	if err != nil {
		fmt.Printf("cannot queue message: %s\n", err)
	}

	if mq.Empty() {
		t.Fatal("queue should have one message")
	}

	if len(mq.items) != 1 {
		t.Fatal("queue should have one message")
	}
}

func TestGetNextWithOneMessage(t *testing.T) {
	mq := createTestQueue[message.MessageEnvelope](1)
	env := message.MessageEnvelope{}
	mq.Add(env)

	_, exists := mq.GetNext()
	if exists != true {
		t.Fatal("there should be one message in queue")
	}
}

func TestGetNextNoMessages(t *testing.T) {
	mq := createTestQueue[message.MessageEnvelope](1)

	_, exists := mq.GetNext()
	if exists != false {
		t.Fatal("queue should be empty, no message to get")
	}
}
