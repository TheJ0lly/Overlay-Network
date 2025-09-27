package node

// In this file we will find the message-processing function.
// I decided to move them here for clarity and modularization of code.

import (
	"log"

	"github.com/TheJ0lly/Overlay-Network/internal/networkmessage"
)

func (currentNode *Node) processNewNodeJoinMessage(envelope *networkmessage.MessageEnvelope) {
	var msg networkmessage.NewNodeJoinMessage
	if err := envelope.GetMessageContent(&msg); err != nil {
		log.Printf("cannot get the content of NewNodeJoinMessage: %s", err)
		return
	}

	if currentNode.Username == msg.ExistingNodeUsername {
		addNode(currentNode, msg.NodeData)
		return
	}

	if attachedNode := currentNode.findNodeInConnectionsByUsername(msg.ExistingNodeUsername); attachedNode != nil {
		addNode(attachedNode, msg.NodeData)
		return
	}
}
