package node

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	Username            string              `json:"Username"`
	Connections         []*Node             `json:"Connections"`
	ConnectionsCapacity uint16              `json:"ConnectionsCapacity"`
	IP                  net.IP              `json:"IP"`
	IsAlive             bool                `json:"IsAlive"`
	Queue               *queue.MessageQueue `json:"-"`
	Depth               uint8               `json:"-"`
	Stop                chan struct{}       `json:"-"`
}

// Create creates and returns a Node. In the case where the IP is invalid the function will return nil.
func Create(username string, ip string, connCap uint16) *Node {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		log.Printf("IP used to create node is invalid: %s", ip)
		return nil
	}

	if connCap < 1 {
		log.Print("connection capacity is set to 0")
		return nil
	}

	return &Node{Username: username, IP: parsedIp, ConnectionsCapacity: connCap, Queue: queue.Create(connCap)}
}

// RunNodeLoop represents the main loop of the current active node. It will handle everything, from incoming/queued messages to state changes.
//
// IT SHOULD BE RUN AS A GOROUTINE. USE `GO` BEFORE THE CALL :)
func (currentNode *Node) RunNodeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			currentNode.Stop <- struct{}{}
			return
		case <-currentNode.Queue.Notify:
			for {
				envelope, exists := currentNode.Queue.GetNext()
				if !exists {
					break
				}
				currentNode.processMessage(&envelope)
			}
		}
	}
}

// processMessage will take the message envelope and based on the type inside the envelope, it will handle the message accordingly.
func (currentNode *Node) processMessage(envelope *message.MessageEnvelope) {
	log.Printf("processing next message of type: %s", envelope.Type.String())
	switch envelope.Type {
	case message.NewNodeJoinType:
		currentNode.processNewNodeJoinMessage(envelope)
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
// Return, if found, the node that matches the username we look for, otherwise `nil`.
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
