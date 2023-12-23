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
	HeartbeatChannel chan uint32
	FailedHeartbeats uint32
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

func (association *NodeAssociation) ScheduleHeartbeat(conn *PfcpConnection) {
	association.HeartbeatsActive = true
	ctx := context.Background()

	for {
		sequence := association.NewSequenceID()
		SendHeartbeatRequest(conn, sequence, association.Addr)

		select {
		case <-time.After(time.Duration(config.Conf.HeartbeatTimeout) * time.Second):
			if !association.HandleHeartbeatTimeout() {
				log.Warn().Msgf("the number of unanswered heartbeats has reached the limit, association deleted: %s", association.Addr)
				close(association.HeartbeatChannel)
				conn.DeleteAssociation(association.Addr)
				return
			}
		case seq := <-association.HeartbeatChannel:
			if sequence == seq {
				association.ResetFailedHeartbeats()
				<-time.After(time.Duration(config.Conf.HeartbeatInterval) * time.Second)
			}
		case <-ctx.Done():
			log.Info().Msgf("HeartbeatScheduler context done | association address: %s", association.Addr)
			return
		}
	}
}

func (association *NodeAssociation) ResetFailedHeartbeats() {
	association.Lock()
	association.FailedHeartbeats = 0
	association.Unlock()
}

func (association *NodeAssociation) HandleHeartbeatTimeout() bool {
	association.Lock()
	defer association.Unlock()

	association.FailedHeartbeats++
	return association.FailedHeartbeats < config.Conf.HeartbeatRetries
}

func (association *NodeAssociation) HandleHeartbeat(sequence uint32) {
	association.Lock()
	defer association.Unlock()

	if association.HeartbeatChannel != nil {
		association.HeartbeatChannel <- sequence
	}
}
