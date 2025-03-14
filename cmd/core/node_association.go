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
	// AssociationStart time.Time // Held until propper failure detection is implemented
	HeartbeatTimeout *time.Timer
}

func NewNodeAssociation(remoteNodeID string, addr string) *NodeAssociation {
	UpfPfcpAssociations.WithLabelValues(remoteNodeID).Set(1)
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
	association.Lock()
	defer association.Unlock()

	association.NextSequenceID += 1
	return association.NextSequenceID
}

func (association *NodeAssociation) ScheduleHeartbeat(conn *PfcpConnection) {
	ctx := context.Background()
	failedHeartbeats := uint32(0)

	for {
		sequence := association.NewSequenceID()
		SendHeartbeatRequest(conn, sequence, association.Addr)

		association.HeartbeatTimeout = time.NewTimer(time.Duration(config.Conf.HeartbeatTimeout) * time.Second)
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
				<-time.After(time.Duration(config.Conf.HeartbeatInterval) * time.Second)
			}
		case <-ctx.Done():
			log.Info().Msgf("HeartbeatScheduler context done | association address: %s", association.Addr)
			return
		}
	}
}

func (association *NodeAssociation) HandleHeartbeat(sequence uint32) {
	association.HeartbeatChannel <- sequence
}
