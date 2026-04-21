package node

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/TheJ0lly/Overlay-Network/internal/logging"
	"github.com/TheJ0lly/Overlay-Network/internal/message"
	"github.com/TheJ0lly/Overlay-Network/internal/network"
)

func (n *Node) processNetNewNodeJoinMessage(msg *message.NetNewNodeJoinMessage, sender network.IpPortPair) {
	// Here it is okay to create the node with queue and conn capacity 0, because this is a mock node.
	// It's queue won't be used. Maybe make another method? TODO
	newNode, err := Create(msg.JoiningNode.Ip.String(), msg.JoiningNode.Port, msg.JoiningNodeConnCap, 0)
	if err != nil {
		logging.LogError("failed to create new node object: %s", err)
		return
	}
	newNode.DepthVision = msg.JoiningNodeView
	logging.LogDebug("new node has depth: %d", newNode.DepthVision)

	var skipNodes []network.IpPortPair = []network.IpPortPair{sender}

	// It means we are the node that is being attached to, we need to skip the sender node
	// As they append us themselves.
	var attachedNode *Node
	if attachedNode = findNodeByIpPortPairInNode(n, msg.AttachedNode, n.DepthVision); attachedNode == nil {
		logging.LogInfo("couldn't find attached node %s in visible nodes", msg.AttachedNode.NetString())
		env, err := message.CreateMessageEnvelope(
			message.NetNewNodeJoin,
			msg,
			n.GetIpPortPair(),
		)
		if err != nil {
			logging.LogError("failed to serialize response to net join message - %s", err)
			return
		}

		go n.ForwardMessage(
			&env,
			skipNodes...,
		)
		return
	}

	var updatedNodeConns NodeIPPMap = make(NodeIPPMap)

	if network.CompareIpPortPair(attachedNode.GetIpPortPair(), n.GetIpPortPair()) {
		logging.LogDebug("we are the node that is being attached to")
		skipNodes = append(skipNodes, newNode.GetIpPortPair())
		attachedNode = n

		// If we receive a join message with us being the attached node, it means we can remove the entry from the ongoing join queries list
		n.Stat.JoinQueriesOngoing = slices.DeleteFunc(n.Stat.JoinQueriesOngoing, func(joinQueryOngoingPair network.IpPortPair) bool {
			return network.CompareIpPortPair(newNode.GetIpPortPair(), joinQueryOngoingPair)
		})

		if len(n.Conns) == cap(n.Conns) {
			if replacedNode := n.replaceFirstDeadNode(newNode); replacedNode != nil {
				// Here we should forward an update message to update the connections of the new node
				logging.LogDebug("replacing dead node %v with node %v", replacedNode.GetIpPortPair(), newNode.GetIpPortPair())
				msg.ReplacedNode = replacedNode.GetIpPortPair()
			}
		} else {
			logging.LogDebug("added new node - %s", newNode)
			n.Conns = append(n.Conns, newNode)
			msg.ReplacedNode = network.NullIpPortPair
		}
		// The manual addition of THIS node as a primary connection
		newNodeKnownConns := make([]network.IpPortPair, 0, 1)
		newNodeKnownConns = append(newNodeKnownConns, n.GetIpPortPair())
		updatedNodeConns[newNode.GetIpPortPair().Hash()] = newNodeKnownConns
		nIpp := n.GetIpPortPair()
		createIpPortPairMapForNode(n, newNode.DepthVision-1, updatedNodeConns, &nIpp)
		logging.LogDebug("creating update info for new node: %s", newNode)
	} else {
		// If there is a replced node, it means we must find the node and replace its data
		if !network.CompareIpPortPair(network.NullIpPortPair, msg.ReplacedNode) {
			if replacedNode := findNodeByIpPortPairInNode(attachedNode, msg.ReplacedNode, attachedNode.DepthVision); replacedNode != nil {
				*replacedNode = *newNode
			}
		} else {
			attachedNode.Conns = append(attachedNode.Conns, newNode)
			logging.LogDebug("added new node - %s", newNode)
		}
	}
	logging.LogDebug("attached node state - %s", attachedNode)

	env, err := message.CreateMessageEnvelope(
		message.NetNewNodeJoin,
		msg,
		n.GetIpPortPair(),
	)
	if err != nil {
		logging.LogError("failed to serialize response to net join message - %s", err)
		return
	}

	go n.ForwardMessage(
		&env,
		skipNodes...,
	)

	// If we do not have have a direct interaction with the new node, we won't be having to forward anything.
	if len(updatedNodeConns) == 0 {
		return
	}

	timeToWait := 100

	// Artificial timer so that we do not risk sending an update for an inexistent node
	logging.LogDebug("waiting for %d ms to send the update for the new node", timeToWait)
	time.Sleep(time.Duration(timeToWait) * time.Millisecond)

	updateMsg := message.NetUpdateMessage{
		UpdatedNode: newNode.GetIpPortPair(),
		Conns:       updatedNodeConns,
	}

	env, err = message.CreateMessageEnvelope(message.NetUpdate, &updateMsg, n.GetIpPortPair())
	if err != nil {
		logging.LogError("failed to create update message for new node - %s", err)
		return
	}

	go n.ForwardMessage(&env)
}

func (n *Node) processNetNewNodeQueryMessage(msg *message.NetNewNodeJoinQueryMessage, sender network.IpPortPair) {
	var b []byte
	var err error

	// This is the response we send to the query.
	if b, err = message.SerializeNewMessageEnvelope(
		message.NetNewNodeJoinQuery,
		&message.NetNewNodeJoinQueryMessage{
			NewNode:   n.GetIpPortPair(),
			Timestamp: time.Now().UnixMilli(),
		},
		n.GetIpPortPair(),
	); err != nil {
		logging.LogError("cannot marshal query response - will not proceed with new node query")
		return
	}

	if err = network.SendToDest(b, msg.NewNode, time.Duration(n.DeathTimer)); err != nil {
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
			Type:   message.NetNewNodeJoinQuery,
			Data:   b,
			Sender: n.GetIpPortPair(),
		},
		msg.NewNode,
		sender,
	)
}

func (n *Node) processNetLifeLineMessage(msg message.NetLifeLineMessage, sender network.IpPortPair) {
	var nd *Node
	if nd = findNodeByIpPortPairInNode(n, msg.Node, n.DepthVision); nd == nil {
		logging.LogDebug("could not find node: %s", msg.Node.NetString())
	} else {
		nd.Alive = true
		nd.LastTimeAlive = time.Now().UnixMilli()
	}
	logging.LogDebug("received lifeline for node: %s", sender.NetString())

	if env, err := message.CreateMessageEnvelope(message.NetLifeLine, &msg, n.GetIpPortPair()); err != nil {
		logging.LogError("could not recreate death announcement envelope: %s", err)
	} else {
		go n.ForwardMessage(&env, sender)
	}

}

func (n *Node) processDeathAnnouncementMessage(msg *message.NetDeathAnnouncementMessage, sender network.IpPortPair) {
	for i := range msg.DeadNodes {
		deadNode := msg.DeadNodes[i]
		if node := findNodeByIpPortPairInNode(n, deadNode, n.DepthVision); node != nil {
			node.Alive = false
			continue
		}
		logging.LogDebug("the dead node %v is not known", deadNode)
	}

	if env, err := message.CreateMessageEnvelope(message.NetDeathAnnouncement, msg, n.GetIpPortPair()); err != nil {
		logging.LogError("could not recreate death announcement envelope: %s", err)
	} else {
		go n.ForwardMessage(&env, sender)
	}
}

func (n *Node) processNetNewNodeJoinConfirmMessage(sender network.IpPortPair) {
	confirmMessageData := message.NetNewNodeJoinConfirmMessage{
		IsSuitable: true,
	}
	// Here I sense a bug, due to the fact that if a node indeed finishes the joing process before this, they should be a part of the new join query, but that adds a lot of concurrency problems.
	// Will think about it.
	if cap(n.Conns) > len(n.Conns) && len(n.Stat.JoinQueriesOngoing) != 0 && len(n.Stat.JoinQueriesOngoing)+len(n.Conns) >= cap(n.Conns) {
		logging.LogError("current node has the maximum allowed number of ongoing join queries - will not participate as a candidate")
		confirmMessageData.IsSuitable = false
	} else if cap(n.Conns) == len(n.Conns) {
		logging.LogDebug("capacity of primary connections is full! checking for dead nodes")
		if len(n.findExistingDeadNodes()) == 0 {
			logging.LogDebug("there is no dead node to replace")
			confirmMessageData.IsSuitable = false
		}
	}

	var err error
	var b []byte
	if b, err = message.SerializeNewMessageEnvelope(message.NetNewNodeJoinConfirm, &confirmMessageData, n.GetIpPortPair()); err != nil {
		logging.LogError("could not create join confirm envelope: %s", err)
		return
	}

	if err = network.SendToDest(b, sender, time.Duration(n.DeathTimer)); err != nil {
		logging.LogError("could not send confirm message: %s", err)
		return
	}
	n.Stat.JoinQueriesOngoing = append(n.Stat.JoinQueriesOngoing, sender)
	logging.LogDebug("sent confirm message with isSuitable=%v", confirmMessageData.IsSuitable)
}

func (n *Node) processNetUpdateMessage(msg message.NetUpdateMessage, sender network.IpPortPair) {
	updatedNode := findNodeByIpPortPairInNode(n, msg.UpdatedNode, n.DepthVision)
	if updatedNode == nil {
		logging.LogInfo("could not find the updated node")
	} else {
		logging.LogInfo("found the updated node: %v", updatedNode)
		logging.LogInfo("targeted node state before: %s", updatedNode)
		nodeIpp := n.GetIpPortPair()
		putIpPortPairsAsNodesInNode(updatedNode, updatedNode.DepthVision, msg.Conns, nodeIpp)
		logging.LogInfo("targeted node state after: %s", updatedNode)
	}

	env, err := message.CreateMessageEnvelope(message.NetUpdate, &msg, n.GetIpPortPair())
	if err != nil {
		logging.LogError("could not create update envelope: %s", err)
		return
	}

	go n.ForwardMessage(&env, sender)
}
