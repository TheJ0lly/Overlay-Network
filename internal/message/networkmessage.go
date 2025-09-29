package message

import (
	"encoding/json"
)

// NetNewNodeJoinMessage will be received by the node when a node joins the network.
// It is used to update the internal state of the node, and to forward to other nodes.
type NetNewNodeJoinMessage struct {
	ExistingNodeUsername string          `json:"ExistingNodeUsername"`
	NodeData             json.RawMessage `json:"NodeData"`
}
