package main

import (
	"log"
	"net"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PfcpFunc func(conn *PfcpConnection, addr *net.UDPAddr, msg message.Message)

type PfcpHanderMap map[uint8]PfcpFunc

func (h PfcpHanderMap) Handle(conn *PfcpConnection, addr *net.UDPAddr, buf []byte) {
	log.Printf("Handling PFCP message from %s", addr)
	msg, err := message.Parse(buf)
	if err != nil {
		log.Printf("ignored undecodable message: %x, error: %s", buf, err)
		return
	}
	log.Printf("Parsed PFCP message: %s", msg)
	if handler, ok := h[msg.MessageType()]; ok {
		handler(conn, addr, msg)
	} else {
		log.Printf("got unexpected message %s: %s, from: %s", msg.MessageTypeName(), msg, addr)
	}
}

func handlePfcpHeartbeatRequest(conn *PfcpConnection, addr *net.UDPAddr, msg message.Message) {
	hbreq := msg.(*message.HeartbeatRequest)
	ts, err := hbreq.RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		log.Printf("got Heartbeat Request with invalid TS: %s, from: %s", err, addr)
		return
	} else {
		log.Printf("got Heartbeat Request with TS: %s, from: %s", ts, addr)
	}

	// #TODO: add sequence tracking for individual sessions
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

func handlePfcpAssociationSetupRequest(conn *PfcpConnection, addr *net.UDPAddr, msg message.Message) {
	asreq := msg.(*message.AssociationSetupRequest)
	log.Print(asreq)
}
