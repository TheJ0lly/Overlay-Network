package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"slices"
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
			NewNode:   currNode.GetIpPortPair(),
			Timestamp: initialTimestamp,
		},
		currNode.GetIpPortPair(),
	); err != nil {
		logging.LogErrorWithExit("could not create join query message - %s", err)
	}

	if b, err = message.SerializeMessageEnvelope(&env); err != nil {
		logging.LogErrorWithExit("could not marshal initial message envelope for joining a new network - %s", err)
	}

	connectionPair := net.JoinHostPort(*connectionIp, fmt.Sprintf("%d", *connectionPort))
	if err = currNode.SendMessageToIp(b, parsedIp, uint16(*connectionPort)); err != nil {
		logging.LogErrorWithExit("could not sent message envelope to node %s", connectionPair)
	}
	logging.LogInfo("sent message to %s: type=%s data=%s", connectionPair, env.Type, env.Data)

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

	type ResponiveNode struct {
		pair        message.IpPortPair
		rttDuration int64
	}

	var responsiveNodes []ResponiveNode

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
				conn.Close()
				continue
			}

			if err = message.DeserializeMessageEnvelope(&env, b); err != nil {
				logging.LogError("error while deserializing message envelope - %s", err)
				conn.Close()
				continue
			}

			msg := message.NetNewNodeJoinQueryMessage{}
			if err := json.Unmarshal(env.Data, &msg); err != nil {
				logging.LogError("error while deserializing message - %s", err)
				conn.Close()
				continue
			}

			rttDuration := msg.Timestamp - initialTimestamp
			if rttDuration <= 0 {
				// a predefined 10 milliseconds mock RTT
				rttDuration = 10
			}

			responsiveNodes = append(responsiveNodes, ResponiveNode{
				pair: message.IpPortPair{
					Ip:   msg.NewNode.Ip,
					Port: msg.NewNode.Port,
				},
				rttDuration: rttDuration,
			})
		}
	}

	if len(responsiveNodes) == 0 {
		logging.LogErrorWithExit("could not find a suitable node to attach to")
	}

	confirmEnv, err := message.CreateMessageEnvelope(
		message.NetNewNodeJoinConfirm,
		&message.NetNewNodeJoinConfirmMessage{
			// As of now does not matter, but maybe we add some RTT exclusion over X
			IsSuitable: true,
		},
		currNode.GetIpPortPair())

	if err != nil {
		logging.LogErrorWithExit("could not create message envelope for join confirm: %s", err)
	}

	confirmEnvBytes, err := message.SerializeMessageEnvelope(&confirmEnv)
	if err != nil {
		logging.LogErrorWithExit("could not serialize net join message - %s", err)
	}

	slices.SortStableFunc(responsiveNodes, func(a, b ResponiveNode) int {
		return int(a.rttDuration - b.rttDuration)
	})

	var bestNode message.IpPortPair
	var gotConnChan chan struct{} = make(chan struct{}, 1)

	running = true
	for i := range responsiveNodes {
		// At this point we assume that the slice is sorted based on timestamp.
		// Thus if we iterate over the slice, we should get the best candidates.
		reNo := responsiveNodes[i]

		if err = currNode.SendMessageToIp(confirmEnvBytes, reNo.pair.Ip, reNo.pair.Port); err != nil {
			logging.LogErrorWithExit("could not send net join message - %s", err)
		}
		logging.LogInfo("sent message to responsive node %s: type=%s data=%s", reNo.pair, confirmEnv.Type, confirmEnv.Data)

		if list, err = net.Listen("tcp", currNode.GetNodeAddress()); err != nil {
			logging.LogErrorWithExit("could not start listener for the initial message - %s", err)
		}

		// 3 * RTT is the window for accepting a new connection from a node.
		go func() {
			tick := time.NewTicker((time.Millisecond * time.Duration(responsiveNodes[i].rttDuration)) * 3)
			select {
			case <-gotConnChan:
				tick.Stop()
				logging.LogDebug("got a responsive node connected - timeout cancelled")
				return
			case <-tick.C:
				logging.LogInfo("timeout for node - %v", responsiveNodes[i].pair)
				list.Close()
			}
		}()

		conn, err = list.Accept()
		if err != nil {
			logging.LogError("%s", err)
			list.Close()
			continue
		}
		gotConnChan <- struct{}{}

		b, err := io.ReadAll(conn)
		if err != nil {
			logging.LogError("could not read all bytes: %s", err)
			conn.Close()
			list.Close()
			continue
		}

		confirmEnvRespEnv := message.MessageEnvelope{}

		if err := json.Unmarshal(b, &confirmEnvRespEnv); err != nil {
			logging.LogError("could not unmarshal message envelope: %s", err)
			conn.Close()
			list.Close()
			continue
		}

		msg := message.NetNewNodeJoinConfirmMessage{}
		if err := json.Unmarshal(confirmEnvRespEnv.Data, &msg); err != nil {
			logging.LogError("could not unmarshal message: %s", err)
			conn.Close()
			list.Close()
			continue
		}

		if !msg.IsSuitable {
			logging.LogInfo("candidate node %v refused attachment - moving on", responsiveNodes[i].pair)
			conn.Close()
			list.Close()
			continue
		}

		// Maybe we replace this with a message, but maybe not
		if cap(currNode.Conns) > len(currNode.Conns) {
			if newNode, err := node.Create(reNo.pair.Ip.String(), reNo.pair.Port, 0, 0); err != nil {
				logging.LogError("could not add the new node: %s - moving on", err)
				conn.Close()
				list.Close()
				continue
			} else {
				currNode.Conns = append(currNode.Conns, newNode)
				newNode.LastTimeAlive = time.Now().UnixMilli()
				bestNode.Ip = reNo.pair.Ip
				bestNode.Port = reNo.pair.Port
				logging.LogDebug("added new node - %s", newNode)
				logging.LogDebug("attached node state - %s", currNode)
				list.Close()
				conn.Close()
				// Here we break out of the loop since we found a good node and it confirmed the attachment.
				break
			}
		}
	}
	if env, err = message.CreateMessageEnvelope(
		message.NetNewNodeJoin,
		&message.NetNewNodeJoinMessage{
			AttachedNode: bestNode,
			JoiningNode: message.IpPortPair{
				Ip:   net.ParseIP(*ip),
				Port: uint16(*port),
			},
			ReplacedNode: message.NullIpPortPair,
		},
		message.IpPortPair{
			Ip:   net.ParseIP(*ip),
			Port: uint16(*port),
		},
	); err != nil {
		logging.LogErrorWithExit("could not create the join message envelope: %s", err)
	} else {
		if b, err = message.SerializeMessageEnvelope(&env); err != nil {
			logging.LogErrorWithExit("could not serialize the join message envelope: %s", err)
		}

		if err = currNode.SendMessageToIp(b, bestNode.Ip, bestNode.Port); err != nil {
			logging.LogErrorWithExit("could not send join message: %s", err)
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
	lifelineTimer := flag.Uint("lifeline", defaultUninitInt, "the duration in seconds between lifeline messages")
	deathannounceTimer := flag.Uint("death", defaultUninitInt, "the duration in seconds between last lifeline message until we announce its death")
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
	if *lifelineTimer == defaultUninitInt {
		logging.LogErrorWithExit("lifeline duration is 0 - must be greater than 0")
	}
	if *deathannounceTimer == defaultUninitInt {
		logging.LogErrorWithExit("death duration is 0 - must be greater than 0")
	}

	currNode, err := node.Create(*ip, uint16(*port), uint16(*connsCap), uint16(*queueCap))
	if err != nil {
		logging.LogErrorWithExit("%s", err)
	}
	currNode.LifeLineTimer = uint8(*lifelineTimer)
	logging.LogDebug("setting lifeline timer duration to: %d", currNode.LifeLineTimer)

	currNode.DeathTimer = uint8(*deathannounceTimer)
	logging.LogDebug("setting death timer duration to: %d", currNode.DeathTimer)

	if *newNet {
		JoinNewNetwork(currNode, connectionIp, connectionPort, port, ip, connsCap)
	}

	currNode.MainLoop()
}
