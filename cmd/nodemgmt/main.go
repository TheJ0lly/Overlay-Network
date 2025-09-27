package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
	"github.com/TheJ0lly/Overlay-Network/internal/node"
	"github.com/TheJ0lly/Overlay-Network/internal/queue"
)

func main() {
	currentNode := node.Node{Username: "Node1", IP: net.ParseIP("192.168.1.2"), ConnectionsCapacity: 2, Queue: queue.Create(2)}

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

	currentNode.Queue.Add(env)

	// Node 3
	mock = node.Node{Username: "Node3", IP: net.ParseIP("192.168.1.3")}
	b, _ = json.Marshal(mock)

	msg = message.NewNodeJoinMessage{
		ExistingNodeUsername: "Node1",
		NodeData:             b,
	}

	b, _ = json.Marshal(msg)

	env = message.MessageEnvelope{
		Type: message.NewNodeJoinType,
		Data: b,
	}

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	currentNode.Queue.Add(env)

	ctx, cancelFunc := context.WithCancelCause(context.Background())
	go currentNode.RunNodeLoop(ctx)

	select {
	case <-currentNode.Stop:
		// For the other CLI tool which will come in the future,
		// that will send a message through localhost that ends the node, so that we do not need to kill it from task manager
	case rcvSig := <-signalChan:
		log.Printf("operating system signal received: %s - shutting down\n", rcvSig)
		cancelFunc(fmt.Errorf("os signal received"))
	}
}
