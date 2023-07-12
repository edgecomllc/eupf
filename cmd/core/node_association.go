package core

import (
	"context"
	"github.com/edgecomllc/eupf/cmd/config"
	"time"
)

type NodeAssociation struct {
	ID               string
	Addr             string
	NextSessionID    uint64
	Sessions         map[uint64]*Session
	HeartbeatRetries uint32
	cancelRetries    context.CancelFunc
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

func (association *NodeAssociation) RefreshRetries() {
	association.HeartbeatRetries = 0
	if association.cancelRetries != nil {
		association.cancelRetries()
	}
	association.cancelRetries = nil
}

func (association *NodeAssociation) IsExpired() bool {
	return association.HeartbeatRetries > config.Conf.HeartbeatRetries
}

func (association *NodeAssociation) IsHeartbeatScheduled() bool {
	return association.cancelRetries != nil
}

// ScheduleHeartbeatRequest schedules a series of heartbeat requests to be sent to the remote node. Return a cancellation function to stop the scheduled requests.
func (association *NodeAssociation) ScheduleHeartbeatRequest(duration time.Duration, conn *PfcpConnection) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context, duration time.Duration) {
		i := uint32(0)
		ticker := time.NewTicker(duration)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if i >= config.Conf.HeartbeatRetries {
				conn.DeleteAssociation(association.Addr)
				ticker.Stop()
				return
			}
			SendHeartbeatRequest(conn, association.Addr)
		}
	}(ctx, duration)
	return cancel
}
