package message

import (
	"encoding/json"
	"fmt"

	"github.com/TheJ0lly/Overlay-Network/internal/network"
)

type MessageType uint16

const (
	NetNewNodeJoin MessageType = iota
	NetNewNodeJoinConfirm
	NetNewNodeJoinQuery
	NetLifeLine
	NetDeathAnnouncement
)

func (mt MessageType) String() string {
	switch mt {
	case NetNewNodeJoin:
		return "NetNewNodeJoin"
	case NetNewNodeJoinConfirm:
		return "NetNewNodeJoinConfirm"
	case NetNewNodeJoinQuery:
		return "NetNewNodeJoinQuery"
	case NetLifeLine:
		return "NetLifeLine"
	case NetDeathAnnouncement:
		return "NetDeathAnnouncement"
	default:
		return "unknown"
	}
}

// MessageEnvelope covers the message such that it will be easier to find out what message type it contains.
type MessageEnvelope struct {
	Type   MessageType        `json:"Type"`
	Data   json.RawMessage    `json:"Data"`
	Sender network.IpPortPair `json:"Sender"`
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

func CreateMessageEnvelope(mt MessageType, msg SerializableMessage, sender network.IpPortPair) (MessageEnvelope, error) {
	if b, err := msg.Serialize(); err != nil {
		return MessageEnvelope{}, fmt.Errorf("failed to marshal message - %s", err)
	} else {
		return MessageEnvelope{
			Type:   mt,
			Data:   b,
			Sender: sender,
		}, nil
	}
}

func SerializeNewMessageEnvelope(mt MessageType, msg SerializableMessage, sender network.IpPortPair) ([]byte, error) {
	if b, err := msg.Serialize(); err != nil {
		return nil, fmt.Errorf("failed to serialize message data - %s", err)
	} else {
		return SerializeMessageEnvelope(&MessageEnvelope{
			Type:   mt,
			Data:   b,
			Sender: sender,
		})
	}
}

// NetNewNodeJoinMessage is the message a node receives when a new node has queried and find a place to attach.
type NetNewNodeJoinMessage struct {
	JoiningNode  network.IpPortPair `json:"JoiningNode"`
	AttachedNode network.IpPortPair `json:"AttachedNode"`
	ReplacedNode network.IpPortPair `json:"ReplacedNode"`
}

func (msg *NetNewNodeJoinMessage) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

// NetNewNodeJoinConfirmMessage is the message that a node will receive when a new node determines that this node is the best place to attach
type NetNewNodeJoinConfirmMessage struct {
	IsSuitable bool `json:"IsSuitable"`
}

func (msg *NetNewNodeJoinConfirmMessage) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

// NetNewNodeJoinQueryMessage is the message a node sends when joining a network for the FIRST TIME ever. It will also be used for RTT.
// The fields represent the data of the sending node, since this message is used as a request and a response.
type NetNewNodeJoinQueryMessage struct {
	NewNode   network.IpPortPair `json:"NewNode"`
	Timestamp int64              `json:"Timestamp"`
}

func (msg *NetNewNodeJoinQueryMessage) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

// TODO Add a field where it writes the IpPortPair which is supposed to be alive, because it's causing a lot of issues with the LIFELINE.
// The sender of the envelope should change each time, but the alive node should remain the same.
type NetLifeLineMessage struct {
	Node network.IpPortPair `json:"Node"`
}

// NetLifeLineMessage is a message that will be sent periodically to let the other nodes that this node is alive.
// The only data needed, as of now, for this type of message is the message type and the sender, so that we can link with the Death Certificate.
// We can return (nil, nil) for now as we do not require any data in the actual message.
func (msg *NetLifeLineMessage) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type NetDeathAnnouncementMessage struct {
	DeadNodes []network.IpPortPair `json:"DeadNodes"`
}

func (msg *NetDeathAnnouncementMessage) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}
