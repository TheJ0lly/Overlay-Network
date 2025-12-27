package main

import (
	"flag"

	"github.com/TheJ0lly/Overlay-Network/internal/logging"
	"github.com/TheJ0lly/Overlay-Network/internal/node"
)

// run to send a message from another terminal. Testing phase
// echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080

func main() {
	ip := flag.String("ip", "", "The IP address to start the node on")
	port := flag.Uint("port", 0, "The port to start the node on")
	connsCap := flag.Uint("conncap", 0, "The maximum capacity for primary connections")
	flag.BoolVar(&logging.Debug, "debug", false, "Turn on debug logging")

	flag.Parse()

	const portMax = (1 << 16) - 1

	if *ip == "" {
		logging.LogErrorWithExit("IP is an empty string")
	}
	if *port > portMax {
		logging.LogErrorWithExit("Port is bigger than the max value allowed %d", portMax)
	}
	if *connsCap == 0 {
		logging.LogErrorWithExit("Conns capacity is 0 - must be greater than 0")
	}

	currNode, err := node.Create(*ip, uint16(*port), uint16(*connsCap))

	if err != nil {
		logging.LogErrorWithExit("%s", err)
	}

	currNode.MainLoop()
}
