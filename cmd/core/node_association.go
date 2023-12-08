package core

import (
	"context"
	"github.com/rs/zerolog/log"
	"sync"
	"time"

	"github.com/edgecomllc/eupf/cmd/config"
)

type NodeAssociation struct {
	ID               string
	Addr             string
	NextSessionID    uint64
	NextSequenceID   uint32
	Sessions         map[uint64]*Session
	HeartbeatChannel chan uint32
	FailedHeartbeats uint8
	HeartbeatsActive bool
	sync.Mutex
	// AssociationStart time.Time // Held until propper failure detection is implemented
}

func NewNodeAssociation(remoteNodeID string, addr string) *NodeAssociation {
	return &NodeAssociation{
		ID:               remoteNodeID,
		Addr:             addr,
		NextSessionID:    1,
		NextSequenceID:   1,
		Sessions:         make(map[uint64]*Session),
		HeartbeatChannel: make(chan uint32),
		// AssociationStart: time.Now(),
	}
}

func (association *NodeAssociation) NewLocalSEID() uint64 {
	association.NextSessionID += 1
	return association.NextSessionID
}

func (association *NodeAssociation) NewSequenceID() uint32 {
	association.NextSequenceID += 1
	return association.NextSequenceID
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
			seq := association.NewSequenceID()
			SendHeartbeatRequest(conn, seq, association.Addr)
		}
	}(ctx, duration)
	return cancel
}

func (association *NodeAssociation) HeartbeatScheduler(conn *PfcpConnection) {

	ctx, cancel := context.WithCancel(context.Background())
	//timeout := config.Conf.HeartbeatInterval

	var sequence uint32
	sequence = association.NewSequenceID()
	SendHeartbeatRequest(conn, sequence, association.Addr)

	for {
		select {
		case <-time.After(3 * time.Second): //timeout
			association.Lock()
			association.FailedHeartbeats++
			if association.FailedHeartbeats >= 5 {
				log.Info().Msgf("the number of unanswered heartbeats has reached the limit, association deleted: %s", association.Addr)
				close(association.HeartbeatChannel)
				conn.DeleteAssociation(association.Addr)
				association.Unlock()
				return
			} else {
				sequence = association.NewSequenceID()
				SendHeartbeatRequest(conn, sequence, association.Addr)
			}
			association.Unlock()
		case <-ctx.Done():
			log.Info().Msgf("HeartbeatScheduler context done | association address: %s", association.Addr)
			cancel()
			return
		case seq := <-association.HeartbeatChannel:
			if sequence == seq {
				sequence = association.NewSequenceID()
				SendHeartbeatRequest(conn, sequence, association.Addr)
			}
		}
		time.Sleep(2 * time.Second)
	}
}
