package core

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type GtpPathManager struct {
	localAddress string
	peers        map[uint32]uint16
}

func New(localAddress net.IP) *GtpPathManager {
	return &GtpPathManager{}
}

func (gtpPathManager *GtpPathManager) Run() {
	go func() {
		for {
			time.Sleep(time.Duration(config.Conf.HeartbeatInterval) * time.Second)
		}
	}()
}

func (gtpPathManager *GtpPathManager) sendEcho(seq uint16) error {
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

	// udpAddr, err := net.ResolveUDPAddr("udp", gtpPathManager.localAddress)
	// if err != nil {
	// 	log.Panic().Msgf("Can't resolve local UDP address: %s", err.Error())
	// 	return err
	// }
	conn, err := net.Dial("udp", "127.0.0.1:2152")
	if err != nil {
		return fmt.Errorf("... failed: %v", err)
	}

	conn.Write(gtpEchoRequest.Bytes())
	buf := make([]byte, 1500)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("... failed: %v", err)
	}

	response := gopacket.NewPacket(buf[:n], layers.LayerTypeGTPv1U, gopacket.Default)
	if gtpLayer := response.Layer(layers.LayerTypeGTPv1U); gtpLayer != nil {
		gtp, _ := gtpLayer.(*layers.GTPv1U)

		if gtp.MessageType != 2 { //GTPU_ECHO_RESPONSE
			return fmt.Errorf("unexpected gtp response: %d", gtp.MessageType)
		}
		if gtp.SequenceNumber != seq {
			return fmt.Errorf("unexpected gtp sequence: %d", gtp.SequenceNumber)
		}
		if gtp.TEID != 0 {
			return fmt.Errorf("unexpected gtp TEID: %d", gtp.TEID)
		}
	} else {
		return errors.New("unexpected response")
	}

	return nil
}
