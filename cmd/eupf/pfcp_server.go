package main

import (
	"log"
	"net"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PfcpServer struct {
	Conn *net.UDPConn
}

func CreatePfcpServer(addr string) *PfcpServer {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf("Can't resolve UDP address: %s", err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Can't listen UDP address: %s", err)
	}
	return &PfcpServer{Conn: udpConn}
}

func (u *PfcpServer) Run() {
	defer u.Conn.Close()
	for {
		buf := make([]byte, 1500)
		n, addr, err := u.Conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP socket: %s", err)
			continue
		}
		log.Printf("Received %d bytes from %s", n, u.Conn.RemoteAddr())
		go handlePfcpMessage(u.Conn, addr, buf[:n])
	}
}

func handlePfcpMessage(conn *net.UDPConn, addr *net.UDPAddr, buf []byte) {
	log.Printf("Handling PFCP message from %s", addr)
	msg, err := message.Parse(buf)
	if err != nil {
		log.Printf("ignored undecodable message: %x, error: %s", buf, err)
		return
	}
	log.Printf("Parsed PFCP message: %s", msg)

	hbreq, ok := msg.(*message.HeartbeatRequest)
	if !ok {
		log.Printf("got unexpected message: %s, from: %s", msg.MessageTypeName(), addr)
		return
	}

	ts, err := hbreq.RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		log.Printf("got Heartbeat Request with invalid TS: %s, from: %s", err, addr)
		return
	} else {
		log.Printf("got Heartbeat Request with TS: %s, from: %s", ts, addr)
	}

	var seq uint32 = 1
	hbres, err := message.NewHeartbeatResponse(seq, ie.NewRecoveryTimeStamp(time.Now())).Marshal()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := conn.WriteTo(hbres, addr); err != nil {
		log.Fatal(err)
	}
	log.Printf("sent Heartbeat Response to: %s", addr)
}
