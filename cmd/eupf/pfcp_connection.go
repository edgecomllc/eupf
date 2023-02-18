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

func (c *PfcpConnection) Run() {
	buf := make([]byte, 1500)
	for {
		n, _, err := c.udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP socket: %s", err)
			continue
		}
		log.Printf("Received %d bytes from %s", n, c.udpConn.RemoteAddr())
		c.pfcpHandlerMap.Handle(c, buf[:n])
	}
}

func (c *PfcpConnection) Close() {
	c.udpConn.Close()
}

func (c *PfcpConnection) Send(b []byte) (int, error) {
	return c.udpConn.WriteToUDP(b, c.RemoteAddr())
}

func (c *PfcpConnection) RemoteAddr() *net.UDPAddr {
	return c.udpConn.RemoteAddr().(*net.UDPAddr)
}
