package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
	"github.com/TheJ0lly/Overlay-Network/internal/node"
	"github.com/TheJ0lly/Overlay-Network/internal/queue"
)

func main() {
	currentNode := node.Node{Username: "Node1", IP: net.ParseIP("192.168.1.2"), ConnectionsCapacity: 2, Queue: queue.CreateQueue(1), ProcessingMessage: false, KeepRunning: true}

	mock := node.Node{Username: "Node2", IP: net.ParseIP("192.168.1.1")}
	b, _ := json.Marshal(mock)

	msg := message.NewNodeJoinMessage{
		ExistingNodeUsername: "Node1",
		NodeData:             b,
	}

	b, _ = json.Marshal(msg)

	env := message.MessageEnvelope{
		Type: message.NewNodeJoinType,
		Data: b,
	}

	fmt.Println(currentNode.Connections) // empty
	currentNode.Queue.AddToQueue(env)
	fmt.Println(currentNode.Queue.IsEmpty()) // false
	currentNode.NodeLoop()
	fmt.Println(currentNode.Queue.IsEmpty()) // true
	fmt.Println(currentNode.Connections[0])  // Node2
}
