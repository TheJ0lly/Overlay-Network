package node

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/TheJ0lly/Overlay-Network/internal/logging"
	"github.com/TheJ0lly/Overlay-Network/internal/network"
)

type Stats struct {
	JoinQueriesOngoing []network.IpPortPair `json:"JoinQueriesOngoing"`
	MessagesReceived   map[string]uint64    `json:"MessagesReceived"`
	MessagesForwarded  map[string]uint64    `json:"MessagesForwarded"`
	SendErrors         uint64               `json:"SendErrors"`

	JoinCandidateResponses uint64 `json:"JoinCandidateResponses"`
	JoinCandidateRejects   uint64 `json:"JoinCandidateRejects"`

	DeathAnnouncementsSent     uint64 `json:"DeathAnnouncementsSent"`
	DeathAnnouncementsReceived uint64 `json:"DeathAnnouncementsReceived"`

	DeadHopAttempts         uint64  `json:"DeadHopAttempts"`
	DeadHopNodesGathered    uint64  `json:"DeadHopNodesGathered"`
	DeadHopNodesGatheredAvg float64 `json:"DeadHopNodesGatheredAvg"`

	QueueDrops uint64 `json:"QueueDrops"`

	NodesReplaced      uint64 `json:"NodesReplaced"`
	NewNodeRejects     uint64 `json:"NewNodeRejects"`
	DuplicatedMessages uint64 `json:"DuplicatedMessages"`
}

func NewStats() Stats {
	return Stats{
		JoinQueriesOngoing:         []network.IpPortPair{},
		MessagesReceived:           map[string]uint64{},
		MessagesForwarded:          map[string]uint64{},
		SendErrors:                 0,
		JoinCandidateResponses:     0,
		JoinCandidateRejects:       0,
		DeathAnnouncementsSent:     0,
		DeathAnnouncementsReceived: 0,
		DeadHopAttempts:            0,
		QueueDrops:                 0,
		NodesReplaced:              0,
		NewNodeRejects:             0,
		DuplicatedMessages:         0,
	}
}

func (s *Stats) ExportJson(port uint16) {
	if b, err := json.MarshalIndent(s, "", "\t"); err != nil {
		logging.LogError("could not export stats - could not marshal data")
	} else {
		if err = os.WriteFile(fmt.Sprintf("./stats/Stats_Node_%d.json", port), b, 0666); err != nil {
			logging.LogError("could not export stats - could not create file")
		}
	}

}
