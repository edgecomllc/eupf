package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/wmnsk/go-pfcp/message"
)

type IdTracker struct {
	bitmap *roaring.Bitmap
}

func NewIdTracker() *IdTracker {
	return &IdTracker{
		bitmap: roaring.NewBitmap(),
	}
}

func (t *IdTracker) GetNext() uint32 {
	newId := uint32(0)
	// We have relatively few IDs, so linear search is fine
	for ; ; newId++ {
		if exists := t.bitmap.Contains(newId); !exists {
			break
		}
	}
	t.bitmap.Add(newId)
	return newId
}

func (t *IdTracker) Release(id uint32) {
	t.bitmap.Remove(id)
}

var (
	FarIdTracker         = NewIdTracker()
	QerIdTracker         = NewIdTracker()
	UplinkPdrIdTracker   = NewIdTracker()
	DownlinkPdrIdTracker = NewIdTracker()
)

type Session struct {
	LocalSEID    uint64
	RemoteSEID   uint64
	UplinkPDRs   map[uint32]SPDRInfo
	DownlinkPDRs map[uint32]SPDRInfo
	FARs         map[uint32]SFarInfo
	QERs         map[uint32]SQerInfo
}

type SPDRInfo struct {
	PdrInfo  PdrInfo
	Teid     uint32
	Ipv4     net.IP
	GlobalId uint32
}

type SFarInfo struct {
	FarInfo  FarInfo
	GlobalId uint32
}

type SQerInfo struct {
	QerInfo  QerInfo
	GlobalId uint32
}

func (s *Session) GetFAR(id uint32) SFarInfo {
	return s.FARs[id]
}

func (s *Session) GetQER(id uint32) SQerInfo {
	return s.QERs[id]
}

func (s *Session) GetUplinkPDR(pdrId uint16) SPDRInfo {
	return s.UplinkPDRs[uint32(pdrId)]
}

func (s *Session) GetDownlinkPDR(pdrId uint16) SPDRInfo {
	return s.DownlinkPDRs[uint32(pdrId)]
}

// This code duplication is ugly, but i'm not enough well versed in Golang to avoid it.

func (s *Session) PutFAR(id uint32, farInfo FarInfo) uint32 {
	// if far is already present, update and return global id, else assign new global id
	if sFarInfo, ok := s.FARs[id]; ok {
		s.FARs[id] = SFarInfo{
			FarInfo:  farInfo,
			GlobalId: sFarInfo.GlobalId,
		}
		return sFarInfo.GlobalId
	} else {
		newId := FarIdTracker.GetNext()
		s.FARs[id] = SFarInfo{
			FarInfo:  farInfo,
			GlobalId: newId,
		}
		return newId
	}
}

func (s *Session) PutQER(id uint32, qerInfo QerInfo) uint32 {
	if sQerInfo, ok := s.QERs[id]; ok {
		s.QERs[id] = SQerInfo{
			QerInfo:  qerInfo,
			GlobalId: sQerInfo.GlobalId,
		}
		return sQerInfo.GlobalId
	} else {
		newId := QerIdTracker.GetNext()
		s.QERs[id] = SQerInfo{
			QerInfo:  qerInfo,
			GlobalId: newId,
		}
		return newId
	}
}

func (s *Session) PutUplinkPDR(id uint32, pdrInfo SPDRInfo) uint32 {
	if sQerInfo, ok := s.UplinkPDRs[id]; ok {
		oldGlobalId := sQerInfo.GlobalId
		pdrInfo.GlobalId = oldGlobalId
		s.UplinkPDRs[id] = pdrInfo
		return oldGlobalId
	} else {
		newId := UplinkPdrIdTracker.GetNext()
		pdrInfo.GlobalId = newId
		s.UplinkPDRs[id] = pdrInfo
		return newId
	}
}

func (s *Session) PutDownlinkPDR(id uint32, pdrInfo SPDRInfo) uint32 {
	if sQerInfo, ok := s.DownlinkPDRs[id]; ok {
		oldGlobalId := sQerInfo.GlobalId
		pdrInfo.GlobalId = oldGlobalId
		s.DownlinkPDRs[id] = pdrInfo
		return oldGlobalId
	} else {
		newId := DownlinkPdrIdTracker.GetNext()
		pdrInfo.GlobalId = newId
		s.DownlinkPDRs[id] = pdrInfo
		return newId
	}
}

func (s *Session) RemoveUplinkPDR(id uint32) uint32 {
	oldGlobalId := s.UplinkPDRs[id].GlobalId
	UplinkPdrIdTracker.Release(oldGlobalId)
	delete(s.UplinkPDRs, id)
	return oldGlobalId
}

func (s *Session) RemoveDownlinkPDR(id uint32) uint32 {
	oldGlobalId := s.DownlinkPDRs[id].GlobalId
	DownlinkPdrIdTracker.Release(oldGlobalId)
	delete(s.DownlinkPDRs, id)
	return oldGlobalId
}

func (s *Session) RemoveFAR(id uint32) uint32 {
	oldGlobalId := s.FARs[id].GlobalId
	FarIdTracker.Release(oldGlobalId)
	delete(s.FARs, id)
	return oldGlobalId
}

func (s *Session) RemoveQER(id uint32) uint32 {
	oldGlobalId := s.QERs[id].GlobalId
	QerIdTracker.Release(oldGlobalId)
	delete(s.QERs, id)
	return oldGlobalId
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
