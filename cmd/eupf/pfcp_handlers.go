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

func handlePfcpSessionEstablishmentRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {
	req := msg.(*message.SessionEstablishmentRequest)
	log.Printf("Got Session Establishment Request from: %s. \n %s", addr, req)
	// Check if the PFCP Session Establishment Request contains a Node ID for which a PFCP association was already established
	if req.NodeID == nil {
		log.Printf("Got Session Establishment Request without NodeID from: %s", addr)
		return fmt.Errorf("session establishment request without NodeID from: %s", addr)
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
		response_bytes, err := message.NewSessionEstablishmentResponse(0,
			0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation),
		).Marshal()
		if err != nil {
			log.Print(err)
			return err
		}
		if _, err := conn.Send(response_bytes, addr); err != nil {
			log.Print(err)
			return err
		}
	}
	if req.CPFSEID == nil {
		return fmt.Errorf("not found CP F-SEID")
	}
	fseid, err := req.CPFSEID.FSEID()
	if err != nil {
		return err
	}

	// We are using same SEID as SMF
	conn.nodeAssociations[remote_nodeID].Sessions[fseid.SEID] = Session{
		SEID: fseid.SEID,
	}

	// #TODO: Handle failed PDR applies
	

	var v4 net.IP
	addrv4, err := net.ResolveIPAddr("ip4", conn.nodeId)
	if err == nil {
		v4 = addrv4.IP.To4()
	}
	// #TODO: support v6
	var v6 net.IP
	// Send SessionEstablishmentResponse
	response_bytes, err := message.NewSessionEstablishmentResponse(
		0, 0,
		fseid.SEID,
		req.SequenceNumber,
		0,
		ie.NewCause(ie.CauseRequestAccepted),
		ie.NewNodeID("", "", conn.nodeId),		
		ie.NewFSEID(fseid.SEID, v4, v6),
	).Marshal()
	if err != nil {
		return err
	}
	if _, err := conn.Send(response_bytes, addr); err != nil {
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
