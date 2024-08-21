package core

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PfcpPathManager struct {
	conn                    *PfcpConnection
	localAddress            string
	servers                 map[string]uint32
	checkInterval           time.Duration
	ctx                     context.Context
	cancelCtx               context.CancelFunc
	ongoingAssociationSetup map[string]context.CancelFunc
}

func NewPfcpPathManager(conn *PfcpConnection, localAddress string, interval time.Duration) *PfcpPathManager {
	ctx, cancelCtx := context.WithCancel(context.Background())
	return &PfcpPathManager{
		conn:                    conn,
		localAddress:            localAddress,
		servers:                 map[string]uint32{},
		checkInterval:           interval,
		ctx:                     ctx,
		cancelCtx:               cancelCtx,
		ongoingAssociationSetup: map[string]context.CancelFunc{}}
}

func (pfcpPathManager *PfcpPathManager) AddPfcpServer(pfcpServerAddress string) {
	pfcpPathManager.servers[pfcpServerAddress] = 0
}

func (pfcpPathManager *PfcpPathManager) Run() {
	pfcpPathManager.initiateAssociationSetup()

	go func() {
		ticker := time.NewTicker(pfcpPathManager.checkInterval)
		defer ticker.Stop()
		for {
			select {
			case <-pfcpPathManager.ctx.Done():
				return
			case <-ticker.C:
				pfcpPathManager.cancelOngoningAssociationSetup()
			}
		}
	}()
}

func (pfcpPathManager *PfcpPathManager) Stop() {
	pfcpPathManager.cancelCtx()
}

func (pfcpPathManager *PfcpPathManager) isAssociationEstablished(addr string) bool {
	udpAddr, err := net.ResolveUDPAddr("udp", addr+":8805")
	if err != nil {
		log.Error().Msgf("Failed to resolve udp address from PFCP peer address %s. Error: %s\n", addr, err.Error())
		return true
	}
	_, ok := pfcpPathManager.conn.NodeAssociations[udpAddr.IP.String()]
	return ok
}

func (pfcpPathManager *PfcpPathManager) initiateAssociationSetup() {
	for server, _ := range pfcpPathManager.servers {
		pfcpPathManager.ongoingAssociationSetup[server] =
			pfcpPathManager.scheduleAssociationSetupRequest(
				time.Duration(config.Conf.AssociationSetupTimeout)*time.Second, server)
	}
}

func (pfcpPathManager *PfcpPathManager) hasOngoningAssociationSetup(server string) bool {
	return pfcpPathManager.ongoingAssociationSetup[server] != nil
}

func (pfcpPathManager *PfcpPathManager) cancelOngoningAssociationSetup() {
	for server, _ := range pfcpPathManager.servers {
		if pfcpPathManager.isAssociationEstablished(server) && pfcpPathManager.hasOngoningAssociationSetup(server) {
			log.Debug().Msgf("Stop sending Association Setup Request to %s", server)
			pfcpPathManager.ongoingAssociationSetup[server]()
			pfcpPathManager.ongoingAssociationSetup[server] = nil
		}
	}
}

func (pfcpPathManager *PfcpPathManager) scheduleAssociationSetupRequest(duration time.Duration, associationAddr string) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context, duration time.Duration) {
		ticker := time.NewTicker(duration)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pfcpPathManager.servers[associationAddr] += 1
				pfcpPathManager.sendAssociationSetupRequest(pfcpPathManager.servers[associationAddr], associationAddr)
			}
		}
	}(ctx, duration)
	return cancel
}

func EncodeFQDN(fqdn string) []byte {
	b := make([]byte, len(fqdn) /*+1*/)

	var offset = 0
	for _, label := range strings.Split(fqdn, ".") {
		l := len(label)
		//b[offset] = uint8(l)
		copy(b[offset: /*+1*/], label)
		offset += l /*+ 1*/
	}

	return b
}

func NewNodeIDHuawei(ipv4, ipv6, fqdn string) *ie.IE {
	var p []byte

	switch {
	case ipv4 != "":
		p = make([]byte, 5)
		p[0] = ie.NodeIDIPv4Address
		copy(p[1:], net.ParseIP(ipv4).To4())
	case ipv6 != "":
		p = make([]byte, 17)
		p[0] = ie.NodeIDIPv6Address
		copy(p[1:], net.ParseIP(ipv6).To16())
	case fqdn != "":
		p = make([]byte, 2+len([]byte(fqdn)))
		p[0] = ie.NodeIDFQDN
		copy(p[1:], EncodeFQDN(fqdn))
	default: // all params are empty
		return nil
	}

	return ie.New(ie.NodeID, p)
}

func newIeNodeIDHuawei(nodeID string) *ie.IE {
	ip := net.ParseIP(nodeID)
	if ip != nil {
		if ip.To4() != nil {
			return NewNodeIDHuawei(nodeID, "", "")
		}
		return NewNodeIDHuawei("", nodeID, "")
	}
	return NewNodeIDHuawei("", "", nodeID)
}

func (pfcpPathManager *PfcpPathManager) sendAssociationSetupRequest(sequenceID uint32, associationAddr string) {
	conn := pfcpPathManager.conn

	AssociationSetupRequest := message.NewAssociationSetupRequest(sequenceID,
		newIeNodeIDHuawei(conn.nodeId),
		ie.NewRecoveryTimeStamp(conn.RecoveryTimestamp),
		ie.NewUPFunctionFeatures(conn.featuresOctets[:]...),
		ie.NewVendorSpecificIE(32787, 2011, []byte{0x02, 0x00, 0x08, 0x30, 0x30, 0x30, 0x34, 0x64, 0x67, 0x77, 0x34, conn.n3Address.To4()[0], conn.n3Address.To4()[1], conn.n3Address.To4()[2], conn.n3Address.To4()[3]}),
		//CHOICE
		//	user-plane-element-weight
		//		enterprise-id: ---- 0x7db(2011)
		//		weight-value: ---- 0x1(1)
		ie.NewVendorSpecificIE(32803, 2011, []byte{1}),
		//CHOICE
		//	lock-information
		//		enterprise-id: ---- 0x7db(2011)
		//		lock-information-value: ---- 0x0(0)
		ie.NewVendorSpecificIE(32806, 2011, []byte{0}),
		//CHOICE
		//	apn-support-mode
		//		enterprise-id: ---- 0x7db(2011)
		//		apn-support-mode-value: ---- 0x0(0)
		ie.NewVendorSpecificIE(32857, 2011, []byte{0}),
		//CHOICE
		//	sx-uf-flag
		//		enterprise-id: ---- 0x7db(2011)
		//		spare: ---- 0x0(0)
		//		nb-iot-value: ---- 0x1(1)
		//		dual-connectivity-with-nr-value: ---- 0x1(1)
		ie.NewVendorSpecificIE(32900, 2011, []byte{3}),
		//CHOICE
		//	high-bandwidth
		//		enterprise-id: ---- 0x7db(2011)
		//		high-bandwidth-value: ---- 0x1(1)
		ie.NewVendorSpecificIE(32901, 2011, []byte{1}),
	)
	log.Info().Msgf("Sent Association Setup Request to: %s", associationAddr)

	udpAddr, err := net.ResolveUDPAddr("udp", associationAddr+":8805")
	if err != nil {
		log.Error().Msgf("Failed to resolve udp address from PFCP peer address %s. Error: %s\n", associationAddr, err.Error())
		return
	}
	if err := conn.SendMessage(AssociationSetupRequest, udpAddr); err != nil {
		log.Info().Msgf("Failed to send Association Setup Request: %s\n", err.Error())
	}
}
