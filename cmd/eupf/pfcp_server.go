package main

import (
	"log"
	"net"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

// #TODO:
func CreateAndRunPfcpServer(addr string) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf("Can't resolve UDP address: %s", err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Can't listen UDP address: %s", err)
	}
	log.Printf("Listening for PFCP on %s", udpConn.LocalAddr())
	defer udpConn.Close()
	for {
		buf := make([]byte, 1500)
		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP socket: %s", err)
			continue
		}
		log.Printf("Received %d bytes from %s", n, udpConn.RemoteAddr())
		go handlePfcpMessage(udpConn, addr, buf[:n])
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

	switch msg.MessageType() {
	case message.MsgTypeHeartbeatRequest:
		handlePfcpHeartbeatRequest(conn, addr, msg)
	default:
		log.Printf("got unexpected message %s: %s, from: %s", msg.MessageTypeName(), msg, addr)
	}
}

func handlePfcpHeartbeatRequest(conn *net.UDPConn, addr *net.UDPAddr, msg message.Message) {
	hbreq := msg.(*message.HeartbeatRequest)
	// hbreq, ok := msg.(*message.HeartbeatRequest)
	// if !ok {
	// 	log.Printf("got unexpected message: %s, from: %s", msg.MessageTypeName(), addr)
	// 	return
	// }
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

	if _, err := conn.WriteToUDP(hbres, addr); err != nil {
		log.Fatal(err)
	}
	log.Printf("sent Heartbeat Response to: %s", addr)
}
