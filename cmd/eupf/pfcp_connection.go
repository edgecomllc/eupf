package main

import (
	"log"
	"net"
)

type PfcpConnection struct {
	udpConn        *net.UDPConn
	pfcpHandlerMap PfcpHanderMap
}

func CreatePfcpConnection(addr string, pfcpHandlerMap PfcpHanderMap) (*PfcpConnection, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf("Can't resolve UDP address: %s", err)
		return nil, err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Can't listen UDP address: %s", err)
		return nil, err
	}
	return &PfcpConnection{
		udpConn:        udpConn,
		pfcpHandlerMap: pfcpHandlerMap,
	}, nil
}

func (c *PfcpConnection) Run() {
	buf := make([]byte, 1500)
	for {
		n, addr, err := c.udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP socket: %s", err)
			continue
		}
		log.Printf("Received %d bytes from %s", n, c.udpConn.RemoteAddr())
		go c.pfcpHandlerMap.Handle(c, addr, buf[:n])
	}
}

func (c *PfcpConnection) Close() {
	c.udpConn.Close()
}

func (c *PfcpConnection) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	return c.udpConn.WriteToUDP(b, addr)
}