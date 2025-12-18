package message

import (
	"encoding/json"
	"net"
)

type MessageType uint16

const (
	NetNewNodeJoin MessageType = iota
)

// MessageEnvelope covers the message such that it will be easier to find out what message type it contains.
type MessageEnvelope struct {
	Type MessageType `json:"Type"`
	Data []byte      `json:"Data"`
}

// SerializeMessageEnvelope takes a message envelope and turns it into a byte slice.
func SerializeMessageEnvelope(msgEnv *MessageEnvelope) ([]byte, error) {
	return json.Marshal(msgEnv)
}

// DeserializeMessageEnvelope takes the data coming from the network and stores it in a user-passed envelope.
func DeserializeMessageEnvelope(env *MessageEnvelope, data []byte) error {
	return json.Unmarshal(data, env)
}

// NetNewNodeJoinMessage is the message a node receives when a new node will try to join the network.
type NetNewNodeJoinMessage struct {
	NodeIp net.IP `json:"NodeIp"`
}
