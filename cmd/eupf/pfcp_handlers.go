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

	hbres, err := message.NewHeartbeatResponse(hbreq.SequenceNumber, ie.NewRecoveryTimeStamp(time.Now())).Marshal()
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

// https://www.etsi.org/deliver/etsi_ts/129200_129299/129244/16.04.00_60/ts_129244v160400p.pdf page 95
func handlePfcpAssociationSetupRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {
	asreq := msg.(*message.AssociationSetupRequest)
	log.Printf("Got Association Setup Request from: %s. \n %s", addr, asreq)
	if asreq.NodeID == nil {
		log.Printf("Got Association Setup Request without NodeID from: %s", addr)
		return fmt.Errorf("association setup request without NodeID from: %s", addr)
	}
	// Get NodeID
	remote_nodeID, err := asreq.NodeID.NodeID()
	if err != nil {
		log.Printf("Got Association Setup Request with invalid NodeID from: %s", addr)
		return err
	}
	// Check if the PFCP Association Setup Request contains a Node ID for which a PFCP association was already established
	if _, ok := conn.nodeAssociations[remote_nodeID]; ok {
		log.Printf("Association Setup Request with NodeID: %s from: %s already exists", remote_nodeID, addr)
		// retain the PFCP sessions that were established with the existing PFCP association and that are requested to be retained, if the PFCP Session Retention Information IE was received in the request; otherwise, delete the PFCP sessions that were established with the existing PFCP association;
		log.Println("Session retention is not yet implemented")
	}

	// If the PFCP Association Setup Request contains a Node ID for which a PFCP association was already established
	// proceed with establishing the new PFCP association (regardless of the Recovery Timestamp received in the request), overwriting the existing association;
	// if the request is accepted:
	// shall store the Node ID of the CP function as the identifier of the PFCP association;
	// Create RemoteNode from AssociationSetupRequest
	remoteNode := NodeAssociation{
		ID:   remote_nodeID,
		Addr: addr.String(),
	}
	// Add or replace RemoteNode to NodeAssociationMap
	conn.nodeAssociations[remote_nodeID] = remoteNode
	log.Printf("Added RemoteNode: %s to NodeAssociationMap", remoteNode.ID)

	// shall send a PFCP Association Setup Response including:
	asres := message.NewAssociationSetupResponse(asreq.SequenceNumber,
		ie.NewCause(ie.CauseRequestAccepted), // a successful cause
		ie.NewNodeID("", "", conn.nodeId),    // its Node ID; Currently only support FQDN
		ie.NewUPFunctionFeatures(),           // information of all supported optional features in the UP function; We don't support any optional features at the moment
		// ... other IEs
		//	optionally one or more UE IP address Pool Information IE which contains a list of UE IP Address Pool Identities per Network Instance, S-NSSAI and IP version;
		//	optionally the NF Instance ID of the UPF if available
	)

	// Send AssociationSetupResponse
	response_bytes, err := asres.Marshal()
	if err != nil {
		log.Print(err)
		return err
	}
	if _, err := conn.Send(response_bytes, addr); err != nil {
		log.Print(err)
		return err
	}
	return nil
}
