package message

import (
	"encoding/json"
	"fmt"
)

type MessageType uint8

const (
	// Network messages
	NetNewNodeJoinType MessageType = iota
	NetQueryPublicIpReqType
	NetQueryPublicIpRespType

	// System messages
	SysCreateNewNodeType
	SysStartNodeType
	SysStopNodeType
)

func (mt MessageType) String() string {
	switch mt {
	case NetNewNodeJoinType:
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
	case *NetNewNodeJoinMessage:
		if me.Type != NetNewNodeJoinType {
			goto errorMsg
		}
	case *NetQueryPublicIpReq: // nothing to handle
	case *NetQueryPublicIpResp:
		if me.Type != NetQueryPublicIpRespType {
			goto errorMsg
		}
	case *SysCreateNewNode:
		if me.Type != SysCreateNewNodeType {
			goto errorMsg
		}
	case *SysStartNode:
		if me.Type != SysStartNodeType {
			goto errorMsg
		}
	case *SysStopNode:
		if me.Type != SysStopNodeType {
			goto errorMsg
		}
	default:
		return fmt.Errorf("unknown message type - got: %s - expected: %s", t, me.Type.String())
	}
	return json.Unmarshal(me.Data, container)

errorMsg:
	return fmt.Errorf("passed container does not match the message type: %s", me.Type.String())
}
