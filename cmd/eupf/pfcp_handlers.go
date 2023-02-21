package main

import (
	"fmt"
	"log"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PfcpFunc func(conn *PfcpConnection, msg message.Message) error

type PfcpHanderMap map[uint8]PfcpFunc

func (handlerMap PfcpHanderMap) Handle(conn *PfcpConnection, buf []byte) error {
	log.Printf("Handling PFCP message from %s", conn.RemoteAddr())
	msg, err := message.Parse(buf)
	if err != nil {
		log.Printf("Ignored undecodable message: %x, error: %s", buf, err)
		return err
	}
	log.Printf("Parsed PFCP message: %s", msg)
	if handler, ok := handlerMap[msg.MessageType()]; ok {
		err := handler(conn, msg)
		if err != nil {
			log.Printf("Error handling PFCP message: %s", err)
			return err
		}
	} else {
		log.Printf("Got unexpected message %s: %s, from: %s", msg.MessageTypeName(), msg, conn.RemoteAddr())
	}
	return nil
}

func handlePfcpHeartbeatRequest(conn *PfcpConnection, msg message.Message) error {
	hbreq := msg.(*message.HeartbeatRequest)
	ts, err := hbreq.RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		log.Printf("Got Heartbeat Request with invalid TS: %s, from: %s", err, conn.RemoteAddr())
		return err
	} else {
		log.Printf("Got Heartbeat Request with TS: %s, from: %s", ts, conn.RemoteAddr())
	}

	// #TODO: Explore how to properly set sequence number
	var seq uint32 = 1
	hbres, err := message.NewHeartbeatResponse(seq, ie.NewRecoveryTimeStamp(time.Now())).Marshal()
	if err != nil {
		log.Print(err)
		return err
	}

	if _, err := conn.Send(hbres); err != nil {
		log.Print(err)
		return err
	}
	log.Printf("Sent Heartbeat Response to: %s", conn.RemoteAddr())
	return nil
}

func handlePfcpAssociationSetupRequest(conn *PfcpConnection, msg message.Message) error {
	asreq := msg.(*message.AssociationSetupRequest)
	log.Printf("Got Association Setup Request from: %s. \n %s", conn.RemoteAddr(), asreq)
	if asreq.NodeID == nil {
		log.Printf("Got Association Setup Request without NodeID from: %s", conn.RemoteAddr())
		return fmt.Errorf("association setup request without NodeID from: %s", conn.RemoteAddr())
	}
	// Get NodeID
	nodeID, err := asreq.NodeID.NodeID()
	if err != nil {
		log.Printf("Got Association Setup Request with invalid NodeID from: %s", conn.RemoteAddr())
		return err
	}
	// Create RemoteNode from AssociationSetupRequest
	remoteNode := RemoteNode{
		ID:   nodeID,
		Addr: conn.RemoteAddr().String(),
	}
	// Add or replace RemoteNode to NodeAssociationMap
	conn.nodeAssociations[nodeID] = remoteNode
	log.Printf("Added RemoteNode: %s to NodeAssociationMap", remoteNode)
	// Create AssociationSetupResponse
	// #TODO: Explore how to properly set sequence number
	var seq uint32 = 1
	asres, err := message.NewAssociationSetupResponse(seq,
		ie.NewRecoveryTimeStamp(time.Now()),
		ie.NewNodeID(nodeID, "", ""),
		ie.NewCause(ie.CauseRequestAccepted),
		// ... other IEs
	).Marshal()
	if err != nil {
		log.Print(err)
		return err
	}
	// Send AssociationSetupResponse
	if _, err := conn.Send(asres); err != nil {
		log.Print(err)
		return err
	}
	return nil
}

// Handle PFCP Association Release Request
func handlePfcpAssociationReleaseRequest(conn *PfcpConnection, msg message.Message) error {
	release_request := msg.(*message.AssociationReleaseRequest)
	log.Printf("Got Association Release Request from: %s. \n %s", conn.RemoteAddr(), release_request)
	if release_request.NodeID == nil {
		log.Printf("Got Association Release Request without NodeID from: %s", conn.RemoteAddr())
		return fmt.Errorf("association release request without NodeID from: %s", conn.RemoteAddr())
	}
	// #TODO: ...

	return nil
	

}