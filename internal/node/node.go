package node

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"slices"

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

func (n *Node) String() string {
	return fmt.Sprintf("[ip=%s, port=%d, conns=%v]", n.Ip, n.Port, n.Conns)
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
	switch msgEnv.Type {
	case message.NetNewNodeJoin:
		msg := message.NetNewNodeJoinMessage{}
		if err := json.Unmarshal(msgEnv.Data, &msg); err != nil {
			return fmt.Errorf("unmarshaling error for %s: %s", msgEnv.Type, err)
		}
		n.ProcessNetNewNodeJoinMessage(&msg, msgEnv.Sender)
		return nil
	case message.NetNewNodeJoinQuery:
		msg := message.NetNewNodeJoinQueryMessage{}
		if err := json.Unmarshal(msgEnv.Data, &msg); err != nil {
			return fmt.Errorf("unmarshaling error for %s: %s", msgEnv.Type, err)
		}
		n.ProcessNetNewNodeQueryMessage(&msg, msgEnv.Sender)
		return nil
	default:
		return fmt.Errorf("unknown message type: %d", msgEnv.Type)
	}
}

// processMessageGoroutine handles the queue of messages and processes them.
func (n *Node) processMessageGoroutine() {
	for {
		msg := <-n.Queue

		logging.LogInfo("started processing new message: type=%s data=%s sender=%v", msg.Type, msg.Data, msg.Sender)

		if err := n.handleMessage(&msg); err != nil {
			logging.LogError("%s", err)
		}
		logging.LogDebug("messages left in queue: %d", len(n.Queue))
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
			logging.LogInfo("queue is full (%d)! new message will be discarded", c)
			continue
		}

		// We push the Message Envelope in the channel for processing
		n.Queue <- env
	}
}

func (n *Node) SendMessageToIp(msg []byte, ip net.IP, port uint16) error {
	destNodeIp := net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))

	conn, err := net.Dial("tcp", destNodeIp)
	if err != nil {
		return fmt.Errorf("cannot send message to node %s - %s", destNodeIp, err)
	}
	defer conn.Close()

	size, err := conn.Write(msg)
	if err != nil {
		return fmt.Errorf("cannot send message to node %s - %s", destNodeIp, err)
	}

	if size != len(msg) {
		return fmt.Errorf("sent %d bytes - expected %d", size, len(msg))
	}

	return nil
}

func (n *Node) SendMessageToNode(msg []byte, other *Node) error {
	return n.SendMessageToIp(msg, other.Ip, other.Port)
}

func (n *Node) GetNodeAddress() string {
	return net.JoinHostPort(n.Ip.String(), fmt.Sprintf("%d", n.Port))
}

func (n *Node) ForwardMessage(env *message.MessageEnvelope, skipSenderList ...message.MessageSenderData) error {
	if len(n.Conns) == 0 {
		logging.LogInfo("message will not be forwarded")
		return fmt.Errorf("no other nodes connected to this node")
	}

	b, err := message.SerializeMessageEnvelope(env)
	if err != nil {
		logging.LogInfo("message will not be forwarded")
		return fmt.Errorf("cannot serialize original envelope")
	}

	for i := range n.Conns {
		conn := n.Conns[i]

		// If the sender of the envelope is within our known connections, we skip sending it to him.
		if conn.Ip.String() == env.Sender.Ip.String() && conn.Port == env.Sender.Port {
			logging.LogDebug("jumping over node: %s", conn.GetNodeAddress())
			continue
		}

		// Now we look through the list of senders to skip
		if slices.ContainsFunc(skipSenderList, func(sender message.MessageSenderData) bool {
			return conn.Ip.String() == sender.Ip.String() && conn.Port == sender.Port
		}) {
			logging.LogDebug("jumping over node: %s", conn.GetNodeAddress())
			continue
		}

		logging.LogInfo("forwarded message to node: %s", conn.GetNodeAddress())
		if err = n.SendMessageToNode(b, conn); err != nil {
			logging.LogError("could not forward message - %s", err)
		}
	}
	return nil
}

func (n *Node) FindNodeBasedOnIpAndPort(ip string, port uint16) *Node {
	if n.Ip.String() == ip && n.Port == port {
		return n
	}

	var toRet *Node
	for i := range n.Conns {
		conn := n.Conns[i]
		if conn.Ip.String() == ip && conn.Port == port {
			return conn
		}

		if toRet = findNodeBasedOnIpAndPortInNode(conn, ip, port); toRet != nil {
			return toRet
		}
	}
	return nil
}

func findNodeBasedOnIpAndPortInNode(node *Node, ip string, port uint16) *Node {
	for i := range node.Conns {
		conn := node.Conns[i]
		if conn.Ip.String() == ip && conn.Port == port {
			return conn
		}

		if toRet := findNodeBasedOnIpAndPortInNode(conn, ip, port); toRet != nil {
			return toRet
		}
	}
	return nil
}
