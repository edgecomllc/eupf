package core

import (
	"github.com/edgecomllc/eupf/cmd/config"
	"time"
)

type NodeAssociation struct {
	ID            string
	Addr          string
	NextSessionID uint64
	Sessions      map[uint64]*Session
	HbRetries     uint32
	LastContact   time.Time
}

func NewNodeAssociation(remoteNodeID string, addr string) *NodeAssociation {
	return &NodeAssociation{
		ID:            remoteNodeID,
		Addr:          addr,
		NextSessionID: 1,
		Sessions:      make(map[uint64]*Session),
	}
}

func (association *NodeAssociation) NewLocalSEID() uint64 {
	association.NextSessionID += 1
	return association.NextSessionID
}

func (association *NodeAssociation) CheckInContact() {
	association.HbRetries = 0
	association.LastContact = time.Now()
}

func (association *NodeAssociation) IsExpired() bool {
	if association.HbRetries > config.Conf.HeartBeatRetries {
		return true
	}
	return false
}
