package message

import (
	"encoding/json"
)

type MessageType uint8

const (
	NewNodeJoinType MessageType = iota
)

type MessageEnvelope struct {
	Type MessageType     `json:"Type"`
	Data json.RawMessage `json:"Data"`
}

func (me *MessageEnvelope) GetMessageContent(container any) error {
	return json.Unmarshal(me.Data, container)
}

type NewNodeJoinMessage struct {
	ExistingNodeUsername string          `json:"ExistingNodeUsername"`
	NodeData             json.RawMessage `json:"NodeData"`
}
