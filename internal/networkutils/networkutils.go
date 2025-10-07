package networkutils

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
)

func GetIPFromDNS() net.IP {
	DNSs := []string{
		"https://api.ipify.org",
		"https://checkip.amazonaws.com",
		"https://ifconfig.me/ip",
		"https://ipinfo.io/ip",
	}

	pubIp := ""

	for _, dns := range DNSs {
		resp, err := http.Get(dns)
		if err != nil {
			fmt.Printf("error while getting public IP through DNS: %s", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("error while getting public IP through DNS: %s", err)
			continue
		}

		pubIp = string(body)
		break
	}

	return net.ParseIP(pubIp)
}

func SendMessage(dest string, msg message.MessageEnvelope, timeDuration time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", dest, timeDuration)
	if err != nil {
		return nil, err
	}

	buff, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	if err := SendMessageSize(conn, len(buff)); err != nil {
		return nil, err
	}

	n, err := conn.Write(buff)
	if err != nil {
		return nil, err
	}

	if n != len(buff) {
		return nil, fmt.Errorf("sent %d bytes, but message has %d bytes", n, len(buff))
	}

	return conn, nil
}

func SendMessageSize(conn net.Conn, msgSize int) error {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(msgSize))

	n, err := conn.Write(buff)
	if err != nil {
		return err
	}

	fmt.Printf("sent to %s node %d bytes\n", conn, n)
	return nil
}

func ReceiveMessage(conn net.Conn) (message.MessageEnvelope, error) {
	msgSize, err := ReceiveMessageSize(conn)
	if err != nil {
		return message.MessageEnvelope{}, err
	}

	buff := make([]byte, msgSize)
	n, err := conn.Read(buff)
	if err != nil {
		return message.MessageEnvelope{}, err
	}

	if n != msgSize {
		return message.MessageEnvelope{}, fmt.Errorf("received %d bytes, but expected %d bytes", n, msgSize)
	}

	var msg message.MessageEnvelope
	if err := json.Unmarshal(buff, &msg); err != nil {
		return message.MessageEnvelope{}, fmt.Errorf("error while unmarshaling: %s", err)
	}

	return msg, nil
}

func ReceiveMessageSize(conn net.Conn) (int, error) {
	buff := make([]byte, 4)

	// TODO: I don't know if we should check for the bytes received? Because we send a 4 bytes across the network...
	_, err := conn.Read(buff)
	if err != nil {
		return 0, err
	}

	return int(binary.BigEndian.Uint32(buff)), nil
}
