package main

import (
	"flag"
	"fmt"
	"net"
)

func main() {
	port := flag.Uint("port", 0, "The port the nodemgmt will listen to")
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		fmt.Printf("could not start nodemgmt: %s\n", err)
		return
	}

	for {
		_, err := listener.Accept()
		if err != nil {
			fmt.Printf("could not accept ovneto connection: %s\n", err)
			continue
		}

		// handle messages
	}
}
