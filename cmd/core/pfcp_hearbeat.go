package core

import (
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
	"log"
	"net"
	"time"
)

func HandlePfcpHeartbeatRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	hbreq := msg.(*message.HeartbeatRequest)
	if association := conn.GetAssociation(addr); association != nil {
		association.CheckInContact()
	}
	ts, err := hbreq.RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		log.Printf("Got Heartbeat Request with invalid TS: %s, from: %s", err, addr)
		return nil, err
	} else {
		log.Printf("Got Heartbeat Request with TS: %s, from: %s", ts, addr)
	}

	hbres := message.NewHeartbeatResponse(hbreq.SequenceNumber, ie.NewRecoveryTimeStamp(time.Now()))
	log.Printf("Sent Heartbeat Response to: %s", addr)
	return hbres, nil
}

func HandlePfcpHeartbeatResponse(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	hbresp := msg.(*message.HeartbeatResponse)
	ts, err := hbresp.RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		log.Printf("Got Heartbeat Response with invalid TS: %s, from: %s", err, addr)
	} else {
		log.Printf("Got Heartbeat Response with TS: %s, from: %s", ts, addr)
	}
	if association := conn.GetAssociation(addr); association != nil {
		association.CheckInContact()
	}
	return nil, err
}

func SendHearbeatReqeust(conn *PfcpConnection, addr string) {
	hbreq := message.NewHeartbeatRequest(0, ie.NewRecoveryTimeStamp(time.Now()), nil)
	log.Printf("Sent Heartbeat Request to: %s", addr)
	// create net.UDPAddr from string and port from config
	udpAddr, err := net.ResolveUDPAddr("udp", addr+":8805")
	if err == nil {
		if err := conn.SendMessage(hbreq, udpAddr); err != nil {
			log.Printf("Failed to send heartbeat request: %s\n", err.Error())
		}
	} else {
		log.Printf("Failed to send Hearbit Request: %s\n", err.Error())
	}
}
