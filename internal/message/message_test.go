package message

import (
	"encoding/json"
	"slices"
	"testing"
)

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
