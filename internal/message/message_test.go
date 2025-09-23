package message

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestMessageTypeString(t *testing.T) {
	if NewNodeJoinType.String() != "NewNodeJoin" {
		t.Fatal("incorrect output")
	}

	if MessageType(255).String() != "Unknown" {
		t.Fatal("incorrect output")
	}
}

func TestGetContentMessageWithCorrectContainer(t *testing.T) {
	nnj := NewNodeJoinMessage{ExistingNodeUsername: "Node1", NodeData: json.RawMessage{'1', '2', '3'}}
	nnjB, err := json.Marshal(nnj)
	if err != nil {
		t.Fatalf("error while marshaling: %s", err)
	}

	env := MessageEnvelope{Type: NewNodeJoinType, Data: nnjB}

	nnj2 := NewNodeJoinMessage{}

	if err = env.GetMessageContent(&nnj2); err != nil {
		t.Fatalf("error while getting message content: %s", err)
	}

	if nnj2.ExistingNodeUsername != nnj.ExistingNodeUsername || slices.Compare(nnj.NodeData, nnj2.NodeData) != 0 {
		t.Fatalf("message contents do not match with the original")
	}
}

func TestGetContentMessageWithInorrectContainer(t *testing.T) {
	nnj := NewNodeJoinMessage{ExistingNodeUsername: "Node1", NodeData: json.RawMessage{'1', '2', '3'}}
	nnjB, err := json.Marshal(nnj)
	if err != nil {
		t.Fatalf("error while marshaling: %s", err)
	}

	env := MessageEnvelope{Type: NewNodeJoinType, Data: nnjB}
	type WrongMessageCont struct {
		AnInt int `json:"INT"`
	}
	nnj2 := WrongMessageCont{}

	err = env.GetMessageContent(&nnj2)
	if err == nil {
		t.Fatalf("there should be an error due to incorrect container used for unmarshaling")
	}
}

func TestGetContentMessageWithCorrectContainerAndIncorrectEnvelopeType(t *testing.T) {
	nnj := NewNodeJoinMessage{ExistingNodeUsername: "Node1", NodeData: json.RawMessage{'1', '2', '3'}}
	nnjB, err := json.Marshal(nnj)
	if err != nil {
		t.Fatalf("error while marshaling: %s", err)
	}

	env := MessageEnvelope{Type: 255, Data: nnjB}

	nnj2 := NewNodeJoinMessage{}

	if err = env.GetMessageContent(&nnj2); err == nil {
		t.Fatalf("there should be an error due to incorrect envelope type value")
	}
}
