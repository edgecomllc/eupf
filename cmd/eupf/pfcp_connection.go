package main

import (
	"log"
	"net"
)

type PfcpConnection struct {
	udpConn        *net.UDPConn
	pfcpHandlerMap PfcpHanderMap
}

func CreatePfcpConnection(addr string, pfcpHandlerMap PfcpHanderMap) *PfcpConnection {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf("Can't resolve UDP address: %s", err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Can't listen UDP address: %s", err)
	}
	return &PfcpConnection{
		udpConn:        udpConn,
		pfcpHandlerMap: pfcpHandlerMap,
	}
}

func (c *PfcpConnection) Run() {
	for {
		buf := make([]byte, 1500)
		n, addr, err := c.udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP socket: %s", err)
			continue
		}
		log.Printf("Received %d bytes from %s", n, c.udpConn.RemoteAddr())
		go c.pfcpHandlerMap.Handle(c.udpConn, addr, buf[:n])
	}
}

func (c *PfcpConnection) Close() {
	c.udpConn.Close()
}
