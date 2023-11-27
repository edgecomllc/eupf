package core

import (
	"context"
	"net"
	"time"

	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PfcpPathManager struct {
	localAddress           string
	peers                  map[string]uint32
	checkInterval          time.Duration
	ctx                    context.Context
	cancelCtx              context.CancelFunc
	cancelAssociationSetup map[string]context.CancelFunc
}

func NewPfcpPathManager(localAddress string, interval time.Duration) *PfcpPathManager {
	ctx, cancelCtx := context.WithCancel(context.Background())
	return &PfcpPathManager{localAddress: localAddress, peers: map[string]uint32{}, checkInterval: interval, ctx: ctx, cancelCtx: cancelCtx}
}

func (pfcpPathManager *PfcpPathManager) AddPfcpPath(gtpPeerAddress string) {
	pfcpPathManager.peers[gtpPeerAddress] = 0
}

func (pfcpPathManager *PfcpPathManager) Run(conn *PfcpConnection) {
	for peer, sequenceNumber := range pfcpPathManager.peers {
		pfcpPathManager.cancelAssociationSetup[peer] =
			ScheduleAssociationSetupRequest(time.Duration(config.Conf.HeartbeatTimeout)*time.Second, conn, peer, sequenceNumber)
	}
	go func() {
		ticker := time.NewTicker(pfcpPathManager.checkInterval)
		defer ticker.Stop()
		for {
			select {
			case <-pfcpPathManager.ctx.Done():
				// The context is over, stop processing results
				return
			case <-ticker.C:
				for peer, _ := range pfcpPathManager.peers {
					if !IsAssociationSetupEnded(peer, conn) {
						pfcpPathManager.cancelAssociationSetup[peer]()
					}
				}
			}
		}
	}()
}

func (pfcpPathManager *PfcpPathManager) Stop() {
	pfcpPathManager.cancelCtx()
}

func IsAssociationSetupEnded(addr string, conn *PfcpConnection) bool {
	_, ok := conn.NodeAssociations[addr]
	return ok
}

func ScheduleAssociationSetupRequest(duration time.Duration, conn *PfcpConnection, associationAddr string, seq uint32) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context, duration time.Duration) {
		ticker := time.NewTicker(duration)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			SendAssociationSetupRequest(conn, seq, associationAddr)
		}
	}(ctx, duration)
	return cancel
}

func SendAssociationSetupRequest(conn *PfcpConnection, sequenceID uint32, associationAddr string) {
	asreq := message.NewAssociationSetupRequest(sequenceID,
		newIeNodeID(conn.nodeId),
		ie.NewRecoveryTimeStamp(conn.RecoveryTimestamp),
		ie.NewUPFunctionFeatures(),
		// ie.NewUserPlaneIPResourceInformation(0, 0, "0", "0", "0", 0),
	)
	log.Debug().Msgf("Sent Association Setup Request to: %s", associationAddr)
	udpAddr, err := net.ResolveUDPAddr("udp", associationAddr+":8805")
	if err == nil {
		if err := conn.SendMessage(asreq, udpAddr); err != nil {
			log.Info().Msgf("Failed to send Association Setup Request: %s\n", err.Error())
		}
	} else {
		log.Info().Msgf("Failed to send Association Setup Request: %s\n", err.Error())
	}
}
