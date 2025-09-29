package node

import (
	"encoding/json"
	"testing"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
)

func TestProcessFunctionForNewNodeJoinMessageWhereCurrentNodeIsTheOneAttachedTo(t *testing.T) {
	currNode := Create("Node1", "192.1.1.1", 2, 1)
	if currNode == nil {
		t.Fatal("node should be created successfully")
	}

	mockNode := Create("Node2", "192.168.1.2", 3, 1)
	b, err := json.Marshal(&mockNode)
	if err != nil {
		t.Fatalf("marshaling error for mock node: %s", err)
	}

	msg := message.NetNewNodeJoinMessage{
		ExistingNodeUsername: "Node1",
		NodeData:             b,
	}

	b, err = json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshaling error for message content: %s", err)
	}
	env := message.MessageEnvelope{
		Type: message.NetNewNodeJoinType,
		Data: b,
	}

	currNode.processNewNodeJoinMessage(&env)

	if len(currNode.Connections) != 1 {
		t.Fatalf("there should be one connection")
	}
}

func TestProcessFunctionForNewNodeJoinMessageWhereMessageTypeIsInvalid(t *testing.T) {
	currNode := Create("Node1", "192.1.1.1", 2, 1)
	if currNode == nil {
		t.Fatal("node should be created successfully")
	}

	mockNode := Create("Node2", "192.168.1.2", 3, 1)
	b, err := json.Marshal(&mockNode)
	if err != nil {
		t.Fatalf("marshaling error for mock node: %s", err)
	}

	msg := message.NetNewNodeJoinMessage{
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
	currNode := Create("Node1", "192.1.1.1", 2, 2)
	if currNode == nil {
		t.Fatal("node should be created successfully")
	}

	// We mock an existing connection to which the new mock node will attach to
	otherNode := Create("Node3", "192.168.1.3", 4, 2)
	currNode.Connections = append(currNode.Connections, otherNode)

	mockNode := Create("Node2", "192.168.1.2", 3, 2)
	b, err := json.Marshal(&mockNode)
	if err != nil {
		t.Fatalf("marshaling error for mock node: %s", err)
	}

	msg := message.NetNewNodeJoinMessage{
		ExistingNodeUsername: "Node3",
		NodeData:             b,
	}

	b, err = json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshaling error for message content: %s", err)
	}
	env := message.MessageEnvelope{
		Type: message.NetNewNodeJoinType,
		Data: b,
	}

	currNode.processNewNodeJoinMessage(&env)

	if len(currNode.Connections) != 1 || len(currNode.Connections[0].Connections) != 1 {
		t.Fatalf("there should be one connection to Node1 and one connection to Node3")
	}
}
