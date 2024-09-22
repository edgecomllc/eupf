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
					log.Trace().Msgf("Send GTP Echo request to %s, seq %d", peer, sequenceNumber)
					elapseTime, err := gtpPathManager.sendEcho(peer, sequenceNumber)
					if err != nil {
						log.Warn().Msgf("%v", err)
						continue
					}

					log.Trace().Msgf("Received GTP Echo response from %s, seq %d in %d ms", peer, sequenceNumber, elapseTime.Milliseconds())
					gtpPathManager.peers[peer] = sequenceNumber + 1

				}
			}
		}
	}()
}

func (gtpPathManager *GtpPathManager) Stop() {
	gtpPathManager.cancelCtx()
}

func (gtpPathManager *GtpPathManager) sendEcho(gtpPeerAddress string, seq uint16) (time.Duration, error) {
	gtpEchoRequest := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(gtpEchoRequest, gopacket.SerializeOptions{},
		&layers.GTPv1U{
			Version:     1,
			MessageType: 1, // GTPU_ECHO_REQUEST
			TEID:        0,
			//SequenceNumberFlag: true,
			//SequenceNumber:     seq,
		},
	); err != nil {
		return 0, fmt.Errorf("serializing input packet failed: %v", err)
	}

	udpLocalAddr, err := net.ResolveUDPAddr("udp", gtpPathManager.localAddress)
	if err != nil {
		return 0, fmt.Errorf("can't resolve local UDP address: %v", err)
	}
	udpRemoteAddr, err := net.ResolveUDPAddr("udp", gtpPeerAddress)
	if err != nil {
		return 0, fmt.Errorf("can't resolve remote UDP address: %v", err)
	}
	conn, err := net.DialUDP("udp", udpLocalAddr, udpRemoteAddr)
	if err != nil {
		return 0, fmt.Errorf("can't create UDP connection: %v", err)
	}
	defer conn.Close()

	receiveBuffer := make([]byte, 1500)
	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 3))

	sendTime := time.Now()
	if _, err := conn.Write(gtpEchoRequest.Bytes()); err != nil {
		return 0, fmt.Errorf("can't send echo request: %v", err)
	}

	n, err := conn.Read(receiveBuffer)
	if err != nil {
		return 0, fmt.Errorf("can't read echo response: %v", err)
	}
	elapsedTime := time.Since(sendTime)

	response := gopacket.NewPacket(receiveBuffer[:n], layers.LayerTypeGTPv1U, gopacket.Default)
	if gtpLayer := response.Layer(layers.LayerTypeGTPv1U); gtpLayer != nil {
		gtp, _ := gtpLayer.(*layers.GTPv1U)

		if gtp.MessageType != 2 { //GTPU_ECHO_RESPONSE
			return 0, fmt.Errorf("unexpected gtp echo response: %d", gtp.MessageType)
		}
		//if gtp.SequenceNumberFlag && gtp.SequenceNumber != seq {
		//	return 0, fmt.Errorf("unexpected gtp echo response sequence: %d", gtp.SequenceNumber)
		//}
		if gtp.TEID != 0 {
			return 0, fmt.Errorf("unexpected gtp echo response TEID: %d", gtp.TEID)
		}
	} else {
		return 0, errors.New("unexpected gtp echo response")
	}

	return elapsedTime, nil
}
