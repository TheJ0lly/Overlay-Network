package node

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
)

func createNetNewNodeJoinMessage(ip string) message.NetNewNodeJoinMessage {
	return message.NetNewNodeJoinMessage{
		Ip:       net.ParseIP(ip),
		Port:     8080,
		ConnsCap: 2,
	}
}

func TestProcessNetNewNodeJoinMessage(t *testing.T) {
	currNode, err := Create("127.0.0.1", 8080, 1, 1)
	if err != nil {
		t.Error("could not create node")
	}

	b, err := json.Marshal(createNetNewNodeJoinMessage("127.0.0.2"))
	if err != nil {
		t.Error("could not marshal message")
	}

	env := message.MessageEnvelope{
		Type: message.NetNewNodeJoin,
		Data: b,
	}

	err = currNode.handleMessage(&env)
	if err != nil {
		t.Errorf("could not handle message - %s", err)
	}

	if len(currNode.Conns) != 1 {
		t.Error("node should have one primary connection")
	}
}
