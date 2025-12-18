package main

import (
	"fmt"
	"io"
	"net"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
)

// run to send a message from another terminal. Testing phase
//echo "{\"Type\":0,\"Data\":\"eyJOb2RlSXAiOiIxMjcuMC4wLjEifQ==\"}" > /dev/tcp/::/8080

func main() {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	fmt.Printf("listening on: %s\n", l.Addr())

	conn, err := l.Accept()
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	b, err := io.ReadAll(conn)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	fmt.Printf("Bytes: %s\n", b)

	env := message.MessageEnvelope{}

	err = message.DeserializeMessageEnvelope(&env, b)

	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	fmt.Printf("Message type: %d\nMessage data: %s\n", env.Type, env.Data)
}
