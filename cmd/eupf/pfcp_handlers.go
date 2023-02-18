package main

import (
	"log"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PfcpFunc func(conn *PfcpConnection, msg message.Message) error

type PfcpHanderMap map[uint8]PfcpFunc

func (h PfcpHanderMap) Handle(c *PfcpConnection, buf []byte) error {
	log.Printf("Handling PFCP message from %s", c.RemoteAddr())
	msg, err := message.Parse(buf)
	if err != nil {
		log.Printf("Ignored undecodable message: %x, error: %s", buf, err)
		return err
	}
	log.Printf("Parsed PFCP message: %s", msg)
	if handler, ok := h[msg.MessageType()]; ok {
		err := handler(c, msg)
		if err != nil {
			log.Printf("Error handling PFCP message: %s", err)
			return err
		}
	} else {
		log.Printf("Got unexpected message %s: %s, from: %s", msg.MessageTypeName(), msg, c.RemoteAddr())
	}
	return nil
}

func handlePfcpHeartbeatRequest(c *PfcpConnection, msg message.Message) error {
	hbreq := msg.(*message.HeartbeatRequest)
	ts, err := hbreq.RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		log.Printf("Got Heartbeat Request with invalid TS: %s, from: %s", err, c.RemoteAddr())
		return err
	} else {
		log.Printf("Got Heartbeat Request with TS: %s, from: %s", ts, c.RemoteAddr())
	}

	// #TODO: add sequence tracking for individual sessions
	var seq uint32 = 1
	hbres, err := message.NewHeartbeatResponse(seq, ie.NewRecoveryTimeStamp(time.Now())).Marshal()
	if err != nil {
		log.Print(err)
		return err
	}

	if _, err := c.Send(hbres, c.RemoteAddr()); err != nil {
		log.Print(err)
		return err
	}
	log.Printf("Sent Heartbeat Response to: %s", c.RemoteAddr())
	return nil
}

func handlePfcpAssociationSetupRequest(c *PfcpConnection, msg message.Message) error {
	asreq := msg.(*message.AssociationSetupRequest)
	log.Print(asreq)
	log.Printf("Got Association Setup Request from: %s", c.RemoteAddr())
	return nil
}
