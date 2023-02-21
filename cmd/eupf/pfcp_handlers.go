package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PfcpFunc func(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error

type PfcpHanderMap map[uint8]PfcpFunc

func (handlerMap PfcpHanderMap) Handle(conn *PfcpConnection, buf []byte, addr *net.UDPAddr) error {
	log.Printf("Handling PFCP message from %s", addr)
	msg, err := message.Parse(buf)
	if err != nil {
		log.Printf("Ignored undecodable message: %x, error: %s", buf, err)
		return err
	}
	log.Printf("Parsed PFCP message: %s", msg)
	if handler, ok := handlerMap[msg.MessageType()]; ok {
		err := handler(conn, msg, addr)
		if err != nil {
			log.Printf("Error handling PFCP message: %s", err)
			return err
		}
	} else {
		log.Printf("Got unexpected message %s: %s, from: %s", msg.MessageTypeName(), msg, addr)
	}
	return nil
}

func handlePfcpHeartbeatRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {
	hbreq := msg.(*message.HeartbeatRequest)
	ts, err := hbreq.RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		log.Printf("Got Heartbeat Request with invalid TS: %s, from: %s", err, addr)
		return err
	} else {
		log.Printf("Got Heartbeat Request with TS: %s, from: %s", ts, addr)
	}

	// #TODO: Explore how to properly set sequence number
	// Answer with same Sequence Number as in request
	var seq uint32 = 1
	hbres, err := message.NewHeartbeatResponse(seq, ie.NewRecoveryTimeStamp(time.Now())).Marshal()
	if err != nil {
		log.Print(err)
		return err
	}

	if _, err := conn.Send(hbres, addr); err != nil {
		log.Print(err)
		return err
	}
	log.Printf("Sent Heartbeat Response to: %s", addr)
	return nil
}

func handlePfcpAssociationSetupRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {
	asreq := msg.(*message.AssociationSetupRequest)
	log.Printf("Got Association Setup Request from: %s. \n %s", addr, asreq)
	if asreq.NodeID == nil {
		log.Printf("Got Association Setup Request without NodeID from: %s", addr)
		return fmt.Errorf("association setup request without NodeID from: %s", addr)
	}
	// Get NodeID
	nodeID, err := asreq.NodeID.NodeID()
	if err != nil {
		log.Printf("Got Association Setup Request with invalid NodeID from: %s", addr)
		return err
	}
	// Create RemoteNode from AssociationSetupRequest
	remoteNode := RemoteNode{
		ID:   nodeID,
		Addr: addr.String(),
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
	if _, err := conn.Send(asres, addr); err != nil {
		log.Print(err)
		return err
	}
	return nil
}

// Handle PFCP Association Release Request
func handlePfcpAssociationReleaseRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {
	release_request := msg.(*message.AssociationReleaseRequest)
	log.Printf("Got Association Release Request from: %s. \n %s", addr, release_request)
	if release_request.NodeID == nil {
		log.Printf("Got Association Release Request without NodeID from: %s", addr)
		return fmt.Errorf("association release request without NodeID from: %s", addr)
	}
	// #TODO: ...

	return nil

}
