package node

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/TheJ0lly/Overlay-Network/internal/logging"
	"github.com/TheJ0lly/Overlay-Network/internal/message"
)

// The base structure for all nodes in the network.
type Node struct {
	Ip    net.IP  `json:"Ip"`
	Port  uint16  `json:"Port"`
	Conns []*Node `json:"Conns"`
}

func Create(ip string, port uint16, connCap uint16) (*Node, error) {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return nil, fmt.Errorf("cannot create node due to invalid IP")
	}

	return &Node{Ip: parsedIp, Port: port, Conns: make([]*Node, 0, connCap)}, nil
}

// Listen function returns a net.Listener to handle incoming connections.
func (n *Node) Listen() (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("%s:%d", n.Ip, n.Port))
}

// HandleMessage function acts as a primitive dispatcher of the message to its correct handler.
func (n *Node) HandleMessage(msgEnv *message.MessageEnvelope) error {
	switch msgEnv.Type {
	case message.NetNewNodeJoin:
		msg := message.NetNewNodeJoinMessage{}
		if err := json.Unmarshal(msgEnv.Data, &msg); err != nil {
			return fmt.Errorf("unmarshaling error for %s: %s\n", msgEnv.Type, err)
		}
		n.ProcessNetNewNodeJoinMessage(&msg)
		return nil
	default:
		return fmt.Errorf("unknown message type: %d", msgEnv.Type)
	}
}

// MainLoop function runs the main loop of the node.
// For now, you can run the node with this function, or simply look inside it and copy the code and use it. :)
func (n *Node) MainLoop() error {
	l, err := n.Listen()
	if err != nil {
		logging.LogError("%s", err)
		return err
	}

	logging.LogInfo("listening on: %s", l.Addr())

	for {
		conn, err := l.Accept()
		if err != nil {
			logging.LogError("%s", err)
			return err
		}

		b, err := io.ReadAll(conn)
		if err != nil {
			logging.LogError("%s", err)
			return err
		}

		env := message.MessageEnvelope{}
		err = message.DeserializeMessageEnvelope(&env, b)

		if err != nil {
			logging.LogError("%s", err)
			return err
		}

		// Here we will queue the message, and use a buffered channel to signal the arrival of a new message.
		// And another goroutine will process the messages.
		// For now, we will handle them as they come :)

		if err = n.HandleMessage(&env); err != nil {
			logging.LogError("%s", err)
			return err
		}
	}
}
