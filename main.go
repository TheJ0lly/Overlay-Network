package main

import (
	"flag"

	"github.com/TheJ0lly/Overlay-Network/internal/logging"
	"github.com/TheJ0lly/Overlay-Network/internal/node"
)

// run to send a message from another terminal. Testing phase
// echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080

const defaultUninitInt = 0
const defaultUninitString = ""
const portMax = (1 << 16) - 1

func main() {
	ip := flag.String("ip", defaultUninitString, "The IP address to start the node on")
	port := flag.Uint("port", defaultUninitInt, "The port to start the node on")
	connsCap := flag.Uint("conncap", defaultUninitInt, "The maximum capacity for primary connections")
	queueCap := flag.Uint("queuecap", defaultUninitInt, "The maximum capacity of the message queue")
	flag.BoolVar(&logging.DebugFlag, "debug", false, "Turn on debug logging")

	flag.Parse()

	if *ip == defaultUninitString {
		logging.LogErrorWithExit("IP is an empty string")
	}
	if *port > portMax {
		logging.LogErrorWithExit("Port is bigger than the max value allowed %d", portMax)
	}
	if *connsCap == defaultUninitInt {
		logging.LogErrorWithExit("Conns capacity is 0 - must be greater than 0")
	}
	if *queueCap == defaultUninitInt {
		logging.LogErrorWithExit("Queue capacity is 0 - must be greater than 0")
	}

	currNode, err := node.Create(*ip, uint16(*port), uint16(*connsCap), uint16(*queueCap))

	if err != nil {
		logging.LogErrorWithExit("%s", err)
	}

	currNode.MainLoop()
}
