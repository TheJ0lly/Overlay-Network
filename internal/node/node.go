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
// Maybe add an OwnerID (or whatever), to uniquely identify who sent what message
type Node struct {
	Ip    net.IP                       `json:"Ip"`
	Port  uint16                       `json:"Port"`
	Conns []*Node                      `json:"Conns"`
	Queue chan message.MessageEnvelope `json:"-"`
}

func Create(ip string, port uint16, connCap uint16, queueCap uint16) (*Node, error) {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return nil, fmt.Errorf("cannot create node due to invalid IP: %s", ip)
	}

	return &Node{
		Ip:    parsedIp,
		Port:  port,
		Conns: make([]*Node, 0, connCap),
		Queue: make(chan message.MessageEnvelope, queueCap),
	}, nil
}

// listen function returns a net.Listener to handle incoming connections.
func (n *Node) listen() (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("%s:%d", n.Ip, n.Port))
}

// handleMessage function acts as a dispatcher of the message to its correct handler.
func (n *Node) handleMessage(msgEnv *message.MessageEnvelope) error {
	defer fmt.Println()
	switch msgEnv.Type {
	case message.NetNewNodeJoin:
		msg := message.NetNewNodeJoinMessage{}
		if err := json.Unmarshal(msgEnv.Data, &msg); err != nil {
			return fmt.Errorf("unmarshaling error for %s: %s", msgEnv.Type, err)
		}
		n.ProcessNetNewNodeJoinMessage(&msg)
		return nil
	default:
		return fmt.Errorf("unknown message type: %d", msgEnv.Type)
	}
}

// processMessageGoroutine handles the queue of messages and processes them.
func (n *Node) processMessageGoroutine() {
	for {
		msg := <-n.Queue

		logging.LogInfo("Started processing new message:")
		logging.LogInfo("Type: %s", msg.Type)
		logging.LogInfo("Data: %s", msg.Data)

		if err := n.handleMessage(&msg); err != nil {
			logging.LogError("%s", err)
		}
	}
}

// MainLoop function runs the main loop of the node.
// For now, you can run the node with this function, or simply look inside it and copy the code and use it. :)
func (n *Node) MainLoop() error {
	l, err := n.listen()
	if err != nil {
		logging.LogError("%s", err)
		return err
	}

	logging.LogInfo("listening on: %s", l.Addr())

	go n.processMessageGoroutine()

	for {
		conn, err := l.Accept()
		if err != nil {
			logging.LogError("%s", err)
			return err
		}
		defer conn.Close()

		b, err := io.ReadAll(conn)
		if err != nil {
			logging.LogError("%s", err)
			return err
		}

		env := message.MessageEnvelope{}
		err = message.DeserializeMessageEnvelope(&env, b)

		if err != nil {
			// Maybe add some RESEND/ACK convention for messages TODO
			logging.LogError("%s", err)
		}

		// Don't know if channels can truly get past their limit, might as well use "==", TODO
		if c := cap(n.Queue); len(n.Queue) == c {
			logging.LogInfo("Queue is full (%d)! New message will be discarded", c)
			continue
		}

		// We push the Message Envelope in the channel for processing
		n.Queue <- env
	}
}
