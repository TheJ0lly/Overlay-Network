package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
	"github.com/TheJ0lly/Overlay-Network/internal/networkutils"
	"github.com/TheJ0lly/Overlay-Network/internal/node"
	"github.com/TheJ0lly/Overlay-Network/internal/queue"
)

var nodeServicePathDir string

func SetNewNode(newNode *node.Node) error {
	if err := node.MarshalToFile(newNode, nodeServicePathDir); err != nil {
		return fmt.Errorf("error while marshaling new node to file: %s", err)
	}
	return nil
}

func GetExistentNode(username string, network string) (node.Node, error) {
	files, err := os.ReadDir(".")
	if err != nil {
		return node.Node{}, fmt.Errorf("error occurred while reading the directory of the binary: %s", err)
	}

	foundFile := false
	userFile := fmt.Sprintf("%s_%s", username, network)
	for _, f := range files {
		if f.Name() == userFile {
			foundFile = true
			break
		}
	}

	if !foundFile {
		return node.Node{}, fmt.Errorf("found no file with %s", userFile)
	}

	b, err := os.ReadFile(userFile)
	if err != nil {
		return node.Node{}, fmt.Errorf("error while reading node file: %s", err)
	}

	var existentNode node.Node

	if err := json.Unmarshal(b, &existentNode); err != nil {
		return node.Node{}, fmt.Errorf("error while unmarshaling node data: %s", err)
	}

	return existentNode, nil
}

func main() {
	username := flag.String("user", "", "The username of the node - unique for each node in each network")
	depthVision := flag.Uint("depthVision", 0, "The vision depth of the node - how many layers of node it can see over the primary connections")
	ip := flag.String("ip", "", "The IP of the current node - only use if there is a static public IP, otherwise leave empty")
	connectionIp := flag.String("connIp", "", "The public IP of the node to connect to")
	port := flag.Uint("port", 8080, "The port that will be used to listen for incoming messages")
	connectionPort := flag.Uint("connPort", 0, "The port of the node to connecting to")
	useDns := flag.Bool("dns", false, "Use the predefined Domain Name Servers to get the public IP - will change in the future to allow the passing of a DNS file in a separate flag to query user selected DNS")
	connectionCap := flag.Uint("connCap", 0, "The number of primary connections this node will store")
	messageQueueCap := flag.Uint("msgCap", 0, "The number of messages that the queue can store")
	newNode := flag.Bool("new", false, "This flag signals that this is a new node - will not overwrite an existing node")
	network := flag.String("network", "", "The name of the network to connect to - there may be the same username on different networks, used for differentiating users locally")
	help := flag.Bool("help", false, "Show usage and parameters")

	flag.Parse()

	if *help {
		fmt.Fprintf(os.Stderr, "Usage of nodeservice: - WARNING: Do not use this binary on its own, use ovneto tool\n")
		flag.PrintDefaults()
		return
	}

	nodeServicePath, err := os.Executable()
	if err != nil {
		fmt.Printf("could not get the ovneto absolute path: %s - cannot use any related feature\n", err)
		return
	}

	nodeServicePathDir, err = filepath.Abs(nodeServicePath)
	if err != nil {
		fmt.Printf("could not get the ovneto absolute path: %s - cannot use any related feature\n", err)
		return
	}

	nodeServicePathDir = filepath.Dir(nodeServicePathDir)
	fmt.Printf("the node service path: %s\n", nodeServicePath)

	// =============== Get public IP ===============
	var publicIP net.IP = nil

	if *useDns {
		fmt.Printf("using Domain Name Servers to get public IP\n")
		publicIP = networkutils.GetIPFromDNS()
	} else if *connectionIp != "" && *connectionPort != 0 {
		mockNode := node.Node{
			IP:   net.ParseIP(*connectionIp),
			Port: uint16(*connectionPort),
		}

		conn, err := networkutils.SendMessage(
			mockNode.GetNodeAddress(),
			message.MessageEnvelope{
				Type: message.NetQueryPublicIpReqType,
				Data: json.RawMessage{},
			}, 10*time.Second)

		if err != nil {
			fmt.Printf("could not send NetQueryPublicIpReq message: %s - will stop now\n", err)
			return
		}

		// Here we receive a NetQueryPublicResp
		envelope, err := networkutils.ReceiveMessage(conn)
		if err != nil {
			fmt.Printf("could not receive NetQueryPublicIpResp message: %s - will stop now\n", err)
			return
		}

		var queryResp message.NetQueryPublicIpResp
		if err := envelope.GetMessageContent(&queryResp); err != nil {
			fmt.Printf("cannot get the content of NetQueryPublicIpResp: %s - will stop now\n", err)
			return
		}

		publicIP = queryResp.PublicIP

	} else if *ip != "" && *port != 0 {
		publicIP = net.ParseIP(*ip)
	} else {
		fmt.Printf("insufficient information or wrong flag usage, thus cannot get the public IP - will stop now\n")
		return
	}

	if publicIP == nil {
		fmt.Printf("could not get the public IP - will stop now")
		return
	}

	// =============== Get/Set node data

	var currentNode node.Node

	if *newNode {
		currentNode = node.Node{
			Username:            *username,
			Connections:         make([]*node.Node, 0, *connectionCap),
			ConnectionsCapacity: uint16(*connectionCap),
			IP:                  publicIP,
			Port:                uint16(*port),
			IsAlive:             true,
			NetworkName:         *network,
			Queue:               queue.Create[message.MessageEnvelope](uint16(*messageQueueCap)),
			QueueCap:            uint16(*messageQueueCap),
			Depth:               uint8(*depthVision),
			Stop:                make(chan struct{}),
		}

		if err := SetNewNode(&currentNode); err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Printf("trying to get existent user: %s with network: %s\n", *username, *network)
		currentNode, err = GetExistentNode(*username, *network)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancelFunc := context.WithCancelCause(context.Background())

	// Create the message receiving function with mutex (who gets priority the message receiver or handler routine?)
	go currentNode.RunMessageQueueLoop(ctx)
	go currentNode.RunNodeLoop(ctx)

	select {
	case sig := <-signals:
		switch sig {
		case syscall.SIGTERM:
			cancelFunc(fmt.Errorf("SIGTERM has been sent to this node"))
		case syscall.SIGINT:
			cancelFunc(fmt.Errorf("SIGINT has been sent to this node"))
		}
	case <-currentNode.Stop:
		cancelFunc(fmt.Errorf("stop signal has been received"))
	}

	fmt.Printf("node stopped - reason: %s\n", context.Cause(ctx))
}
