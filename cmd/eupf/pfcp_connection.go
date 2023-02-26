package main

import (
	"log"
	"net"

	"github.com/wmnsk/go-pfcp/message"
)

type Session struct {
	SEID uint64
}

type SessionMap map[uint64]Session

type NodeAssociationMap map[string]NodeAssociation

type NodeAssociation struct {
	ID       string
	Addr     string
	Sessions SessionMap
}

type PfcpConnection struct {
	udpConn          *net.UDPConn
	pfcpHandlerMap   PfcpHanderMap
	nodeAssociations NodeAssociationMap
	nodeId           string
	nodeAddrV4       net.IP
}

func CreatePfcpConnection(addr string, pfcpHandlerMap PfcpHanderMap, nodeId string) (*PfcpConnection, error) {
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
	response_bytes := make([]byte, msg.MarshalLen())
	if err := msg.MarshalTo(response_bytes); err != nil {
		log.Print(err)
		return err
	}
	if _, err := connection.Send(response_bytes, addr); err != nil {
		log.Print(err)
		return err
	}
	return nil
}
