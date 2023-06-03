package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/wmnsk/go-pfcp/message"
)

type IdTracker struct {
	bitmap  *roaring.Bitmap
	maxSize uint32
}

func NewIdTracker(size uint32) *IdTracker {
	newBitmap := roaring.NewBitmap()
	newBitmap.Flip(0, uint64(size))

	return &IdTracker{
		bitmap:  newBitmap,
		maxSize: size,
	}
}

func (t *IdTracker) GetNext() (next uint32, err error) {

	i := t.bitmap.Iterator()
	if i.HasNext() {
		next := i.Next()
		t.bitmap.Remove(next)
		return next, nil
	}

	return 0, errors.New("pool is empty")
}

func (t *IdTracker) Release(id uint32) {
	if id >= t.maxSize {
		return
	}

	t.bitmap.Add(id)
}

type Session struct {
	LocalSEID    uint64
	RemoteSEID   uint64
	UplinkPDRs   map[uint32]SPDRInfo
	DownlinkPDRs map[uint32]SPDRInfo
	FARs         map[uint32]SFarInfo
	QERs         map[uint32]SQerInfo
}

type SPDRInfo struct {
	PdrInfo PdrInfo
	Teid    uint32
	Ipv4    net.IP
}

type SFarInfo struct {
	FarInfo  FarInfo
	GlobalId uint32
}

type SQerInfo struct {
	QerInfo  QerInfo
	GlobalId uint32
}

func (s *Session) NewFar(smfId uint32, ebpfId uint32, farInfo FarInfo) {
	s.FARs[smfId] = SFarInfo{
		FarInfo:  farInfo,
		GlobalId: ebpfId,
	}
}

func (s *Session) UpdateFar(smfId uint32, farInfo FarInfo) {
	sFarInfo := s.FARs[smfId]
	sFarInfo.FarInfo = farInfo
	s.FARs[smfId] = sFarInfo
}

func (s *Session) GetFar(smfId uint32) SFarInfo {
	return s.FARs[smfId]
}

func (s *Session) RemoveFar(smfId uint32) SFarInfo {
	sFarInfo := s.FARs[smfId]
	delete(s.FARs, smfId)
	return sFarInfo
}

func (s *Session) NewQer(smfId uint32, ebpfId uint32, qerInfo QerInfo) {
	s.QERs[smfId] = SQerInfo{
		QerInfo:  qerInfo,
		GlobalId: ebpfId,
	}
}

func (s *Session) UpdateQer(smfId uint32, qerInfo QerInfo) {
	sQerInfo := s.QERs[smfId]
	sQerInfo.QerInfo = qerInfo
	s.QERs[smfId] = sQerInfo
}

func (s *Session) GetQer(smfId uint32) SQerInfo {
	return s.QERs[smfId]
}

func (s *Session) RemoveQer(smfId uint32) SQerInfo {
	sQerInfo := s.QERs[smfId]
	delete(s.QERs, smfId)
	return sQerInfo
}

func (s *Session) PutUplinkPDR(smfId uint32, info SPDRInfo) {
	s.UplinkPDRs[smfId] = info
}

func (s *Session) GetUplinkPDR(smfId uint16) SPDRInfo {
	return s.UplinkPDRs[uint32(smfId)]
}

func (s *Session) RemoveUplinkPDR(smfId uint32) SPDRInfo {
	sPdrInfo := s.UplinkPDRs[smfId]
	delete(s.UplinkPDRs, smfId)
	return sPdrInfo
}

func (s *Session) PutDownlinkPDR(smfId uint32, info SPDRInfo) {
	s.DownlinkPDRs[smfId] = info
}

func (s *Session) GetDownlinkPDR(smfId uint16) SPDRInfo {
	return s.DownlinkPDRs[uint32(smfId)]
}

func (s *Session) RemoveDownlinkPDR(smfId uint32) SPDRInfo {
	sPdrInfo := s.DownlinkPDRs[smfId]
	delete(s.DownlinkPDRs, smfId)
	return sPdrInfo
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
