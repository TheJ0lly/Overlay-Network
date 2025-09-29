package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
	"github.com/TheJ0lly/Overlay-Network/internal/networkutils"
	"github.com/TheJ0lly/Overlay-Network/internal/node"
	"github.com/TheJ0lly/Overlay-Network/internal/queue"
)

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

	// =============== Get public IP ===============
	var publicIP net.IP

	if *useDns {
		fmt.Printf("using Domain Name Servers to get public IP\n")
		publicIP = networkutils.GetIPFromDNS()
	} else if *connectionIp != "" && *connectionPort != 0 {
		mockNode := node.Node{
			IP:   net.ParseIP(*connectionIp),
			Port: uint16(*connectionPort),
		}
		// Need to create another type of NET message that queries for the public IP
		networkutils.SendMessage(&mockNode, message.MessageEnvelope{})
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
	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Printf("error occurred while reading the directory of the binary: %s\n", err)
		return
	}

	foundFile := false
	userFile := fmt.Sprintf("%s_%s", *username, *network)
	for _, f := range files {
		if f.Name() == userFile {
			foundFile = true
			break
		}
	}

	var currentNode node.Node

	if !foundFile {
		fmt.Printf("found no file with %s\n", userFile)
		if !*newNode {
			fmt.Printf("the `new` flag is not used and node file does not exist - will stop now\n")
			return
		}
		fmt.Printf("will create a new node file: %s\n", userFile)

		currentNode = node.Node{
			Username:            *username,
			Connections:         make([]*node.Node, 0, *connectionCap),
			ConnectionsCapacity: uint16(*connectionCap),
			IP:                  publicIP,
			Port:                uint16(*port),
			IsAlive:             true,
			NetworkName:         *network,
			Queue:               queue.Create[message.MessageEnvelope](uint16(*messageQueueCap)),
			Depth:               uint8(*depthVision),
			Stop:                make(chan struct{}),
		}

		if err := node.MarshalToFile(&currentNode); err != nil {
			fmt.Printf("error while marshaling new node to file: %s\n", err)
			return
		}
	} else {
		// read the user_network file
	}

	// Node Loop
}
