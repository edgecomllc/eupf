package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/wmnsk/go-pfcp/message"
)

type Session struct {
	LocalSEID    uint64
	RemoteSEID   uint64
	UplinkPDRs   map[uint32]SPDRInfo
	DownlinkPDRs map[uint32]SPDRInfo
	FARs         map[uint32]FarInfo
	QERs         map[uint32]QerInfo
}

type SPDRInfo struct {
	PdrInfo PdrInfo
	Teid    uint32
	Ipv4    net.IP
}

func (s *Session) GetFAR(id uint32) FarInfo {
	return s.FARs[id]
}

func (s *Session) GetQER(id uint32) QerInfo {
	return s.QERs[id]
}

func (s *Session) GetUplinkPDR(pdrId uint16) SPDRInfo {
	return s.UplinkPDRs[uint32(pdrId)]
}

func (s *Session) GetDownlinkPDR(pdrId uint16) SPDRInfo {
	return s.DownlinkPDRs[uint32(pdrId)]
}

func (s *Session) PutFAR(id uint32, farInfo FarInfo) {
	s.FARs[id] = farInfo
}

func (s *Session) PutQER(id uint32, qerInfo QerInfo) {
	s.QERs[id] = qerInfo
}

func (s *Session) PutUplinkPDR(pdrId uint16, pdrInfo SPDRInfo) {
	s.UplinkPDRs[uint32(pdrId)] = pdrInfo
}

func (s *Session) PutDownlinkPDR(pdrId uint16, pdrInfo SPDRInfo) {
	s.DownlinkPDRs[uint32(pdrId)] = pdrInfo
}

func (s *Session) RemoveUplinkPDR(pdrId uint16) {
	delete(s.UplinkPDRs, uint32(pdrId))
}

func (s *Session) RemoveDownlinkPDR(pdrId uint16) {
	delete(s.DownlinkPDRs, uint32(pdrId))
}

func (s *Session) RemoveFAR(farId uint32) {
	delete(s.FARs, farId)
}

func (s *Session) RemoveQER(qerId uint32) {
	delete(s.QERs, qerId)
}

type SessionMap map[uint64]Session

type NodeAssociationMap map[string]NodeAssociation

type NodeAssociation struct {
	ID            string
	Addr          string
	NextSessionID uint64
	Sessions      SessionMap
}

func (association *NodeAssociation) NewLocalSEID() uint64 {
	association.NextSessionID += 1
	return association.NextSessionID
}

type PfcpConnection struct {
	udpConn          *net.UDPConn
	pfcpHandlerMap   PfcpHandlerMap
	nodeAssociations NodeAssociationMap
	nodeId           string
	nodeAddrV4       net.IP
	n3Address        net.IP
	mapOperations    ForwardingPlaneController
}

func CreatePfcpConnection(addr string, pfcpHandlerMap PfcpHandlerMap, nodeId string, n3Ip string, mapOperations ForwardingPlaneController) (*PfcpConnection, error) {
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
		nodeAssociations: NodeAssociationMap{},
		nodeId:           nodeId,
		nodeAddrV4:       addrv4,
		n3Address:        n3Addr,
		mapOperations:    mapOperations,
	}, nil
}

func (connection *PfcpConnection) Run() {
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
