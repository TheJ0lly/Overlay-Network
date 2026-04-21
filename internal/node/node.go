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
	"github.com/TheJ0lly/Overlay-Network/internal/network"
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
	DepthVision    uint8                                       `json:"-"`
	Stat           Stats                                       `json:"-"`
}

type NodeIPPMap = map[string][]network.IpPortPair

func (n *Node) String() string {
	return fmt.Sprintf("[ip=%s, port=%d, conns=%v]", n.Ip, n.Port, n.Conns)
}

func CreatePrimaryConnectionNode(ipp network.IpPortPair) *Node {
	return &Node{
		Ip:   ipp.Ip,
		Port: ipp.Port,
	}
}

func Create(ip string, port uint16, connCap uint8, queueCap uint16) (*Node, error) {
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

func createIpPortPairMapForNode(n *Node, layers uint8, cont NodeIPPMap, skipNode *network.IpPortPair) {
	if layers == 0 {
		return
	}

	if len(n.Conns) == 0 {
		return
	}
	ipportPairs := make([]network.IpPortPair, 0, len(n.Conns))

	for i := range n.Conns {
		connIpp := n.Conns[i].GetIpPortPair()
		if skipNode != nil && network.CompareIpPortPair(*skipNode, connIpp) {
			continue
		}
		logging.LogDebug("gathering node for update info: %s", connIpp)
		ipportPairs = append(ipportPairs, connIpp)
		createIpPortPairMapForNode(n.Conns[i], layers-1, cont, skipNode)
	}

	cont[n.GetIpPortPair().Hash()] = ipportPairs
}

func putIpPortPairsAsNodesInNode(n *Node, layers uint8, cont NodeIPPMap, skipNodes ...network.IpPortPair) {
	if layers == 0 {
		return
	}
	nodeHash := n.GetIpPortPair().Hash()

	var nodesIpp []network.IpPortPair
	var ok bool
	// If we cannot find the nodes hash, it means it either does not have any connections, or the vision of the sender is limited
	if nodesIpp, ok = cont[nodeHash]; !ok {
		return
	}

	nodeConnCap := cap(n.Conns)

	// We reset the connections list, and add them once again
	n.Conns = make([]*Node, 0, nodeConnCap)

	for i := range nodesIpp {
		if slices.ContainsFunc(skipNodes, func(ipp network.IpPortPair) bool {
			return network.CompareIpPortPair(nodesIpp[i], ipp)
		}) {
			continue
		}
		conn := CreatePrimaryConnectionNode(nodesIpp[i])
		n.Conns = append(n.Conns, conn)
		putIpPortPairsAsNodesInNode(conn, layers-1, cont, skipNodes...)
	}
}

// listen function returns a net.Listener to handle incoming connections.
func (n *Node) listen() (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("%s:%d", n.Ip, n.Port))
}

func (n *Node) setLastAliveTimeForNode(pair network.IpPortPair, t int64) {
	for i := range n.Conns {
		if network.CompareIpPortPair(n.Conns[i].GetIpPortPair(), pair) {
			n.Conns[i].LastTimeAlive = t
			n.Conns[i].Alive = true
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
		msg := message.NetLifeLineMessage{}
		if err := json.Unmarshal(msgEnv.Data, &msg); err != nil {
			return fmt.Errorf("unmarshaling error for %s: %s", msgEnv.Type, err)
		}
		n.processNetLifeLineMessage(msg, msgEnv.Sender)
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
	case message.NetUpdate:
		msg := message.NetUpdateMessage{}
		if err := json.Unmarshal(msgEnv.Data, &msg); err != nil {
			return fmt.Errorf("unmarshaling error for %s: %s", msgEnv.Type, err)
		}
		n.processNetUpdateMessage(msg, msgEnv.Sender)
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

// checkQueueForLifelinesForDeadNodes will get the nodes marked as dead, and check if there are lifelines in the queue.
// This mechanism is for reducing the network congestion due to state changes that are yet to be processed.
// Meaning that when we mark a node as dead, we might have a message coming from that node in queue, thus we can remove the death announcement.
func (n *Node) checkQueueForLifelinesForDeadNodes(deadNodes []network.IpPortPair) []network.IpPortPair {
	return slices.DeleteFunc(deadNodes, func(pair network.IpPortPair) bool {
		return n.Queue.ContainsFunc(func(me message.MessageEnvelope) bool {
			val := network.CompareIpPortPair(me.Sender, pair)
			logging.LogDebug("found message in queue for possible dead node %v? - %v", pair, val)
			return val
		})
	})
}

func (n *Node) findExistingDeadNodes() []network.IpPortPair {
	var deadNodes []network.IpPortPair = nil
	for i := range n.Conns {
		pConn := n.Conns[i]
		if pConn.Alive == false {
			deadNodes = append(deadNodes, pConn.GetIpPortPair())
		}
	}
	return n.checkQueueForLifelinesForDeadNodes(deadNodes)
}

// findNewDeadNodes will get the IpPortPair of each node that has (time.Now - LastTimeAlive) > DeathTimer.
func (n *Node) findNewDeadNodes() []network.IpPortPair {
	d := time.Second * time.Duration(n.DeathTimer)
	now := time.Now().UnixMilli()

	var deadNodes []network.IpPortPair = nil
	for i := range n.Conns {
		pConn := n.Conns[i]
		if pConn.Alive == false {
			continue
		}

		if (now - pConn.LastTimeAlive) > d.Milliseconds() {
			logging.LogDebug("found possible dead node: %s", pConn)
			deadNodes = append(deadNodes, pConn.GetIpPortPair())
		}
	}
	return n.checkQueueForLifelinesForDeadNodes(deadNodes)
}

func (n *Node) setNodesDead(deadNodes []network.IpPortPair) {
	for i := range n.Conns {
		if slices.ContainsFunc(deadNodes, func(deadNode network.IpPortPair) bool {
			return network.CompareIpPortPair(deadNode, n.Conns[i].GetIpPortPair())
		}) && n.Conns[i].Alive == true {
			logging.LogDebug("new node has been marked as dead: %v - %v", n.Conns[i].Ip, n.Conns[i].Port)
			n.Conns[i].Alive = false
		}
	}
}

func (n *Node) sendLifeLineAnnouncement() {
	logging.LogDebug("starting lifeline announcement")

	env, err := message.CreateMessageEnvelope(
		message.NetLifeLine,
		&message.NetLifeLineMessage{Node: n.GetIpPortPair()},
		n.GetIpPortPair(),
	)

	if err != nil {
		logging.LogError("could not serialize lifeline message for periodical update")
		return
	}

	logging.LogDebug("sending lifeline")
	go n.ForwardMessage(&env)
}

func (n *Node) sendDeathAnnouncement(deadNodes []network.IpPortPair) {
	logging.LogDebug("starting death annoucement")

	env, err := message.CreateMessageEnvelope(message.NetDeathAnnouncement, &message.NetDeathAnnouncementMessage{
		DeadNodes: deadNodes,
	}, n.GetIpPortPair())

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
			n.sendLifeLineAnnouncement()
			n.LifeLineTicker.Reset(time.Duration(n.LifeLineTimer) * time.Second)
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

func (n *Node) GetNodeAddress() string {
	ipp := n.GetIpPortPair()
	return ipp.NetString()
}

func (n *Node) gatherNodesToSendTo() []network.IpPortPair {
	toRet := make([]network.IpPortPair, 0)
	for i := range n.Conns {
		// TODO Here we should create the logic for dead hopping
		if n.Conns[i].Alive {
			toRet = append(toRet, n.Conns[i].GetIpPortPair())
			continue
		}
	}
	return toRet
}

func (n *Node) ForwardMessage(env *message.MessageEnvelope, skipSenderList ...network.IpPortPair) {
	if len(n.Conns) == 0 {
		logging.LogError("cannot forward, no other nodes connected to this node")
		return
	}

	b, err := message.SerializeMessageEnvelope(env)
	if err != nil {
		logging.LogError("cannot forward, cannot serialize original envelope")
		return
	}

	destNodes := n.gatherNodesToSendTo()
	network.SendToMultipleDest(b, destNodes, skipSenderList, time.Duration(n.DeathTimer))
}

func findNodeByIpPortPairInNode(node *Node, ipp network.IpPortPair, layer uint8) *Node {
	if layer == 0 {
		return nil
	}

	if network.CompareIpPortPair(node.GetIpPortPair(), ipp) {
		return node
	}

	for i := range node.Conns {
		conn := node.Conns[i]
		connIpp := conn.GetIpPortPair()
		if network.CompareIpPortPair(connIpp, ipp) {
			return conn
		}

		if toRet := findNodeByIpPortPairInNode(conn, ipp, layer-1); toRet != nil {
			return toRet
		}
	}
	return nil
}

func (n *Node) GetIpPortPair() network.IpPortPair {
	return network.IpPortPair{
		Ip:   n.Ip,
		Port: n.Port,
	}
}

func (n *Node) replaceFirstDeadNode(newNode *Node) *Node {
	if idx := slices.IndexFunc(n.Conns, func(nod *Node) bool {
		logging.LogDebug("node %v is alive? %v", nod.GetIpPortPair(), nod.Alive)
		return !nod.Alive
	}); idx != -1 {
		oldNode := n.Conns[idx]
		n.Conns[idx] = newNode
		return oldNode
	}
	return nil
}
