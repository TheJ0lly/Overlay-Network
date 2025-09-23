package message

import (
	"encoding/json"
	"fmt"
)

type MessageType uint8

const (
	NewNodeJoinType MessageType = iota
)

func (mt MessageType) String() string {
	switch mt {
	case NewNodeJoinType:
		return "NewNodeJoin"
	default:
		return "Unknown"
	}
}

type MessageEnvelope struct {
	Type MessageType     `json:"Type"`
	Data json.RawMessage `json:"Data"`
}

// GetMessageContent will unmarshal the message inside the envelope.
//
// WARNING: PASS THE ARGUMENT AS A POINTER WITH `&`, BECAUSE OF UNMARSHALING
func (me *MessageEnvelope) GetMessageContent(container any) error {
	// We need to check for pointers in the switch, otherwise it won't work
	switch t := container.(type) {
	case *NewNodeJoinMessage:
		if me.Type != NewNodeJoinType {
			return fmt.Errorf("passed container does not match the message type: %s", me.Type.String())
		}
	default:
		return fmt.Errorf("unknown message type - got: %s - expected: %s", t, me.Type.String())
	}
	return json.Unmarshal(me.Data, container)
}

type NewNodeJoinMessage struct {
	ExistingNodeUsername string          `json:"ExistingNodeUsername"`
	NodeData             json.RawMessage `json:"NodeData"`
}
