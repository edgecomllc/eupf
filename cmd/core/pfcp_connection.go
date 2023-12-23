package core

import (
	"fmt"
	"net"
	"time"

	"github.com/edgecomllc/eupf/cmd/core/service"

	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/rs/zerolog/log"

	"github.com/wmnsk/go-pfcp/message"
)

type PfcpConnection struct {
	udpConn           *net.UDPConn
	pfcpHandlerMap    PfcpHandlerMap
	NodeAssociations  map[string]*NodeAssociation
	nodeId            string
	nodeAddrV4        net.IP
	n3Address         net.IP
	mapOperations     ebpf.ForwardingPlaneController
	RecoveryTimestamp time.Time
	ResourceManager   *service.ResourceManager
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

	return &PfcpConnection{
		udpConn:           udpConn,
		pfcpHandlerMap:    pfcpHandlerMap,
		NodeAssociations:  map[string]*NodeAssociation{},
		nodeId:            nodeId,
		nodeAddrV4:        udpAddr.IP,
		n3Address:         n3Addr,
		mapOperations:     mapOperations,
		RecoveryTimestamp: time.Now(),
		ResourceManager:   resourceManager,
	}, nil
}

func (connection *PfcpConnection) Run() {
	go func() {
		for {
			connection.RefreshAssociations()
			time.Sleep(time.Duration(config.Conf.HeartbeatInterval) * time.Second)
		}
	}()
	buf := make([]byte, 1500)
	for {
		n, addr, err := connection.Receive(buf)
		if err != nil {
			log.Warn().Msgf("Error reading from UDP socket: %s", err.Error())
			time.Sleep(1 * time.Second)
			continue
		}
		log.Debug().Msgf("Received %d bytes from %s", n, addr)
		connection.Handle(buf[:n], addr)
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
	for _, assoc := range connection.NodeAssociations {
		if !assoc.HeartbeatsActive {
			go assoc.ScheduleHeartbeat(connection)
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
