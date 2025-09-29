package networkutils

import (
	"io"
	"log"
	"net"
	"net/http"

	"github.com/TheJ0lly/Overlay-Network/internal/message"
	"github.com/TheJ0lly/Overlay-Network/internal/node"
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
			log.Printf("error while getting public IP through DNS: %s", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error while getting public IP through DNS: %s", err)
			continue
		}

		pubIp = string(body)
		break
	}

	return net.ParseIP(pubIp)
}

func SendMessage(dest *node.Node, msg message.MessageEnvelope) net.Conn {
	return nil
}
