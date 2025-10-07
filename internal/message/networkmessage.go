package message

import (
	"encoding/json"
	"net"
)

// NetNewNodeJoinMessage will be received by the node when a node joins the network.
// It is used to update the internal state of the node, and to forward to other nodes.
type NetNewNodeJoinMessage struct {
	ExistingNodeUsername string          `json:"ExistingNodeUsername"`
	NodeData             json.RawMessage `json:"NodeData"`
}

// NetQueryPublicIpReq will be received by the node when a node wants to get query its new public IP.
type NetQueryPublicIpReq struct{}

// NetQueryPublicIpResp will be sent to the node that made the NetQueryPublicIpReq and send its public IP.
type NetQueryPublicIpResp struct {
	PublicIP net.IP `json:"PublicIP"`
}
