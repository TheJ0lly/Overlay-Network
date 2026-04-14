package node

import (
	"encoding/json"
	"net"
	"slices"
	"time"

	"github.com/TheJ0lly/Overlay-Network/internal/logging"
	"github.com/TheJ0lly/Overlay-Network/internal/message"
)

func (n *Node) processNetNewNodeJoinMessage(msg *message.NetNewNodeJoinMessage, sender message.IpPortPair) {
	// Here it is okay to create the node with queue and conn capacity 0, because this is a mock node.
	// It's queue won't be used. Maybe make another method? TODO
	newNode, err := Create(msg.JoiningNode.Ip.String(), msg.JoiningNode.Port, 0, 0)
	if err != nil {
		logging.LogError("failed to create new node object: %s", err)
		return
	}

	var skipNodeIp string = ""
	var skipNodePort uint16 = 0
	// It means we are the node that is being attached to, we need to skip the sender node
	// As they append us themselves.
	if msg.AttachedNode.Ip.String() == n.Ip.String() && msg.AttachedNode.Port == n.Port {
		logging.LogDebug("we are the node that is being attached to")
		skipNodeIp = newNode.Ip.String()
		skipNodePort = newNode.Port

		// If we receive a join message with us being the attached node, it means we can remove the entry from the ongoing join queries list
		n.Stat.JoinQueriesOngoing = slices.DeleteFunc(n.Stat.JoinQueriesOngoing, func(pair message.IpPortPair) bool {
			return message.CompareIpPortPair(message.IpPortPair{Ip: newNode.Ip, Port: newNode.Port}, pair)
		})

		if len(n.Conns) == cap(n.Conns) {
			logging.LogDebug("We are at max capacity, looking for dead nodes")
			if replacedNode := n.replaceFirstDeadNode(newNode); replacedNode != nil {
				msg.ReplacedNode = replacedNode.getIpPortPair()
			}
		} else {
			logging.LogDebug("added new node - %s", newNode)
			n.Conns = append(n.Conns, newNode)
			msg.ReplacedNode = message.NullIpPortPair
		}
		logging.LogDebug("1 attached node state - %s", n)

	} else {
		if attachNode := n.findNodeBasedOnIpAndPort(msg.AttachedNode.Ip.String(), msg.AttachedNode.Port); attachNode != nil {
			// If there is a replced node, it means we must find the node and replace its data
			if !message.CompareIpPortPair(message.NullIpPortPair, msg.ReplacedNode) {
				if replacedNode := attachNode.findNodeBasedOnIpAndPort(msg.ReplacedNode.Ip.String(), msg.ReplacedNode.Port); replacedNode != nil {
					*replacedNode = *newNode
				}
			} else {
				attachNode.Conns = append(attachNode.Conns, newNode)
				logging.LogDebug("added new node - %s", newNode)
				logging.LogDebug("2 attached node state - %s", attachNode)
			}
		}
	}

	env, err := message.CreateMessageEnvelope(
		message.NetNewNodeJoin,
		// Before creating the envelope, we should add the replaced dead node, if there is any
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

	go n.ForwardMessage(
		&env,
		message.IpPortPair{
			Ip:   net.ParseIP(skipNodeIp),
			Port: skipNodePort,
		},
		sender,
	)
}

func (n *Node) processNetNewNodeQueryMessage(msg *message.NetNewNodeJoinQueryMessage, sender message.IpPortPair) {
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

	if err = n.SendMessageToIp(b, msg.NewNode.Ip, msg.NewNode.Port); err != nil {
		logging.LogError("could not send join query response - %s", err)
	}

	// The forwarding begins
	if b, err = json.Marshal(msg); err != nil {
		logging.LogError("cannot marshal original query message - message will not be forwarded")
		return
	}

	// We put both the original sender(the node who's joining) and the one possibly forwards the message to us.
	// In the case of receiving the message directly from the joining node, the last 2 senders are the same.
	go n.ForwardMessage(
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
	)
}

func (n *Node) processNetLifeLineMessage(sender message.IpPortPair) {
	if nd := n.findNodeBasedOnIpAndPort(sender.Ip.String(), sender.Port); nd == nil {
		logging.LogDebug("could not find node: %s:%d", sender.Ip.String(), sender.Port)
		return
	} else {
		nd.Alive = true
		nd.LastTimeAlive = time.Now().UnixMilli()
	}
	logging.LogDebug("received lifeline for node: %s:%d", sender.Ip.String(), sender.Port)

}

func (n *Node) processDeathAnnouncementMessage(msg *message.NetDeathAnnouncementMessage, sender message.IpPortPair) {
	shouldUpdateTheOtherNodes := false
	for i := range msg.DeadNodes {
		deadNode := msg.DeadNodes[i]
		if node := n.findNodeBasedOnIpAndPort(deadNode.Ip.String(), deadNode.Port); node != nil {
			node.Alive = false
			shouldUpdateTheOtherNodes = true
			continue
		}
		logging.LogDebug("the dead node %v is not known", deadNode)
	}

	if shouldUpdateTheOtherNodes {
		if env, err := message.CreateMessageEnvelope(message.NetDeathAnnouncement, msg, message.IpPortPair{
			Ip:   n.Ip,
			Port: n.Port,
		}); err != nil {
			logging.LogError("could not recreate death announcement envelope: %s", err)
		} else {
			go n.ForwardMessage(&env, sender)
		}
	}
}

func (n *Node) processNetNewNodeJoinConfirmMessage(sender message.IpPortPair) {

	notSuitableEnv, err := message.CreateMessageEnvelope(
		message.NetNewNodeJoinConfirm,
		&message.NetNewNodeJoinConfirmMessage{
			IsSuitable: false,
		},
		message.IpPortPair{
			Ip:   n.Ip,
			Port: n.Port,
		},
	)
	if err != nil {
		logging.LogError("could not create join confirm envelope: %s", err)
		return
	}

	isSuitableEnv, err := message.CreateMessageEnvelope(
		message.NetNewNodeJoinConfirm,
		&message.NetNewNodeJoinConfirmMessage{
			IsSuitable: true,
		},
		message.IpPortPair{
			Ip:   n.Ip,
			Port: n.Port,
		},
	)
	if err != nil {
		logging.LogError("could not create join confirm envelope: %s", err)
		return
	}

	notSuitableBytes, err := message.SerializeMessageEnvelope(&notSuitableEnv)
	if err != nil {
		logging.LogError("could not serialize join confirm envelope: %s", err)
		return
	}

	suitableBytes, err := message.SerializeMessageEnvelope(&isSuitableEnv)
	if err != nil {
		logging.LogError("could not serialize join confirm envelope: %s", err)
		return
	}

	// Here I sense a bug, due to the fact that if a node indeed finishes the joing process before this, they should be a part of the new join query, but that adds a lot of concurrency problems.
	// Will think about it.
	if cap(n.Conns) > len(n.Conns) && len(n.Stat.JoinQueriesOngoing) != 0 && len(n.Stat.JoinQueriesOngoing)+len(n.Conns) >= cap(n.Conns) {
		logging.LogError("current node has the maximum allowed number of ongoing join queries - will not participate as a candidate")
		goto notSuitable
	}

	if cap(n.Conns) == len(n.Conns) {
		logging.LogDebug("capacity of primary connections is full! checking for dead nodes")
		if len(n.findExistingDeadNodes()) == 0 {
			logging.LogDebug("there is no dead node to replace")
			goto notSuitable
		}
	}

	if err = n.SendMessageToIp(suitableBytes, sender.Ip, sender.Port); err != nil {
		logging.LogError("could not send confirm message: %s", err)
	}
	n.Stat.JoinQueriesOngoing = append(n.Stat.JoinQueriesOngoing, sender)
	logging.LogDebug("sent confirm message with SUITABLE")
	return

notSuitable:
	if err = n.SendMessageToIp(notSuitableBytes, sender.Ip, sender.Port); err != nil {
		logging.LogError("could not send confirm message: %s", err)
	}
	logging.LogDebug("sent confirm message with NOT SUITABLE")
}
