package core

import (
	"context"
	"github.com/edgecomllc/eupf/cmd/config"
	"time"
)

type NodeAssociation struct {
	ID            string
	Addr          string
	NextSessionID uint64
	Sessions      map[uint64]*Session
	HbRetries     uint32
	cancelRetries context.CancelFunc
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
	if association.cancelRetries != nil {
		association.cancelRetries()
	}
	association.cancelRetries = nil
}

func (association *NodeAssociation) IsExpired() bool {
	return association.HbRetries > config.Conf.HeartBeatRetries
}

func SendTimeoutHeartbeatRequests(duration time.Duration, conn *PfcpConnection, association string) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context, duration time.Duration) {
		i := uint32(0)
		ticker := time.NewTicker(duration)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if i >= config.Conf.HeartBeatRetries {
				conn.DeleteAssociation(association)
				ticker.Stop()
				return
			}
			SendHearbeatReqeust(conn, association)
		}
	}(ctx, duration)
	return cancel
}
