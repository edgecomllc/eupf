package core

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/edgecomllc/eupf/cmd/core/service"

	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/rs/zerolog/log"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PfcpConnection struct {
	udpConn           *net.UDPConn
	pfcpHandlerMap    PfcpHandlerMap
	associationMutex  *sync.Mutex
	NodeAssociations  map[string]*NodeAssociation
	nodeId            string
	nodeAddrV4        net.IP
	n3Address         net.IP
	mapOperations     ebpf.ForwardingPlaneController
	RecoveryTimestamp time.Time
	featuresOctets    []uint8
	ResourceManager   *service.ResourceManager
	heartbeatFailedC  chan string
	nodes             []string
}

func (connection *PfcpConnection) GetAssociation(assocAddr string) *NodeAssociation {
	if assoc, ok := connection.NodeAssociations[assocAddr]; ok {
		return assoc
	}
	return nil
}

func CreatePfcpConnection(addr string, pfcpHandlerMap PfcpHandlerMap, nodeId string, n3Ip string, mapOperations ebpf.ForwardingPlaneController, resourceManager *service.ResourceManager) (*PfcpConnection, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Warn().Msgf("Can't resolve UDP address: %s", err.Error())
		return nil, err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Warn().Msgf("Can't listen UDP address: %s", err.Error())
		return nil, err
	}

	n3Addr := net.ParseIP(n3Ip)
	if n3Addr == nil {
		return nil, fmt.Errorf("failed to parse N3 IP address ID: %s", n3Ip)
	}
	log.Info().Msgf("Starting PFCP connection: %v with Node ID: %v and N3 address: %v", udpAddr, nodeId, n3Addr)

	featuresOctets := []uint8{0, 0, 0}
	// featuresOctets := []uint8{0, 0}
	// featuresOctets[0] = setBit(featuresOctets[0], 1)
	// featuresOctets[0] = setBit(featuresOctets[0], 2)
	// featuresOctets[0] = setBit(featuresOctets[0], 6)
	// featuresOctets[0] = setBit(featuresOctets[0], 7)

	featuresOctets[1] = setBit(featuresOctets[1], 0)
	if config.Conf.FeatureFTUP {
		featuresOctets[0] = setBit(featuresOctets[0], 4)
	}
	if config.Conf.FeatureUEIP {
		featuresOctets[2] = setBit(featuresOctets[2], 2)
	}

	return &PfcpConnection{
		udpConn:           udpConn,
		pfcpHandlerMap:    pfcpHandlerMap,
		associationMutex:  &sync.Mutex{},
		NodeAssociations:  map[string]*NodeAssociation{},
		nodeId:            nodeId,
		nodeAddrV4:        udpAddr.IP,
		n3Address:         n3Addr,
		mapOperations:     mapOperations,
		RecoveryTimestamp: time.Now(),
		featuresOctets:    featuresOctets,
		ResourceManager:   resourceManager,
		heartbeatFailedC:  make(chan string),
		nodes:             []string{},
	}, nil
}

func (connection *PfcpConnection) SetRemoteNodes(nodes []string) {
	connection.nodes = nodes
}

func (connection *PfcpConnection) Run() {

	ticker := time.NewTicker(time.Duration(config.Conf.AssociationSetupTimeout) * time.Second)
	buf := make([]byte, 1500)

	for {
		select {
		case <-ticker.C:
			connection.RefreshAssociations()
		case associationAddr := <-connection.heartbeatFailedC:
			connection.DeleteAssociation(associationAddr)
		default:
			connection.udpConn.SetReadDeadline(time.Now().Add(time.Second))
			n, addr, err := connection.Receive(buf)
			if err != nil {
				if err.(*net.OpError).Timeout() {
					continue
				}
				log.Warn().Msgf("Error reading from UDP socket: %s", err.Error())
				time.Sleep(1 * time.Second)
				continue
			}
			log.Debug().Msgf("Received %d bytes from %s", n, addr)
			connection.Handle(buf[:n], addr)
		}
	}
}

func (connection *PfcpConnection) Close() {
	connection.udpConn.Close()
}

func (connection *PfcpConnection) Receive(b []byte) (n int, addr *net.UDPAddr, err error) {
	return connection.udpConn.ReadFromUDP(b)
}

func (connection *PfcpConnection) Handle(b []byte, addr *net.UDPAddr) {
	err := connection.pfcpHandlerMap.Handle(connection, b, addr)
	if err != nil {
		log.Warn().Msgf("Error handling PFCP message: %s", err.Error())
	}
}

func (connection *PfcpConnection) Send(b []byte, addr *net.UDPAddr) (int, error) {
	return connection.udpConn.WriteTo(b, addr)
}

func (connection *PfcpConnection) SendMessage(msg message.Message, addr *net.UDPAddr) error {
	responseBytes := make([]byte, msg.MarshalLen())
	if err := msg.MarshalTo(responseBytes); err != nil {
		log.Warn().Msg(err.Error())
		return err
	}
	if _, err := connection.Send(responseBytes, addr); err != nil {
		log.Warn().Msg(err.Error())
		return err
	}
	return nil
}

func (connection *PfcpConnection) RefreshAssociations() {
	for _, node := range connection.nodes {
		if connection.GetAssociation(node) == nil {
			connection.sendAssociationSetupRequest(node)
		}
	}
}

// DeleteAssociation deletes an association and all sessions associated with it.
func (connection *PfcpConnection) DeleteAssociation(assocAddr string) {
	assoc := connection.GetAssociation(assocAddr)
	log.Info().Msgf("Pruning expired node association: %s", assocAddr)
	for sessionId, session := range assoc.Sessions {
		log.Info().Msgf("Deleting session: %d", sessionId)
		connection.DeleteSession(session)
	}
	delete(connection.NodeAssociations, assocAddr)
}

// DeleteSession deletes a session and all PDRs, FARs and QERs associated with it.
func (connection *PfcpConnection) DeleteSession(session *Session) {
	for _, far := range session.FARs {
		_ = connection.mapOperations.DeleteFar(far.GlobalId)
	}
	for _, qer := range session.QERs {
		_ = connection.mapOperations.DeleteQer(qer.GlobalId)
	}
	pdrContext := NewPDRCreationContext(session, connection.ResourceManager)
	for _, PDR := range session.PDRs {
		_ = pdrContext.deletePDR(PDR, connection.mapOperations)
	}
}

func (connection *PfcpConnection) GetSessionCount() int {
	count := 0
	for _, assoc := range connection.NodeAssociations {
		count += len(assoc.Sessions)
	}
	return count
}

func (connection *PfcpConnection) GetAssiciationCount() int {
	return len(connection.NodeAssociations)
}

func (connection *PfcpConnection) ReleaseResources(seID uint64) {
	if connection.ResourceManager == nil {
		return
	}

	if connection.ResourceManager.IPAM != nil {
		connection.ResourceManager.IPAM.ReleaseIP(seID)
	}

	if connection.ResourceManager.FTEIDM != nil {
		connection.ResourceManager.FTEIDM.ReleaseTEID(seID)
	}
}

func (connection *PfcpConnection) sendAssociationSetupRequest(associationAddr string) {

	AssociationSetupRequest := message.NewAssociationSetupRequest(0,
		newIeNodeIDHuawei(connection.nodeId),
		ie.NewRecoveryTimeStamp(connection.RecoveryTimestamp),
		ie.NewUPFunctionFeatures(connection.featuresOctets[:]...),
		ie.NewVendorSpecificIE(32787, 2011, []byte{0x02, 0x00, 0x08, 0x30, 0x30, 0x30, 0x34, 0x64, 0x67, 0x77, 0x34, connection.n3Address.To4()[0], connection.n3Address.To4()[1], connection.n3Address.To4()[2], connection.n3Address.To4()[3]}),
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
	if err := connection.SendMessage(AssociationSetupRequest, udpAddr); err != nil {
		log.Info().Msgf("Failed to send Association Setup Request: %s\n", err.Error())
	}
}
