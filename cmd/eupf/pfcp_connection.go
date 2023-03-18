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
}

type SPDRInfo struct {
	PdrInfo PdrInfo
	Teid    uint32
	Ipv4    net.IP
}

func (s *Session) CreateUpLinkPDR(bpfMapOperations ForwardingPlaneController, pdrId uint16, pdrInfo SPDRInfo) error {
	if err := bpfMapOperations.PutPdrUpLink(pdrInfo.Teid, pdrInfo.PdrInfo); err != nil {
		log.Printf("Can't put uplink PDR: %s", err)
		return err
	}
	s.UplinkPDRs[uint32(pdrId)] = pdrInfo
	return nil
}

func (s *Session) CreateDownLinkPDR(bpfMapOperations ForwardingPlaneController, pdrId uint16, pdrInfo SPDRInfo) error {
	if err := bpfMapOperations.PutPdrDownLink(pdrInfo.Ipv4, pdrInfo.PdrInfo); err != nil {
		log.Printf("Can't put uplink PDR: %s", err)
		return err
	}
	s.DownlinkPDRs[uint32(pdrId)] = pdrInfo
	return nil
}

func (s *Session) CreateFAR(bpfMapOperations ForwardingPlaneController, id uint32, farInfo FarInfo) error {
	if err := bpfMapOperations.PutFar(id, farInfo); err != nil {
		log.Printf("Can't put FAR: %s", err)
		return err
	}
	s.FARs[id] = farInfo
	return nil
}

func (s *Session) UpdateUpLinkPDR(bpfMapOperations ForwardingPlaneController, pdrId uint16, pdrInfo SPDRInfo) error {
	if err := bpfMapOperations.UpdatePdrUpLink(pdrInfo.Teid, pdrInfo.PdrInfo); err != nil {
		log.Printf("Can't update uplink PDR: %s", err)
		return err
	}
	s.UplinkPDRs[uint32(pdrId)] = pdrInfo
	return nil
}

func (s *Session) UpdateDownLinkPDR(bpfMapOperations ForwardingPlaneController, pdrId uint16, pdrInfo SPDRInfo) error {
	if err := bpfMapOperations.UpdatePdrDownLink(pdrInfo.Ipv4, pdrInfo.PdrInfo); err != nil {
		log.Printf("Can't update uplink PDR: %s", err)
		return err
	}
	s.DownlinkPDRs[uint32(pdrId)] = pdrInfo
	return nil
}

func (s *Session) UpdateFAR(bpfMapOperations ForwardingPlaneController, id uint32, farInfo FarInfo) error {
	if err := bpfMapOperations.UpdateFar(id, farInfo); err != nil {
		log.Printf("Can't update FAR: %s", err)
		return err
	}
	s.FARs[id] = farInfo
	return nil
}

func (s *Session) RemoveUplinkPDR(bpfMapOperations ForwardingPlaneController, pdrId uint16) error {
	delete(s.UplinkPDRs, uint32(pdrId))
	return bpfMapOperations.DeletePdrUpLink(s.UplinkPDRs[uint32(pdrId)].Teid)
}

func (s *Session) RemoveDownlinkPDR(bpfMapOperations ForwardingPlaneController, pdrId uint16) error {
	delete(s.DownlinkPDRs, uint32(pdrId))
	return bpfMapOperations.DeletePdrDownLink(s.DownlinkPDRs[uint32(pdrId)].Ipv4)
}

// Removes PDR from one of the maps. If same PDR is present in both maps, it will be removed from both maps.
func (s *Session) RemovePDR(bpfMapOperations ForwardingPlaneController, pdrId uint16) error {
	if err := s.RemoveUplinkPDR(bpfMapOperations, pdrId); err != nil {
		return err
	}
	if err := s.RemoveDownlinkPDR(bpfMapOperations, pdrId); err != nil {
		return err
	}
	return nil
}

func (s *Session) RemoveFAR(bpfMapOperations ForwardingPlaneController, farId uint32) error {
	delete(s.FARs, farId)
	return bpfMapOperations.DeleteFar(farId)
}

func (s *Session) Cleanup(bpfMapOperations ForwardingPlaneController) error {
	for _, pdrInfo := range s.UplinkPDRs {
		if err := bpfMapOperations.DeletePdrUpLink(pdrInfo.Teid); err != nil {
			return err
		}
	}
	for _, pdrInfo := range s.DownlinkPDRs {
		if err := bpfMapOperations.DeletePdrDownLink(pdrInfo.Ipv4); err != nil {
			return err
		}
	}
	for id := range s.FARs {
		if err := bpfMapOperations.DeleteFar(id); err != nil {
			return err
		}
	}
	return nil
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
	bpfMapOperations ForwardingPlaneController
}

func CreatePfcpConnection(addr string, pfcpHandlerMap PfcpHanderMap, nodeId string, bpfMapOperations ForwardingPlaneController) (*PfcpConnection, error) {
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
	addrv4, err := net.ResolveIPAddr("ip4", nodeId)
	if err != nil {
		return nil, err
	}
	return &PfcpConnection{
		udpConn:          udpConn,
		pfcpHandlerMap:   pfcpHandlerMap,
		nodeAssociations: NodeAssociationMap{},
		nodeId:           nodeId,
		nodeAddrV4:       addrv4.IP,
		bpfMapOperations: bpfMapOperations,
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
