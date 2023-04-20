package main

import (
	"log"
	"net"

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

func (s *Session) CreateUpLinkPDR(pdrId uint16, pdrInfo SPDRInfo) {
	s.UplinkPDRs[uint32(pdrId)] = pdrInfo
}

func (s *Session) CreateDownLinkPDR(pdrId uint16, pdrInfo SPDRInfo) {
	s.DownlinkPDRs[uint32(pdrId)] = pdrInfo
}

func (s *Session) CreateFAR(id uint32, farInfo FarInfo) {
	s.FARs[id] = farInfo
}

func (s *Session) CreateQER(id uint32, qerInfo QerInfo) {
	s.QERs[id] = qerInfo
}

func (s *Session) UpdateUpLinkPDR(pdrId uint16, pdrInfo SPDRInfo) {
	s.UplinkPDRs[uint32(pdrId)] = pdrInfo
}

func (s *Session) UpdateDownLinkPDR(pdrId uint16, pdrInfo SPDRInfo) {
	s.DownlinkPDRs[uint32(pdrId)] = pdrInfo
}

func (s *Session) UpdateFAR(id uint32, farInfo FarInfo) {
	s.FARs[id] = farInfo
}

func (s *Session) UpdateQER(id uint32, qerInfo QerInfo) {
	s.QERs[id] = qerInfo
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
	pfcpHandlerMap   PfcpHanderMap
	nodeAssociations NodeAssociationMap
	nodeId           string
	nodeAddrV4       net.IP
	n3Address        net.IP
	mapOperations    ForwardingPlaneController
}

func CreatePfcpConnection(addr string, pfcpHandlerMap PfcpHanderMap, nodeId string, n3Ip string, mapOperations ForwardingPlaneController) (*PfcpConnection, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Panicf("Can't resolve UDP address: %s", err)
		return nil, err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Printf("Can't listen UDP address: %s", err)
		return nil, err
	}

	log.Printf("Start PFCP connection: %s", addr)
	addrv4, err := net.ResolveIPAddr("ip4", nodeId)
	if err != nil {
		return nil, err
	}

	log.Printf("Parsing N3 address")
	n3Addr, err := net.ResolveIPAddr("ip4", n3Ip)
	if err != nil {
		return nil, err
	}

	return &PfcpConnection{
		udpConn:          udpConn,
		pfcpHandlerMap:   pfcpHandlerMap,
		nodeAssociations: NodeAssociationMap{},
		nodeId:           nodeId,
		nodeAddrV4:       addrv4.IP,
		n3Address:        n3Addr.IP,
		mapOperations:    mapOperations,
	}, nil
}

func (connection *PfcpConnection) Run() {
	buf := make([]byte, 1500)
	for {
		n, addr, err := connection.udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP socket: %s", err)
			continue
		}
		log.Printf("Received %d bytes from %s", n, addr)
		connection.pfcpHandlerMap.Handle(connection, buf[:n], addr)
	}
}

func (connection *PfcpConnection) Close() {
	connection.udpConn.Close()
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
