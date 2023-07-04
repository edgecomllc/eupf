package core

import (
	"fmt"
	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/edgecomllc/eupf/cmd/ebpf"
	"log"
	"net"
	"time"

	"github.com/wmnsk/go-pfcp/message"
)

type PfcpConnection struct {
	udpConn          *net.UDPConn
	pfcpHandlerMap   PfcpHandlerMap
	NodeAssociations map[string]*NodeAssociation
	nodeId           string
	nodeAddrV4       net.IP
	n3Address        net.IP
	mapOperations    ebpf.ForwardingPlaneController
}

func (connection *PfcpConnection) GetAssociation(assocAddr string) *NodeAssociation {
	if assoc, ok := connection.NodeAssociations[assocAddr]; ok {
		return assoc
	}
	return nil
}

func CreatePfcpConnection(addr string, pfcpHandlerMap PfcpHandlerMap, nodeId string, n3Ip string, mapOperations ebpf.ForwardingPlaneController) (*PfcpConnection, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Panicf("Can't resolve UDP address: %s", err.Error())
		return nil, err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Printf("Can't listen UDP address: %s", err.Error())
		return nil, err
	}

	addrv4 := net.ParseIP(nodeId)
	if addrv4 == nil {
		return nil, fmt.Errorf("failed to parse Node ID: %s", nodeId)
	}

	n3Addr := net.ParseIP(n3Ip)
	if n3Addr == nil {
		return nil, fmt.Errorf("failed to parse N3 IP address ID: %s", n3Ip)
	}
	log.Printf("Starting PFCP connection: %v with Node ID: %v and N3 address: %v", udpAddr, addrv4, n3Addr)

	return &PfcpConnection{
		udpConn:          udpConn,
		pfcpHandlerMap:   pfcpHandlerMap,
		NodeAssociations: map[string]*NodeAssociation{},
		nodeId:           nodeId,
		nodeAddrV4:       addrv4,
		n3Address:        n3Addr,
		mapOperations:    mapOperations,
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
			log.Printf("Error reading from UDP socket: %s", err.Error())
			time.Sleep(1 * time.Second)
			continue
		}
		log.Printf("Received %d bytes from %s", n, addr)
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
		log.Printf("Error handling PFCP message: %s", err.Error())
	}
}

func (connection *PfcpConnection) Send(b []byte, addr *net.UDPAddr) (int, error) {
	return connection.udpConn.WriteTo(b, addr)
}

func (connection *PfcpConnection) SendMessage(msg message.Message, addr *net.UDPAddr) error {
	responseBytes := make([]byte, msg.MarshalLen())
	if err := msg.MarshalTo(responseBytes); err != nil {
		log.Print(err)
		return err
	}
	if _, err := connection.Send(responseBytes, addr); err != nil {
		log.Print(err)
		return err
	}
	return nil
}

// RefreshAssociations checks for expired associations and schedules heartbeats for those that are not expired.
func (connection *PfcpConnection) RefreshAssociations() {
	for assocAddr, assoc := range connection.NodeAssociations {
		if assoc.IsExpired() {
			log.Printf("Pruning expired node association: %s", assocAddr)
			connection.DeleteAssociation(assocAddr)
		}
		if !assoc.IsHeartbeatScheduled() {
			ScheduleHeartbeatRequest(time.Duration(config.Conf.HeartbeatTimeout)*time.Second, connection, assocAddr)
		}
	}
}

// DeleteAssociation deletes an association and all sessions associated with it.
func (connection *PfcpConnection) DeleteAssociation(assocAddr string) {
	assoc := connection.GetAssociation(assocAddr)
	log.Printf("Pruning expired node association: %s", assocAddr)
	for sessionId, session := range assoc.Sessions {
		log.Printf("Deleting session: %d", sessionId)
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
	for _, uplinkPdr := range session.UplinkPDRs {
		_ = connection.mapOperations.DeletePdrUpLink(uplinkPdr.Teid)
	}
	for _, downlinkPdr := range session.DownlinkPDRs {
		if downlinkPdr.Ipv4 != nil {
			_ = connection.mapOperations.DeletePdrDownLink(downlinkPdr.Ipv4)
		}
		if downlinkPdr.Ipv4 != nil {
			_ = connection.mapOperations.DeleteDownlinkPdrIp6(downlinkPdr.Ipv6)
		}
	}
}
