package core

import "time"

type NodeAssociation struct {
	ID            string
	Addr          string
	NextSessionID uint64
	Sessions      map[uint64]*Session
	HbRetries     int
	LastContact   time.Time
}

func NewNodeAssociation(remoteNodeID string, addr string) *NodeAssociation {
	return &NodeAssociation{
		ID:            remoteNodeID,
		Addr:          addr,
		NextSessionID: 0,
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
