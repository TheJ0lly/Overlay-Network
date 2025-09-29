package node

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
	"github.com/TheJ0lly/Overlay-Network/internal/queue"
)

// This structure represents the information and data required by the current active node on a machine.
//
// TODO: Maybe in the future we will create another type of node for the PrimaryConnections node,
// in which we do not have Queue, Depth, and ProcessingMessage
type Node struct {
	Username            string                                `json:"Username"`
	Connections         []*Node                               `json:"Connections"`
	ConnectionsCapacity uint16                                `json:"ConnectionsCapacity"`
	IP                  net.IP                                `json:"IP"`
	Port                uint16                                `json:"Port"`
	IsAlive             bool                                  `json:"IsAlive"`
	NetworkName         string                                `json:"-"`
	Queue               *queue.Queue[message.MessageEnvelope] `json:"-"`
	Depth               uint8                                 `json:"-"`
	Stop                chan struct{}                         `json:"-"`
}

// Create creates and returns a Node. In the case where the IP is invalid the function will return nil.
func Create(username string, ip string, connCap uint16, msgCap uint16) *Node {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		log.Printf("IP used to create node is invalid: %s", ip)
		return nil
	}

	if connCap < 1 {
		log.Print("connection capacity is set to 0")
		return nil
	}

	return &Node{Username: username, IP: parsedIp, ConnectionsCapacity: connCap, Queue: queue.Create[message.MessageEnvelope](msgCap)}
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
	case message.NetNewNodeJoinType:
		currentNode.processNewNodeJoinMessage(envelope)
	default:
		log.Printf("unknown message type with value: %d", envelope.Type)
	}
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

func MarshalToFile(currentNode *Node) error {
	b, err := json.Marshal(currentNode)
	if err != nil {
		return err
	}

	return os.WriteFile(fmt.Sprintf("%s_%s", currentNode.Username, currentNode.NetworkName), b, 0666)
}

func UnmarshalFromFile(file string) (*Node, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	currentNode := &Node{}

	return currentNode, json.Unmarshal(b, &currentNode)
}
