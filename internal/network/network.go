package network

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/TheJ0lly/Overlay-Network/internal/logging"
)

type IpPortPair struct {
	Ip   net.IP
	Port uint16
}

func (ipp IpPortPair) Hash() string {
	buffer := make([]byte, len(ipp.Ip)+2)
	copy(buffer, ipp.Ip)
	buffer = append(buffer, byte(ipp.Port>>8), byte(ipp.Port&0xff))
	return fmt.Sprintf("%X", sha256.Sum256(buffer))
}

var NullIpPortPair = IpPortPair{
	Ip:   net.ParseIP("0.0.0.0"),
	Port: 0,
}

// NetString returns a string in the Host:Port format, be it IPv4 or IPv6.
func (ipp *IpPortPair) NetString() string {
	return net.JoinHostPort(ipp.Ip.String(), fmt.Sprint(ipp.Port))
}

// CompareIpPortPair checks if two IpPortPairs match both in IP and port values.
func CompareIpPortPair(p1, p2 IpPortPair) bool {
	return slices.Compare(p1.Ip, p2.Ip) == 0 && p1.Port == p2.Port
}

func SendToDest(msg json.RawMessage, dest IpPortPair, timeoutInSecs time.Duration) error {

	destNodeHostString := dest.NetString()

	conn, err := net.DialTimeout("tcp", destNodeHostString, time.Second*time.Duration(timeoutInSecs))
	if err != nil {
		return fmt.Errorf("cannot send message to node %s - %s", destNodeHostString, err)
	}
	defer conn.Close()

	size, err := conn.Write(msg)
	if err != nil {
		return fmt.Errorf("cannot send message to node %s - %s", destNodeHostString, err)
	}

	if size != len(msg) {
		return fmt.Errorf("sent %d bytes - expected %d", size, len(msg))
	}

	return nil
}

func SendToMultipleDest(msg json.RawMessage, dests []IpPortPair, skipDests []IpPortPair, timeoutInSecs time.Duration) (sendErrors uint64) {
	sendErrors = 0
	for i := range dests {
		d := dests[i]

		if slices.ContainsFunc(skipDests, func(skipIpp IpPortPair) bool {
			return CompareIpPortPair(d, skipIpp)
		}) {
			logging.LogDebug("jumping over node: %s", d.NetString())
			continue
		}

		if err := SendToDest(msg, d, timeoutInSecs); err != nil {
			logging.LogError("could not forward message - %s", err)
			sendErrors++
			continue
		}
		logging.LogDebug("forwarded message to node: %s", d.NetString())
	}
	return sendErrors
}
