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

	hbres := message.NewHeartbeatResponse(hbreq.SequenceNumber, ie.NewRecoveryTimeStamp(time.Now()))
	if err := conn.SendMessage(hbres, addr); err != nil {
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
		newIeNodeID(conn.nodeId),             // its Node ID;
		ie.NewUPFunctionFeatures(),           // information of all supported optional features in the UP function; We don't support any optional features at the moment
		// ... other IEs
		//	optionally one or more UE IP address Pool Information IE which contains a list of UE IP Address Pool Identities per Network Instance, S-NSSAI and IP version;
		//	optionally the NF Instance ID of the UPF if available
	)

	// Send AssociationSetupResponse

	if err := conn.SendMessage(asres, addr); err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func handlePfcpSessionEstablishmentRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {
	req := msg.(*message.SessionEstablishmentRequest)
	log.Printf("Got Session Establishment Request from: %s. \n %s", addr, req)
	// Check if the PFCP Session Establishment Request contains a Node ID for which a PFCP association was already established
	if req.NodeID == nil || req.CPFSEID == nil {
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		// Send SessionEstablishmentResponse with Cause: Mandatory IE missing
		if err := conn.SendMessage(
			message.NewSessionEstablishmentResponse(0,
				0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseMandatoryIEMissing),
			), addr); err != nil {
			log.Print(err)
			return err
		}
		return fmt.Errorf("session establishment request without mandatory IEs NodeID(%+v), CPFSEID(%+v) from: %s", req.NodeID, req.CPFSEID, addr)
	}
	// Get NodeID
	remote_nodeID, err := req.NodeID.NodeID()
	if err != nil {
		log.Printf("Got Session Establishment Request with invalid NodeID from: %s", addr)
		return err
	}
	// Check if the PFCP Session Establishment Request contains a Node ID for which a PFCP association was already established
	if err := conn.checkNodeAssociation(remote_nodeID); err != nil {
		// shall reject any incoming PFCP Session related messages from that CP function, with a cause indicating that no PFCP association exists with the peer entity
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		// Send SessionEstablishmentResponse with Cause: No PFCP Established Association
		est_resp := message.NewSessionEstablishmentResponse(0,
			0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation),
		)
		if err := conn.SendMessage(est_resp, addr); err != nil {
			log.Print(err)
			return err
		}
	}

	fseid, err := req.CPFSEID.FSEID()
	if err != nil {
		return err
	}

	// if session already exists, return error
	if _, ok := conn.nodeAssociations[remote_nodeID].Sessions[fseid.SEID]; ok {
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		est_resp := message.NewSessionEstablishmentResponse(0,
			0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseRequestRejected),
		)
		if err := conn.SendMessage(est_resp, addr); err != nil {
			log.Print(err)
			return err
		}
	}
	// We are using same SEID as SMF
	conn.nodeAssociations[remote_nodeID].Sessions[fseid.SEID] = Session{
		SEID: fseid.SEID,
	}

	// #TODO: Actually applie rules to the dataplane
	// #TODO: Handle failed applies and return error

	// Print IE's content as is, it looks like there is no way to pretty print them, without implementing fortmatting for the whole go-pfcp library.
	for far := range req.CreateFAR {
		log.Printf("Create FAR: %+v", far)
	}

	for qer := range req.CreateQER {
		log.Printf("Create QER: %+v", qer)
	}

	for urr := range req.CreateURR {
		log.Printf("Create URR: %+v", urr)
	}

	for pdr := range req.CreatePDR {
		log.Printf("Create PDR: %+v", pdr)
	}

	if req.CreateBAR != nil {
		log.Printf("Create BAR: %+v", req.CreateBAR)
	}

	// #TODO: support v6
	var v6 net.IP
	// Send SessionEstablishmentResponse
	est_resp := message.NewSessionEstablishmentResponse(
		0, 0,
		fseid.SEID,
		req.SequenceNumber,
		0,
		ie.NewCause(ie.CauseRequestAccepted),
		newIeNodeID(conn.nodeId),
		ie.NewFSEID(fseid.SEID, conn.nodeAddrV4, v6),
	)
	if err := conn.SendMessage(est_resp, addr); err != nil {
		return err
	}
	return nil
}

// Check if for incoming message NodeID exists in NodeAssociationMap
func (conn *PfcpConnection) checkNodeAssociation(remote_nodeID string) error {
	if _, ok := conn.nodeAssociations[remote_nodeID]; !ok {
		log.Printf("NodeID: %s not found in NodeAssociationMap", remote_nodeID)
		return fmt.Errorf("nodeID: %s not found in NodeAssociationMap", remote_nodeID)
	}
	return nil
}

func newIeNodeID(nodeID string) *ie.IE {
	ip := net.ParseIP(nodeID)
	if ip != nil {
		if ip.To4() != nil {
			return ie.NewNodeID(nodeID, "", "")
		}
		return ie.NewNodeID("", nodeID, "")
	}
	return ie.NewNodeID("", "", nodeID)
}
