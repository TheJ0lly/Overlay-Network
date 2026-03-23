package message

import (
	"encoding/json"
	"fmt"
	"net"
)

type MessageType uint16

const (
	NetNewNodeJoin MessageType = iota
	NetNewNodeJoinQuery
	NetLifeLine
)

func (mt MessageType) String() string {
	switch mt {
	case NetNewNodeJoin:
		return "NetNewNodeJoin"
	case NetNewNodeJoinQuery:
		return "NetNewNodeJoinQuery"
	case NetLifeLine:
		return "NetLifeLine"
	default:
		return "unknown"
	}
}

// MessageSenderData represents a pair of an IP and Port signaling who sent the message.
type MessageSenderData struct {
	Ip   net.IP `json:"Ip"`
	Port uint16 `json:"Port"`
}

// MessageEnvelope covers the message such that it will be easier to find out what message type it contains.
type MessageEnvelope struct {
	Type   MessageType       `json:"Type"`
	Data   json.RawMessage   `json:"Data"`
	Sender MessageSenderData `json:"Sender"`
}

// SerializeMessageEnvelope takes a message envelope and turns it into a byte slice.
func SerializeMessageEnvelope(msgEnv *MessageEnvelope) ([]byte, error) {
	return json.Marshal(msgEnv)
}

// DeserializeMessageEnvelope takes the data coming from the network and stores it in a user-passed envelope.
func DeserializeMessageEnvelope(env *MessageEnvelope, data []byte) error {
	return json.Unmarshal(data, env)
}

type SerializableMessage interface {
	Serialize() ([]byte, error)
}

func CreateMessageEnvelope(mt MessageType, msg SerializableMessage, s MessageSenderData) (MessageEnvelope, error) {
	if b, err := msg.Serialize(); err != nil {
		return MessageEnvelope{}, fmt.Errorf("failed to marshal message - %s", err)
	} else {
		return MessageEnvelope{
			Type:   mt,
			Data:   b,
			Sender: s,
		}, nil
	}
}

// NetNewNodeJoinMessage is the message a node receives when a new node has queried and find a place to attach.
type NetNewNodeJoinMessage struct {
	// The IP of the new node that is joining
	JoinedIp net.IP `json:"Ip"`

	// The IP of the node that we attach to
	AttachedIp net.IP `json:"AttachedIp"`

	// The port of the node that we attach to
	AttachedPort uint16 `json:"AttachedPort"`

	// The port of the new node that is joining
	JoinedPort uint16 `json:"Port"`
}

func (msg *NetNewNodeJoinMessage) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

// NetNewNodeJoinQueryMessage is the message a node sends when joining a network for the FIRST TIME ever. It will also be used for RTT.
// The fields represent the data of the sending node, since this message is used as a request and a response.
type NetNewNodeJoinQueryMessage struct {
	Ip        net.IP `json:"Ip"`
	Port      uint16 `json:"Port"`
	Timestamp int64  `json:"Timestamp"`
}

func (msg *NetNewNodeJoinQueryMessage) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type NetLifeLineMessage struct{}

// NetLifeLineMessage is a message that will be sent periodically to let the other nodes that this node is alive.
// The only data needed, as of now, for this type of message is the message type and the sender, so that we can link with the Death Certificate.
// We can return (nil, nil) for now as we do not require any data in the actual message.
func (msg *NetLifeLineMessage) Serialize() ([]byte, error) {
	return nil, nil
}
