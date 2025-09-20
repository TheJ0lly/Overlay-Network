package node

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
	"github.com/TheJ0lly/Overlay-Network/internal/queue"
)

// This structure represents the information and data required by the current active node on a machine.
//
// TODO: Maybe in the future we will create another type of node for the PrimaryConnections node,
// in which we do not have Queue, Depth, and ProcessingMessage
type Node struct {
	Username            string  `json:"Username"`
	Connections         []*Node `json:"Connections"`
	ConnectionsCapacity uint16  `json:"ConnectionsCapacity"`
	IP                  net.IP  `json:"IP"`
	IsAlive             bool    `json:"IsAlive"`
	Queue               *queue.MessageQueue
	Depth               uint8
	ProcessingMessage   bool
	// REMOVE THIS WHEN MAKING THE NETWORK TEST SUITE
	KeepRunning bool
}

// NodeLoop represents the main loop of the current active node. It will handle everything, from incoming/queued messages to state changes.
func (currentNode *Node) NodeLoop() {
	for currentNode.KeepRunning {
		if currentNode.ProcessingMessage || currentNode.Queue.IsEmpty() {
			continue
		}
		// Here there may be some logic changes, based on different actions, so as of now we will leave it like this, even though it seems pointless :)
		// BUT, it is explicit
		envelope := currentNode.Queue.GetNext()
		currentNode.processMessage(&envelope)
	}
}

// processMessage will take the message envelope and based on the type inside the envelope, it will handle the message accordingly.
func (currentNode *Node) processMessage(envelope *message.MessageEnvelope) {
	switch envelope.Type {
	case message.NewNodeJoinType:
		currentNode.processNewNodeJoinMessage(envelope)
		currentNode.KeepRunning = false

	default:
		slog.Warn(fmt.Sprintf("unknown message type with value: %d", envelope.Type))
	}
}

// addNode is an utility function that deserializes the data of a new node and adds it to an existing node.
// This is merely an state-update function, as the logic of availability will be performed by the node that is performing the addition.
func addNode(existingNode *Node, newNodeData []byte) {
	var newNode Node
	if err := json.Unmarshal(newNodeData, &newNode); err != nil {
		slog.Warn(fmt.Sprintf("cannot deserialize the content of the new node: %s", err))
		return
	}
	existingNode.Connections = append(existingNode.Connections, &newNode)
}

// findNodeInConnectionsByUsername will search through all the nodes reachable from the current node, meaning the primary connections and their respective primary connections.
// Return, if found, the node that matches the username we look for
func (currentNode *Node) findNodeInConnectionsByUsername(username string) *Node {
	for _, otherNode := range currentNode.Connections {
		if otherNode.Username == username {
			return otherNode
		}
		if otherNodeChild := otherNode.findNodeInConnectionsByUsername(username); otherNodeChild != nil {
			return otherNodeChild
		}
	}
	return nil
}
