package node

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"slices"
	"time"

	"github.com/TheJ0lly/Overlay-Network/internal/logging"
	"github.com/TheJ0lly/Overlay-Network/internal/message"
	"github.com/TheJ0lly/Overlay-Network/internal/queue"
)

// The base structure for all nodes in the network.
// Maybe add an OwnerID (or whatever), to uniquely identify who sent what message
type Node struct {
	Ip             net.IP                                      `json:"Ip"`
	Port           uint16                                      `json:"Port"`
	Conns          []*Node                                     `json:"Conns"`
	Queue          queue.MessageQueue[message.MessageEnvelope] `json:"-"`
	Alive          bool                                        `json:"-"`
	LifeLineTimer  uint8                                       `json:"-"`
	LifeLineTicker *time.Ticker                                `json:"-"`
	DeathTimer     uint8                                       `json:"-"`
	LastTimeAlive  int64                                       `json:"-"`
	Stat           Stats                                       `json:"-"`
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
		Ip:            parsedIp,
		Port:          port,
		Conns:         make([]*Node, 0, connCap),
		Queue:         queue.Create[message.MessageEnvelope](queueCap),
		Alive:         true,
		LifeLineTimer: 0,
	}, nil
}

// listen function returns a net.Listener to handle incoming connections.
func (n *Node) listen() (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("%s:%d", n.Ip, n.Port))
}

func (n *Node) setLastAliveTimeForNode(pair message.IpPortPair, t int64) {
	for i := range n.Conns {
		connPair := message.IpPortPair{
			Ip:   n.Conns[i].Ip,
			Port: n.Conns[i].Port,
		}
		if message.CompareIpPortPair(connPair, pair) {
			n.Conns[i].LastTimeAlive = t
			logging.LogDebug("setting last time for node: %v - %v", pair, t)
		}
	}
}

// handleMessage function acts as a dispatcher of the message to its correct handler.
func (n *Node) handleMessage(msgEnv *message.MessageEnvelope) error {
	switch msgEnv.Type {
	case message.NetNewNodeJoin:
		msg := message.NetNewNodeJoinMessage{}
		if err := json.Unmarshal(msgEnv.Data, &msg); err != nil {
			return fmt.Errorf("unmarshaling error for %s: %s", msgEnv.Type, err)
		}
		n.processNetNewNodeJoinMessage(&msg, msgEnv.Sender)
		n.setLastAliveTimeForNode(msgEnv.Sender, time.Now().UnixMilli())
		return nil
	case message.NetNewNodeJoinQuery:
		msg := message.NetNewNodeJoinQueryMessage{}
		if err := json.Unmarshal(msgEnv.Data, &msg); err != nil {
			return fmt.Errorf("unmarshaling error for %s: %s", msgEnv.Type, err)
		}
		n.processNetNewNodeQueryMessage(&msg, msgEnv.Sender)
		n.setLastAliveTimeForNode(msgEnv.Sender, time.Now().UnixMilli())
		return nil
	case message.NetLifeLine:
		n.processNetLifeLineMessage(msgEnv.Sender)
		n.setLastAliveTimeForNode(msgEnv.Sender, time.Now().UnixMilli())
		return nil
	case message.NetDeathAnnouncement:
		msg := message.NetDeathAnnouncementMessage{}
		if err := json.Unmarshal(msgEnv.Data, &msg); err != nil {
			return fmt.Errorf("unmarshaling error for %s: %s", msgEnv.Type, err)
		}
		n.processDeathAnnouncementMessage(&msg, msgEnv.Sender)
		n.setLastAliveTimeForNode(msgEnv.Sender, time.Now().UnixMilli())
		return nil
	case message.NetNewNodeJoinConfirm:
		msg := message.NetNewNodeJoinConfirmMessage{}
		if err := json.Unmarshal(msgEnv.Data, &msg); err != nil {
			return fmt.Errorf("unmarshaling error for %s: %s", msgEnv.Type, err)
		}
		n.processNetNewNodeJoinConfirmMessage(msgEnv.Sender)
		n.setLastAliveTimeForNode(msgEnv.Sender, time.Now().UnixMilli())
		return nil
	default:
		return fmt.Errorf("unknown message type: %d", msgEnv.Type)
	}
}

// processMessageGoroutine handles the queue of messages and processes them.
func (n *Node) processMessageGoroutine() {
	for {
		n.Queue.Wait()
		msg := n.Queue.PopFront()

		logging.LogInfo("started processing new message: type=%s data=%s sender=%v", msg.Type, msg.Data, msg.Sender)

		if err := n.handleMessage(&msg); err != nil {
			logging.LogError("%s", err)
		}

		logging.LogInfo("finished processing message: type=%s data=%s sender=%v", msg.Type, msg.Data, msg.Sender)
		logging.LogDebug("messages left in queue: %d", n.Queue.Length())
	}
}

func (n *Node) resetLifelineTimer() {
	n.LifeLineTicker.Reset(time.Duration(n.LifeLineTimer) * time.Second)
}

func (n *Node) sendLife() {
	logging.LogDebug("starting lifeline announcement")

	env, err := message.CreateMessageEnvelope(
		message.NetLifeLine,
		&message.NetLifeLineMessage{},
		message.IpPortPair{
			Ip:   n.Ip,
			Port: n.Port,
		},
	)

	if err != nil {
		logging.LogError("could not serialize lifeline message for periodical update")
		return
	}

	// Here there are no nodes we should skip, as this is an initiator message coming from this node.
	logging.LogDebug("sending lifeline")
	go n.ForwardMessage(&env)
}

// checkQueueForLifelinesForDeadNodes will get the nodes marked as dead, and check if there are lifelines in the queue.
// This mechanism is for reducing the network congestion due to state changes that are yet to be processed.
// Meaning that when we mark a node as dead, we might have a message coming from that node in queue, thus we can remove the death announcement.
func (n *Node) checkQueueForLifelinesForDeadNodes(deadNodes []message.IpPortPair) []message.IpPortPair {
	return slices.DeleteFunc(deadNodes, func(pair message.IpPortPair) bool {
		return n.Queue.ContainsFunc(func(me message.MessageEnvelope) bool {
			val := message.CompareIpPortPair(me.Sender, pair)
			logging.LogDebug("found message in queue for node %v? - %v", pair, val)
			return val
		})
	})
}

func (n *Node) findExistingDeadNodes() []message.IpPortPair {
	var deadNodes []message.IpPortPair = nil
	for i := range n.Conns {
		pConn := n.Conns[i]
		if pConn.Alive == false {
			deadNodes = append(deadNodes, message.IpPortPair{
				Ip:   pConn.Ip,
				Port: pConn.Port,
			})
		}
	}
	return n.checkQueueForLifelinesForDeadNodes(deadNodes)
}

// findNewDeadNodes will get the IpPortPair of each node that has (time.Now - LastTimeAlive) > DeathTimer.
func (n *Node) findNewDeadNodes() []message.IpPortPair {
	d := time.Second * time.Duration(n.DeathTimer)
	now := time.Now().UnixMilli()

	var deadNodes []message.IpPortPair = nil
	for i := range n.Conns {
		pConn := n.Conns[i]
		if pConn.Alive == false {
			continue
		}

		if (now - pConn.LastTimeAlive) > d.Milliseconds() {
			logging.LogDebug("found possible dead node: %v", now-pConn.LastTimeAlive)
			deadNodes = append(deadNodes, message.IpPortPair{
				Ip:   pConn.Ip,
				Port: pConn.Port,
			})
		}
	}
	return n.checkQueueForLifelinesForDeadNodes(deadNodes)
}

func (n *Node) setNodesDead(deadNodes []message.IpPortPair) {
	for i := range n.Conns {
		connPair := message.IpPortPair{
			Ip:   n.Conns[i].Ip,
			Port: n.Conns[i].Port,
		}

		if slices.ContainsFunc(deadNodes, func(deadNode message.IpPortPair) bool {
			return message.CompareIpPortPair(deadNode, connPair)
		}) && n.Conns[i].Alive == true {
			logging.LogDebug("node has been marked as dead: %v - %v", n.Conns[i].Ip, n.Conns[i].Port)
			n.Conns[i].Alive = false
		}
	}
}

func (n *Node) sendDeathAnnouncement(deadNodes []message.IpPortPair) {
	logging.LogDebug("starting death annoucement")

	env, err := message.CreateMessageEnvelope(message.NetDeathAnnouncement, &message.NetDeathAnnouncementMessage{
		DeadNodes: deadNodes,
	}, message.IpPortPair{
		Ip:   n.Ip,
		Port: n.Port,
	})

	if err != nil {
		logging.LogError("could not create envelope for death announcement: %s", err)
		return
	}

	logging.LogInfo("sending death announcement for: %v", deadNodes)
	go n.ForwardMessage(&env, deadNodes...)
}

// periodicalMessagesLoop is a method that will run in parallel to the main loop, and it will be used as the main place where messages/protocols are initiated.
func (n *Node) periodicalMessagesLoop() {
	n.LifeLineTicker = time.NewTicker(time.Duration(n.LifeLineTimer) * time.Second)
	deathTicker := time.NewTicker(time.Duration(n.DeathTimer) * time.Second)

	for {
		select {
		case <-n.LifeLineTicker.C:
			n.sendLife()
			n.resetLifelineTimer()
		case <-deathTicker.C:
			if deadNodes := n.findNewDeadNodes(); deadNodes != nil {
				n.setNodesDead(deadNodes)
				n.sendDeathAnnouncement(deadNodes)
			}
			deathTicker.Reset(time.Duration(n.DeathTimer) * time.Second)
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
	go n.periodicalMessagesLoop()

	for {
		conn, err := l.Accept()
		if err != nil {
			logging.LogError("%s", err)
			return err
		}

		b, err := io.ReadAll(conn)
		if err != nil {
			logging.LogError("%s", err)
			conn.Close()
			return err
		}

		env := message.MessageEnvelope{}
		err = message.DeserializeMessageEnvelope(&env, b)

		if err != nil {
			// Maybe add some RESEND/ACK convention for messages TODO
			logging.LogError("%s", err)
			conn.Close()
			continue
		}

		if err = n.Queue.Append(env); err != nil {
			logging.LogInfo("message queue error: %s", err)
			conn.Close()
			continue
		}
		conn.Close()
		n.Queue.Notify()
	}
}

func (n *Node) SendMessageToIp(msg []byte, ip net.IP, port uint16) error {
	destNodeIp := net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))

	conn, err := net.DialTimeout("tcp", destNodeIp, time.Second*time.Duration(n.DeathTimer))
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

func (n *Node) ForwardMessage(env *message.MessageEnvelope, skipSenderList ...message.IpPortPair) {
	if len(n.Conns) == 0 {
		logging.LogError("cannot forward, no other nodes connected to this node")
		return
	}

	b, err := message.SerializeMessageEnvelope(env)
	if err != nil {
		logging.LogError("cannot forward, cannot serialize original envelope")
		return
	}

	for i := range n.Conns {
		conn := n.Conns[i]

		// If the sender of the envelope is within our known connections, we skip sending it to him.
		if (conn.Ip.String() == env.Sender.Ip.String() && conn.Port == env.Sender.Port) ||
			conn.Alive == false {
			logging.LogDebug("jumping over node: %s", conn.GetNodeAddress())
			continue
		}

		// Now we look through the list of senders to skip
		if slices.ContainsFunc(skipSenderList, func(sender message.IpPortPair) bool {
			return conn.Ip.String() == sender.Ip.String() && conn.Port == sender.Port
		}) {
			logging.LogDebug("jumping over node: %s", conn.GetNodeAddress())
			continue
		}

		if err = n.SendMessageToNode(b, conn); err != nil {
			logging.LogError("could not forward message - %s", err)
			// Here we should insert the dead hopping mechanism
			continue
		}
		logging.LogDebug("forwarded message to node: %s", conn.GetNodeAddress())
	}
}

func (n *Node) findNodeBasedOnIpAndPort(ip string, port uint16) *Node {
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

func (n *Node) getIpPortPair() message.IpPortPair {
	return message.IpPortPair{
		Ip:   n.Ip,
		Port: n.Port,
	}
}

func (n *Node) replaceFirstDeadNode(newNode *Node) *Node {
	if idx := slices.IndexFunc(n.Conns, func(nod *Node) bool {
		logging.LogDebug("Node %v is alive? %v", nod.getIpPortPair(), nod.Alive)
		return !nod.Alive
	}); idx != -1 {
		logging.LogDebug("replacing dead node %v with node %v", message.IpPortPair{
			Ip:   n.Conns[idx].Ip,
			Port: n.Conns[idx].Port,
		}, message.IpPortPair{
			Ip:   newNode.Ip,
			Port: newNode.Port,
		})
		oldNode := n.Conns[idx]
		n.Conns[idx] = newNode
		return oldNode
	}
	return nil
}
