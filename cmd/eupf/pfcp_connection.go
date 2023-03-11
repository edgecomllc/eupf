package main

import (
	"log"
	"net"

	"github.com/wmnsk/go-pfcp/message"
)

type Session struct {
	LocalSEID  uint64
	RemoteSEID uint64
	updrs      map[uint32]PdrInfo // Uplink PDRs
	dpdrs      map[string]PdrInfo // Downlink PDRs
	fars       map[uint32]FarInfo
}

func (s *Session) CreateUpLinkPDR(bpfObjects *BpfObjects, teid uint32, pdrInfo PdrInfo) error {
	if err := bpfObjects.PutPdrUpLink(teid, pdrInfo); err != nil {
		log.Printf("Can't put uplink PDR: %s", err)
		return err
	}
	s.updrs[teid] = pdrInfo
	return nil
}

func (s *Session) CreateDownLinkPDR(bpfObjects *BpfObjects, ipv4 net.IP, pdrInfo PdrInfo) error {
	if err := bpfObjects.PutPdrDownLink(ipv4, pdrInfo); err != nil {
		log.Printf("Can't put uplink PDR: %s", err)
		return err
	}
	s.dpdrs[ipv4.String()] = pdrInfo
	return nil
}

func (s *Session) CreateFAR(bpfObjects *BpfObjects, id uint32, farInfo FarInfo) error {
	if err := bpfObjects.PutFar(id, farInfo); err != nil {
		log.Printf("Can't put FAR: %s", err)
		return err
	}
	s.fars[id] = farInfo
	return nil
}

func (s *Session) UpdateUpLinkPDR(bpfObjects *BpfObjects, teid uint32, pdrInfo PdrInfo) error {
	if err := bpfObjects.UpdatePdrUpLink(teid, pdrInfo); err != nil {
		log.Printf("Can't update uplink PDR: %s", err)
		return err
	}
	s.updrs[teid] = pdrInfo
	return nil
}

func (s *Session) UpdateDownLinkPDR(bpfObjects *BpfObjects, ipv4 net.IP, pdrInfo PdrInfo) error {
	if err := bpfObjects.UpdatePdrDownLink(ipv4, pdrInfo); err != nil {
		log.Printf("Can't update uplink PDR: %s", err)
		return err
	}
	s.dpdrs[ipv4.String()] = pdrInfo
	return nil
}

func (s *Session) UpdateFAR(bpfObjects *BpfObjects, id uint32, farInfo FarInfo) error {
	if err := bpfObjects.UpdateFar(id, farInfo); err != nil {
		log.Printf("Can't update FAR: %s", err)
		return err
	}
	s.fars[id] = farInfo
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
	bpfObjects       *BpfObjects
}

func CreatePfcpConnection(addr string, pfcpHandlerMap PfcpHanderMap, nodeId string, bpfObjects *BpfObjects) (*PfcpConnection, error) {
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
		bpfObjects:       bpfObjects,
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
