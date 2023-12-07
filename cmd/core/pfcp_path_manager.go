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
	return &PfcpPathManager{localAddress: localAddress, peers: map[string]uint32{},
		checkInterval: interval, ctx: ctx, cancelCtx: cancelCtx, cancelAssociationSetup: map[string]context.CancelFunc{}}
}

func (pfcpPathManager *PfcpPathManager) AddPfcpPath(pfcpPeerAddress string) {
	pfcpPathManager.peers[pfcpPeerAddress] = 0
}

func (pfcpPathManager *PfcpPathManager) Run(conn *PfcpConnection) {
	for peer, _ := range pfcpPathManager.peers {
		pfcpPathManager.cancelAssociationSetup[peer] =
			pfcpPathManager.ScheduleAssociationSetupRequest(time.Duration(config.Conf.AssociationSetupTimeout)*time.Second, conn, peer)
	}
	go func() {
		ticker := time.NewTicker(pfcpPathManager.checkInterval)
		defer ticker.Stop()
		for {
			select {
			case <-pfcpPathManager.ctx.Done():
				return
			case <-ticker.C:
				for peer, _ := range pfcpPathManager.peers {
					if IsAssociationSetupEnded(peer, conn) && pfcpPathManager.cancelAssociationSetup[peer] != nil {
						log.Debug().Msgf("Stop sending Association Setup Request to %s", peer)
						pfcpPathManager.cancelAssociationSetup[peer]()
						pfcpPathManager.cancelAssociationSetup[peer] = nil
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
	udpAddr, err := net.ResolveUDPAddr("udp", addr+":8805")
	if err != nil {
		log.Error().Msgf("Failed to resolve udp address from PFCP peer address %s. Error: %s\n", addr, err.Error())
		return true
	}
	_, ok := conn.NodeAssociations[udpAddr.IP.String()]
	return ok
}

func (pfcpPathManager *PfcpPathManager) ScheduleAssociationSetupRequest(duration time.Duration, conn *PfcpConnection, associationAddr string) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context, duration time.Duration) {
		ticker := time.NewTicker(duration)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pfcpPathManager.peers[associationAddr] += 1
				log.Debug().Msgf("Sequence ID: %d", pfcpPathManager.peers[associationAddr])
				SendAssociationSetupRequest(conn, pfcpPathManager.peers[associationAddr], associationAddr)
			}
		}
	}(ctx, duration)
	return cancel
}

func SendAssociationSetupRequest(conn *PfcpConnection, sequenceID uint32, associationAddr string) {
	asreq := message.NewAssociationSetupRequest(sequenceID,
		newIeNodeID(conn.nodeId),
		ie.NewRecoveryTimeStamp(conn.RecoveryTimestamp),
		ie.NewUPFunctionFeatures(), // information of all supported optional features in the UP function; We don't support any optional features at the moment
		// 0x41 = Spare (0) | Assoc Src Inst (0) | Assoc Net Inst (0) | Teid Range (000) | IPV6 (0) | IPV4 (1)
		//      = 00000001
		// If both the ASSONI and ASSOSI flags are set to "0", this shall indicate that the User Plane IP Resource Information
		// provided can be used by CP function for any Network Instance and any Source Interface of GTP-U user plane in the UP
		// function.
		ie.NewUserPlaneIPResourceInformation(0x1, 0, config.Conf.PfcpNodeId, "", "", 0),
	)
	log.Info().Msgf("Sent Association Setup Request to: %s", associationAddr)
	udpAddr, err := net.ResolveUDPAddr("udp", associationAddr+":8805")
	if err != nil {
		log.Error().Msgf("Failed to resolve udp address from PFCP peer address %s. Error: %s\n", associationAddr, err.Error())
		return
	}
	if err := conn.SendMessage(asreq, udpAddr); err != nil {
		log.Info().Msgf("Failed to send Association Setup Request: %s\n", err.Error())
	}
}
