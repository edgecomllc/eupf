package main

import (
	"log"
	"net"
)

type NodeAssociationMap map[string]RemoteNode

type RemoteNode struct {
	ID   string
	Addr string
}

type PfcpConnection struct {
	udpConn          *net.UDPConn
	pfcpHandlerMap   PfcpHanderMap
	nodeAssociations NodeAssociationMap
}

func CreatePfcpConnection(addr string, pfcpHandlerMap PfcpHanderMap) (*PfcpConnection, error) {
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
	return &PfcpConnection{
		udpConn:        udpConn,
		pfcpHandlerMap: pfcpHandlerMap,
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
