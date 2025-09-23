package node

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
)

func TestCreate(t *testing.T) {
	currNode := Create("Node1", "192.1.1.1", 2)
	if currNode == nil {
		t.Fatal("node should be created successfully")
	}
}

func TestCreateWithInvalidIp(t *testing.T) {
	currNode := Create("Node1", "192.1.1.256", 2)
	if currNode != nil {
		t.Fatal("node should not be created")
	}
}

func TestCreateWithInvalidConnectionCapacity(t *testing.T) {
	currNode := Create("Node1", "192.1.1.1", 0)
	if currNode != nil {
		t.Fatal("node should not be created - connection capacity is 0")
	}
}

func TestRunNodeLoopWithContextWithoutCancel_NormalBehavior(t *testing.T) {
	currNode := Create("Node1", "192.1.1.1", 2)
	if currNode == nil {
		t.Fatal("node should be created successfully")
	}
	ctx := context.WithoutCancel(context.Background())

	go currNode.RunNodeLoop(ctx)
	time.Sleep(1 * time.Second)
	ctx.Done()

	if err := context.Cause(ctx); err != nil {
		t.Fatalf("the node loop should terminate without error: %s", err)
	}
}

func TestRunNodeLoopWithContextWithCancelThenCancelWithError(t *testing.T) {
	currNode := Create("Node1", "192.1.1.1", 2)
	if currNode == nil {
		t.Fatal("node should be created successfully")
	}
	ctx, cancelFunc := context.WithCancelCause(context.Background())

	go currNode.RunNodeLoop(ctx)
	cancelFunc(fmt.Errorf("made up error"))
	time.Sleep(1 * time.Second)
	ctx.Done()

	if err := context.Cause(ctx); err == nil {
		t.Fatalf("the node loop should terminate with an error")
	}
}

func TestProcessMessageWithUnknownMessage(t *testing.T) {
	currNode := Create("Node1", "192.1.1.1", 2)
	if currNode == nil {
		t.Fatal("node should be created successfully")
	}

	mockNode := Create("Node2", "192.168.1.2", 3)
	b, err := json.Marshal(&mockNode)
	if err != nil {
		t.Fatalf("marshaling error for mock node: %s", err)
	}

	msg := message.NewNodeJoinMessage{
		ExistingNodeUsername: "Node1",
		NodeData:             b,
	}

	b, err = json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshaling error for message content: %s", err)
	}
	env := message.MessageEnvelope{
		Type: 255,
		Data: b,
	}

	currNode.Queue.Add(env)

	ctx := context.WithoutCancel(context.Background())
	go currNode.RunNodeLoop(ctx)
	time.Sleep(1 * time.Second)
	ctx.Done()

	if len(currNode.Connections) != 0 {
		t.Fatalf("there should be no connection")
	}
}

func TestProcessMessageWithNewNodeJoinMessage(t *testing.T) {
	currNode := Create("Node1", "192.1.1.1", 2)
	if currNode == nil {
		t.Fatal("node should be created successfully")
	}

	mockNode := Create("Node2", "192.168.1.2", 3)
	b, err := json.Marshal(&mockNode)
	if err != nil {
		t.Fatalf("marshaling error for mock node: %s", err)
	}

	msg := message.NewNodeJoinMessage{
		ExistingNodeUsername: "Node1",
		NodeData:             b,
	}

	b, err = json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshaling error for message content: %s", err)
	}
	env := message.MessageEnvelope{
		Type: message.NewNodeJoinType,
		Data: b,
	}

	currNode.Queue.Add(env)

	ctx := context.WithoutCancel(context.Background())
	go currNode.RunNodeLoop(ctx)
	time.Sleep(1 * time.Second)
	ctx.Done()

	if len(currNode.Connections) != 1 {
		t.Fatalf("there should be one connection")
	}
}

func TestProcessFunctionForNewNodeJoinMessageWhereCurrentNodeIsTheOneAttachedTo(t *testing.T) {
	currNode := Create("Node1", "192.1.1.1", 2)
	if currNode == nil {
		t.Fatal("node should be created successfully")
	}

	mockNode := Create("Node2", "192.168.1.2", 3)
	b, err := json.Marshal(&mockNode)
	if err != nil {
		t.Fatalf("marshaling error for mock node: %s", err)
	}

	msg := message.NewNodeJoinMessage{
		ExistingNodeUsername: "Node1",
		NodeData:             b,
	}

	b, err = json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshaling error for message content: %s", err)
	}
	env := message.MessageEnvelope{
		Type: message.NewNodeJoinType,
		Data: b,
	}

	currNode.processNewNodeJoinMessage(&env)

	if len(currNode.Connections) != 1 {
		t.Fatalf("there should be one connection")
	}
}

func TestProcessFunctionForNewNodeJoinMessageWhereMessageTypeIsInvalid(t *testing.T) {
	currNode := Create("Node1", "192.1.1.1", 2)
	if currNode == nil {
		t.Fatal("node should be created successfully")
	}

	mockNode := Create("Node2", "192.168.1.2", 3)
	b, err := json.Marshal(&mockNode)
	if err != nil {
		t.Fatalf("marshaling error for mock node: %s", err)
	}

	msg := message.NewNodeJoinMessage{
		ExistingNodeUsername: "Node1",
		NodeData:             b,
	}

	b, err = json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshaling error for message content: %s", err)
	}
	env := message.MessageEnvelope{
		Type: 255,
		Data: b,
	}

	currNode.processNewNodeJoinMessage(&env)

	if len(currNode.Connections) != 0 {
		t.Fatalf("there should be no connection")
	}
}

func TestProcessFunctionForNewNodeJoinMessageWhereConnectionNodeIsTheOneAttachedTo(t *testing.T) {
	currNode := Create("Node1", "192.1.1.1", 2)
	if currNode == nil {
		t.Fatal("node should be created successfully")
	}

	// We mock an existing connection to which the new mock node will attach to
	otherNode := Create("Node3", "192.168.1.3", 4)
	currNode.Connections = append(currNode.Connections, otherNode)

	mockNode := Create("Node2", "192.168.1.2", 3)
	b, err := json.Marshal(&mockNode)
	if err != nil {
		t.Fatalf("marshaling error for mock node: %s", err)
	}

	msg := message.NewNodeJoinMessage{
		ExistingNodeUsername: "Node3",
		NodeData:             b,
	}

	b, err = json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshaling error for message content: %s", err)
	}
	env := message.MessageEnvelope{
		Type: message.NewNodeJoinType,
		Data: b,
	}

	currNode.processNewNodeJoinMessage(&env)

	if len(currNode.Connections) != 1 || len(currNode.Connections[0].Connections) != 1 {
		t.Fatalf("there should be one connection to Node1 and one connection to Node3")
	}
}
