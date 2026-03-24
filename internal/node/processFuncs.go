package node

import (
	"encoding/json"
	"net"
	"time"

	"github.com/TheJ0lly/Overlay-Network/internal/logging"
	"github.com/TheJ0lly/Overlay-Network/internal/message"
)

func (n *Node) ProcessNetNewNodeJoinMessage(msg *message.NetNewNodeJoinMessage, sender message.IpPortPair) {
	// Here it is okay to create the node with queue and conn capacity 0, because this is a mock node.
	// It's queue won't be used. Maybe make another method? TODO
	newNode, err := Create(msg.JoinedNode.Ip.String(), msg.JoinedNode.Port, 0, 0)
	if err != nil {
		logging.LogError("failed to create new node object: %s", err)
		return
	}

	// We update the node that is being attached to, if we find it in our local view
	if attachNode := n.findNodeBasedOnIpAndPort(msg.AttachedNode.Ip.String(), msg.AttachedNode.Port); attachNode != nil {
		attachNode.Conns = append(attachNode.Conns, newNode)
		logging.LogDebug("added new node - %s", newNode)
		logging.LogDebug("attached node state - %s", attachNode)
	}

	env, err := message.CreateMessageEnvelope(
		message.NetNewNodeJoin,
		msg,
		message.IpPortPair{
			Ip:   n.Ip,
			Port: n.Port,
		},
	)
	if err != nil {
		logging.LogError("failed to serialize response to net join message - %s", err)
		return
	}

	var skipNodeIp string = ""
	var skipNodePort uint16 = 0
	// It means we are the node that is being attached to, we need to skip the sender node
	// As they append us themselves.
	if msg.AttachedNode.Ip.String() == n.Ip.String() && msg.AttachedNode.Port == n.Port {
		skipNodeIp = newNode.Ip.String()
		skipNodePort = newNode.Port
		logging.LogDebug("we are the node that is being attached to")
	}

	if err = n.ForwardMessage(
		&env,
		message.IpPortPair{
			Ip:   net.ParseIP(skipNodeIp),
			Port: skipNodePort,
		},
		sender,
	); err != nil {
		logging.LogInfo("%s", err)
		return
	}
}

func (n *Node) ProcessNetNewNodeQueryMessage(msg *message.NetNewNodeJoinQueryMessage, sender message.IpPortPair) {
	var env message.MessageEnvelope
	var b []byte
	var err error

	// This is the response we send to the query.
	if env, err = message.CreateMessageEnvelope(
		message.NetNewNodeJoinQuery,
		&message.NetNewNodeJoinQueryMessage{
			NewNode: message.IpPortPair{
				Ip:   n.Ip,
				Port: n.Port,
			},
			Timestamp: time.Now().UnixMilli(),
		},
		message.IpPortPair{
			Ip:   n.Ip,
			Port: n.Port,
		},
	); err != nil {
		logging.LogError("cannot marshal query response - will not proceed with new node query")
		return
	}

	if b, err = message.SerializeMessageEnvelope(&env); err != nil {
		logging.LogError("cannot marshal query response envelope - will not proceed with new node query")
		return
	}

	if cap(n.Conns) > len(n.Conns) {
		if err = n.SendMessageToIp(b, msg.NewNode.Ip, msg.NewNode.Port); err != nil {
			logging.LogError("could not send join query response - %s", err)
			return
		}
	} else {
		logging.LogInfo("current node has reached its maximum primary connection capacity - will not participate as a possible candidate")
	}

	// The forwarding begins
	if b, err = json.Marshal(msg); err != nil {
		logging.LogError("cannot marshal original query message - message will not be forwarded")
		return
	}

	// We put both the original sender(the node who's joining) and the one possibly forwards the message to us.
	// In the case of receiving the message directly from the joining node, the last 2 senders are the same.
	if err = n.ForwardMessage(
		&message.MessageEnvelope{
			Type: message.NetNewNodeJoinQuery,
			Data: b,
			Sender: message.IpPortPair{
				Ip:   n.Ip,
				Port: n.Port,
			},
		},
		message.IpPortPair{
			Ip:   msg.NewNode.Ip,
			Port: msg.NewNode.Port,
		},
		sender,
	); err != nil {
		logging.LogInfo("%s", err)
		return
	}
}

func (n *Node) ProcessNetLifeLineMessage(sender message.IpPortPair) {
	if nd := n.findNodeBasedOnIpAndPort(sender.Ip.String(), sender.Port); nd == nil {
		logging.LogDebug("could not find node: %s:%d", sender.Ip.String(), sender.Port)
		return
	} else {
		nd.Alive = true
		nd.LastTimeAlive = time.Now().UnixMilli()
	}
	logging.LogDebug("received lifeline for node: %s:%d", sender.Ip.String(), sender.Port)

}
