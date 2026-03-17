package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"time"

	"github.com/TheJ0lly/Overlay-Network/internal/logging"
	"github.com/TheJ0lly/Overlay-Network/internal/message"
	"github.com/TheJ0lly/Overlay-Network/internal/node"
)

const defaultUninitInt = 0
const defaultUninitString = ""
const portMax = (1 << 16) - 1

func JoinNewNetwork(currNode *node.Node, connectionIp *string, connectionPort *uint, port *uint, ip *string, connsCap *uint) {
	if *connectionIp == defaultUninitString && *connectionPort == defaultUninitInt {
		logging.LogErrorWithExit("one of these flags have been set, but not both - to join a new network use both flags \"connip\" + \"connport\"")
	}

	parsedIp := net.ParseIP(*connectionIp)
	initialTimestamp := time.Now().UnixMilli()
	var env message.MessageEnvelope
	var b []byte
	var err error
	var list net.Listener
	var conn net.Conn

	if env, err = message.CreateMessageEnvelope(
		message.NetNewNodeJoinQuery,
		&message.NetNewNodeJoinQueryMessage{
			Ip:        currNode.Ip,
			Port:      currNode.Port,
			Timestamp: initialTimestamp,
		},
		message.MessageSenderData{
			Ip:   currNode.Ip,
			Port: currNode.Port,
		},
	); err != nil {
		logging.LogErrorWithExit("could not create join query message - %s", err)
	}

	if b, err = message.SerializeMessageEnvelope(&env); err != nil {
		logging.LogErrorWithExit("could not marshal initial message envelope for joining a new network - %s", err)
	}

	sourceIp := net.JoinHostPort(*connectionIp, fmt.Sprintf("%d", *connectionPort))
	if err = currNode.SendMessageToIp(b, parsedIp, uint16(*connectionPort)); err != nil {
		logging.LogErrorWithExit("could not sent message envelope to node %s", sourceIp)
	}
	logging.LogInfo("sent message to %s: type=%s data=%s", sourceIp, env.Type, env.Data)

	if list, err = net.Listen("tcp", currNode.GetNodeAddress()); err != nil {
		logging.LogErrorWithExit("could not start listener for the initial message - %s", err)
	}

	timeoutChan := make(chan struct{}, 1)
	go func() {
		time.Sleep(time.Second * 10)
		timeoutChan <- struct{}{}
		list.Close()
	}()

	running := true
	var bestIp net.IP = nil
	var bestPort uint16 = 0
	bestTime := int64(math.MaxInt64)

	for running {
		select {
		case <-timeoutChan:
			logging.LogInfo("received timeout - closing join query window")
			running = false
			continue
		default:
			conn, err = list.Accept()
			if err != nil {
				logging.LogError("error while accepting incoming connections - %s", err)
				continue
			}

			b, err = io.ReadAll(conn)
			if err != nil {
				logging.LogError("error while reading from the connection - %s", err)
				continue
			}

			if err = message.DeserializeMessageEnvelope(&env, b); err != nil {
				logging.LogError("error while deserializing message envelope - %s", err)
				continue
			}

			msg := message.NetNewNodeJoinQueryMessage{}
			if err := json.Unmarshal(env.Data, &msg); err != nil {
				logging.LogError("error while deserializing message - %s", err)
				continue
			}

			if newTime := msg.Timestamp - initialTimestamp; newTime < bestTime {
				bestIp = net.ParseIP(msg.Ip.String())
				bestPort = msg.Port
				bestTime = newTime
				logging.LogDebug("found better node to attach to - ip: %s - port: %d", bestIp, bestPort)
			}
		}
	}
	if bestIp == nil || bestPort == 0 {
		logging.LogErrorWithExit("could not find a suitable node to attach to")
	}

	logging.LogInfo("found best node to attach to - ip: %s - port: %d", bestIp, bestPort)

	if env, err = message.CreateMessageEnvelope(
		message.NetNewNodeJoin,
		&message.NetNewNodeJoinMessage{
			AttachedIp:   net.ParseIP(bestIp.String()),
			AttachedPort: bestPort,
			JoinedIp:     net.ParseIP(*ip),
			JoinedPort:   uint16(*port),
		},
		message.MessageSenderData{
			Ip:   currNode.Ip,
			Port: currNode.Port,
		},
	); err != nil {
		logging.LogErrorWithExit("could not create net join message envelope - %s", err)
	}

	if b, err = message.SerializeMessageEnvelope(&env); err != nil {
		logging.LogErrorWithExit("could not serialize net join message - %s", err)
	}

	if err = currNode.SendMessageToIp(b, bestIp, bestPort); err != nil {
		logging.LogErrorWithExit("could not send net join message - %s", err)
	}
	logging.LogInfo("sent message to %s: type=%s data=%s", net.JoinHostPort(bestIp.String(), fmt.Sprintf("%d", bestPort)), env.Type, env.Data)

	// Maybe we replace this with a message, but maybe not
	if cap(currNode.Conns) > len(currNode.Conns) {
		if newNode, err := node.Create(bestIp.String(), bestPort, 0, 0); err != nil {
			logging.LogErrorWithExit("could not add the new node to the current node - %s", err)
		} else {
			currNode.Conns = append(currNode.Conns, newNode)
			logging.LogDebug("added new node - %s", newNode)
			logging.LogDebug("attached node state - %s", currNode)
		}
	}
}

func main() {
	ip := flag.String("ip", defaultUninitString, "the IP address to start the node on")
	port := flag.Uint("port", defaultUninitInt, "the port to start the node on")
	newNet := flag.Bool("newnet", false, "this flag acts as a trigger for joining a new network, along with \"connip\" and \"connport\"")
	connectionIp := flag.String("connip", defaultUninitString, "the IP address to connect to when joining a network for the first time")
	connectionPort := flag.Uint("connport", defaultUninitInt, "the port to connect to when joining a network for the first time")
	connsCap := flag.Uint("conncap", defaultUninitInt, "the maximum capacity for primary connections")
	queueCap := flag.Uint("queuecap", defaultUninitInt, "the maximum capacity of the message queue")
	flag.BoolVar(&logging.DebugFlag, "debug", false, "turn on debug logging")

	flag.Parse()

	if *ip == defaultUninitString {
		logging.LogErrorWithExit("IP is an empty string")
	}
	if *port > portMax {
		logging.LogErrorWithExit("port is bigger than the max value allowed %d", portMax)
	}
	if *connsCap == defaultUninitInt {
		logging.LogErrorWithExit("conns capacity is 0 - must be greater than 0")
	}
	if *queueCap == defaultUninitInt {
		logging.LogErrorWithExit("queue capacity is 0 - must be greater than 0")
	}

	currNode, err := node.Create(*ip, uint16(*port), uint16(*connsCap), uint16(*queueCap))
	if err != nil {
		logging.LogErrorWithExit("%s", err)
	}

	if *newNet {
		JoinNewNetwork(currNode, connectionIp, connectionPort, port, ip, connsCap)
	}

	currNode.MainLoop()
}
