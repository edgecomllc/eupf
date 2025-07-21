package core

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/edgecomllc/eupf/cmd/config"
)

type NodeAssociation struct {
	ID               string
	Addr             string
	NextSessionID    uint64
	NextSequenceID   uint32
	Sessions         map[uint64]*Session
	HeartbeatChannel chan uint32 `json:"-"`
	HeartbeatsActive bool
	sync.Mutex       `json:"-"`
	HeartbeatTimeout *time.Timer
	ctx              context.Context
	ctxCancel        context.CancelFunc
	// AssociationStart time.Time // Held until propper failure detection is implemented
}

func NewNodeAssociation(remoteNodeID string, addr string) *NodeAssociation {
	UpfPfcpAssociations.WithLabelValues(remoteNodeID).Set(1)
	ctx, cancel := context.WithCancel(context.Background())

	return &NodeAssociation{
		ID:               remoteNodeID,
		Addr:             addr,
		NextSessionID:    1,
		NextSequenceID:   1,
		Sessions:         make(map[uint64]*Session),
		HeartbeatChannel: make(chan uint32),
		ctx:              ctx,
		ctxCancel:        cancel,
		// AssociationStart: time.Now(),
	}
}

func (association *NodeAssociation) NewLocalSEID() uint64 {
	association.NextSessionID += 1
	return association.NextSessionID
}

func (association *NodeAssociation) NewSequenceID() uint32 {
	association.Lock()
	defer association.Unlock()

	association.NextSequenceID += 1
	return association.NextSequenceID
}

func (association *NodeAssociation) ScheduleHeartbeat(conn *PfcpConnection) {
	var (
		failedHeartbeats uint32
		sequence         uint32
	)

	heartbeatTicker := time.NewTicker(time.Duration(config.Conf.HeartbeatInterval) * time.Second)
	defer heartbeatTicker.Stop()

	association.HeartbeatTimeout = time.NewTimer(0)
	defer association.HeartbeatTimeout.Stop()

	for {
		select {
		case <-heartbeatTicker.C:
			sequence = association.NewSequenceID()
			SendHeartbeatRequest(conn, sequence, association.Addr)

			association.HeartbeatTimeout.Reset(time.Duration(config.Conf.HeartbeatTimeout) * time.Second)

			select {
			case <-association.HeartbeatTimeout.C:
				failedHeartbeats++
				if failedHeartbeats >= config.Conf.HeartbeatRetries {
					log.Warn().Msgf("the number of unanswered heartbeats has reached the limit, association deleted: %s", association.Addr)
					UpfPfcpAssociations.WithLabelValues(association.ID).Set(0)
					conn.heartbeatFailedC <- association.Addr
					return
				}

			case seq := <-association.HeartbeatChannel:
				if sequence == seq {
					association.HeartbeatTimeout.Stop()
					failedHeartbeats = 0
				}
			case <-association.ctx.Done():
				return
			}

		case <-association.ctx.Done():
			log.Info().Msgf("schedule heartbeat context done, association address: %s", association.Addr)
			return
		}
	}
}

func (association *NodeAssociation) HandleHeartbeat(sequence uint32) {
	association.HeartbeatChannel <- sequence
}

func (association *NodeAssociation) Close() {
	association.ctxCancel()
}
