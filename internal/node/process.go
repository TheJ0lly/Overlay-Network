package node

// In this file we will find the message-processing function.
// I decided to move them here for clarity and modularization of code.

import (
	"encoding/json"
	"log"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
)

func (currentNode *Node) processNewNodeJoinMessage(envelope *message.MessageEnvelope) {
	var msg message.NetNewNodeJoinMessage
	if err := envelope.GetMessageContent(&msg); err != nil {
		log.Printf("cannot get the content of NewNodeJoinMessage: %s", err)
		return
	}

	var newNode Node
	if err := json.Unmarshal(msg.NodeData, &newNode); err != nil {
		log.Printf("cannot deserialize the content of the new node: %s", err)
		return
	}

	// If the new node in the message is this node, skip the addition, just forward
	if newNode.Username == currentNode.Username {
		log.Printf("current node is the new node - skipping the update of the internal state")
		// forward
		return
	}

	if attachedNode := currentNode.findNodeInConnectionsByUsername(msg.ExistingNodeUsername); attachedNode != nil {
		attachedNode.Connections = append(attachedNode.Connections, &newNode)
		log.Printf("added new node to an existing known connection")
		// forward
		return
	}
}
