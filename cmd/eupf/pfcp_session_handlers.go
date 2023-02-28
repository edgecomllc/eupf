package main

import (
	"fmt"
	"log"
	"net"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func handlePfcpSessionEstablishmentRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {
	req := msg.(*message.SessionEstablishmentRequest)
	log.Printf("Got Session Establishment Request from: %s. \n %s", addr, req)
	remote_nodeID, fseid, err := sessionRelatedMessagesGuard(conn, addr, req.SequenceNumber, req.NodeID, req.CPFSEID)
	if err != nil {
		// Rejection message is already sent in sessionRelatedMessagesGuard
		return err
	}
	// if session already exists, return error
	if _, ok := conn.nodeAssociations[remote_nodeID].Sessions[fseid.SEID]; ok {
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		if err := conn.SendMessage(
			message.NewSessionEstablishmentResponse(0, 0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseRequestRejected)), addr); err != nil {
			log.Print(err)
			return err
		}
		return nil
	}
	// We are using same SEID as SMF
	conn.nodeAssociations[remote_nodeID].Sessions[fseid.SEID] = Session{
		SEID: fseid.SEID,
	}

	// #TODO: Actually apply rules to the dataplane
	// #TODO: Handle failed applies and return error
	printSessionEstablishmentRequest(req)

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

func validateNodeIdFSEID(nodeId *ie.IE, FSEID *ie.IE) error {
	if nodeId == nil || FSEID == nil {
		return fmt.Errorf("mandatory IE is missing")
	}
	// get remote node id
	_, err := nodeId.NodeID()
	if err != nil {
		return fmt.Errorf("NodeId is cirrupted")
	}
	// get remote FSEID
	_, err = FSEID.FSEID()
	if err != nil {
		return fmt.Errorf("FSEID is corrupted")
	}
	return nil
}

func sessionRelatedMessagesGuard(conn *PfcpConnection, addr *net.UDPAddr, seq uint32, nodeId *ie.IE, cpfseid *ie.IE) (string, *ie.FSEIDFields, error) {
	if validateNodeIdFSEID(nodeId, cpfseid) != nil {
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		if err := conn.SendMessage(
			message.NewSessionEstablishmentResponse(0, 0, 0, seq, 0, ie.NewCause(ie.CauseMandatoryIEMissing)), addr); err != nil {
			log.Print(err)
			return "", nil, err
		}
		return "", nil, fmt.Errorf("mandatory IE is missing")
	}
	// Errors checked in the validateNodeIdFSEID function
	remote_nodeID, _ := nodeId.NodeID()
	fseid, _ := cpfseid.FSEID()
	// Check if the PFCP Session Establishment Request contains a Node ID for which a PFCP association was already established
	if err := conn.checkNodeAssociation(remote_nodeID); err != nil {
		// shall reject any incoming PFCP Session related messages from that CP function, with a cause indicating that no PFCP association exists with the peer entity
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		// Send SessionEstablishmentResponse with Cause: No PFCP Established Association
		if err := conn.SendMessage(message.NewSessionEstablishmentResponse(0, 0, 0, seq, 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), addr); err != nil {
			log.Print(err)
			return "", nil, err
		}
		return "", nil, err
	}
	return remote_nodeID, fseid, nil
}
