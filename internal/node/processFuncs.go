package node

import (
	"github.com/TheJ0lly/Overlay-Network/internal/logging"
	"github.com/TheJ0lly/Overlay-Network/internal/message"
)

func (n *Node) ProcessNetNewNodeJoinMessage(msg *message.NetNewNodeJoinMessage) {
	newNode, err := Create(msg.Ip.String(), msg.Port, msg.ConnsCap)
	if err != nil {
		logging.LogError("failed to create new node object: %s\n", err)
		return
	}

	// We hold this ConnsCap variable for serialization and usefulness.
	// Maybe we can replace it with "cap()", but then we would need a new structure for serialization
	if len(n.Conns) < cap(n.Conns) {
		n.Conns = append(n.Conns, newNode)
		return
	}

	logging.LogInfo("node connections capacity full - cannot add new node\n")
}
