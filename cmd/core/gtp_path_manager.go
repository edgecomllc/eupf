package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog/log"
)

type GtpPathManager struct {
	localAddress  string
	peers         map[string]uint16
	checkInterval time.Duration
	ctx           context.Context
	cancelCtx     context.CancelFunc
}

func NewGtpPathManager(localAddress string, interval time.Duration) *GtpPathManager {
	ctx, cancelCtx := context.WithCancel(context.Background())
	return &GtpPathManager{localAddress: localAddress, peers: map[string]uint16{}, checkInterval: interval, ctx: ctx, cancelCtx: cancelCtx}
}

func (gtpPathManager *GtpPathManager) AddGtpPath(gtpPeerAddress string) {
	gtpPathManager.peers[gtpPeerAddress] = 0
}

func (gtpPathManager *GtpPathManager) Run() {
	go func() {
		ticker := time.NewTicker(gtpPathManager.checkInterval)
		defer ticker.Stop()
		for {
			select {
			case <-gtpPathManager.ctx.Done():
				// The context is over, stop processing results
				return
			case <-ticker.C:
				for peer, sequenceNumber := range gtpPathManager.peers {
					log.Trace().Msgf("Send GTP Echo to %s, seq %d", peer, sequenceNumber)
					if err := gtpPathManager.sendEcho(peer, sequenceNumber); err != nil {
						log.Warn().Msgf("%v", err)
					} else {
						gtpPathManager.peers[peer] = sequenceNumber + 1
					}
				}
			}
		}
	}()
}

func (gtpPathManager *GtpPathManager) Stop() {
	gtpPathManager.cancelCtx()
}

func (gtpPathManager *GtpPathManager) sendEcho(gtpPeerAddress string, seq uint16) error {
	gtpEchoRequest := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(gtpEchoRequest, gopacket.SerializeOptions{},
		&layers.GTPv1U{
			Version:        1,
			MessageType:    1, // GTPU_ECHO_REQUEST
			TEID:           0,
			SequenceNumber: seq,
		},
	); err != nil {
		return fmt.Errorf("serializing input packet failed: %v", err)
	}

	udpLocalAddr, err := net.ResolveUDPAddr("udp", gtpPathManager.localAddress)
	if err != nil {
		return fmt.Errorf("can't resolve local UDP address: %v", err)
	}
	udpRemoteAddr, err := net.ResolveUDPAddr("udp", gtpPeerAddress)
	if err != nil {
		return fmt.Errorf("can't resolve remote UDP address: %v", err)
	}
	conn, err := net.DialUDP("udp", udpLocalAddr, udpRemoteAddr)
	if err != nil {
		return fmt.Errorf("can't create UDP connection: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write(gtpEchoRequest.Bytes()); err != nil {
		return fmt.Errorf("can't send echo request: %v", err)
	}

	buf := make([]byte, 1500)
	conn.SetReadDeadline(time.Now().Add(time.Second * 3))
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("can't read echo response: %v", err)
	}

	response := gopacket.NewPacket(buf[:n], layers.LayerTypeGTPv1U, gopacket.Default)
	if gtpLayer := response.Layer(layers.LayerTypeGTPv1U); gtpLayer != nil {
		gtp, _ := gtpLayer.(*layers.GTPv1U)

		if gtp.MessageType != 2 { //GTPU_ECHO_RESPONSE
			return fmt.Errorf("unexpected gtp echo response: %d", gtp.MessageType)
		}
		if gtp.SequenceNumber != seq {
			return fmt.Errorf("unexpected gtp echo response sequence: %d", gtp.SequenceNumber)
		}
		if gtp.TEID != 0 {
			return fmt.Errorf("unexpected gtp echo response TEID: %d", gtp.TEID)
		}
	} else {
		return errors.New("unexpected gtp echo response")
	}

	return nil
}
