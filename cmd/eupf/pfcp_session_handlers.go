package main

import (
	"fmt"
	"log"
	"net"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

var errMandatoryIeMissing = fmt.Errorf("mandatory IE missing")
var errNoEstablishedAssociation = fmt.Errorf("no established association")

func handlePfcpSessionEstablishmentRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {
	req := msg.(*message.SessionEstablishmentRequest)
	log.Printf("Got Session Establishment Request from: %s. \n %s", addr, req)
	remote_nodeID, fseid, err := sessionRelatedMessagesGuard(conn, addr, req.NodeID, req.CPFSEID)
	switch err {
	case errMandatoryIeMissing:
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		if err := conn.SendMessage(
			message.NewSessionEstablishmentResponse(0, 0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseMandatoryIEMissing)), addr); err != nil {
			log.Print(err)
			return err
		}
		SerReject.Inc()
		return nil
	case errNoEstablishedAssociation:
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		// Send SessionEstablishmentResponse with Cause: No PFCP Established Association
		if err := conn.SendMessage(message.NewSessionEstablishmentResponse(0, 0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), addr); err != nil {
			log.Print(err)
			return err
		}
		SerReject.Inc()
		return nil
	}

	// if session already exists, return error
	if _, ok := conn.nodeAssociations[remote_nodeID].Sessions[fseid.SEID]; ok {
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		if err := conn.SendMessage(
			message.NewSessionEstablishmentResponse(0, 0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseRequestRejected)), addr); err != nil {
			log.Print(err)
			return err
		}
		SerReject.Inc()
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
	SerSucsess.Inc()
	return nil
}

func validateNodeIdFSEID(nodeId *ie.IE, FSEID *ie.IE) error {
	if nodeId == nil || FSEID == nil {
		return fmt.Errorf("mandatory IE is missing")
	}
	// get remote node id
	_, err := nodeId.NodeID()
	if err != nil {
		return fmt.Errorf("NodeId is corrupted")
	}
	// get remote FSEID
	_, err = FSEID.FSEID()
	if err != nil {
		return fmt.Errorf("FSEID is corrupted")
	}
	return nil
}

func sessionRelatedMessagesGuard(conn *PfcpConnection, addr *net.UDPAddr, nodeId *ie.IE, cpfseid *ie.IE) (string, *ie.FSEIDFields, error) {
	if validateNodeIdFSEID(nodeId, cpfseid) != nil {
		return "", nil, errMandatoryIeMissing
	}
	// Errors checked in the validateNodeIdFSEID function
	remote_nodeID, _ := nodeId.NodeID()
	fseid, _ := cpfseid.FSEID()
	// Check if the PFCP Session Establishment Request contains a Node ID for which a PFCP association was already established
	if conn.checkNodeAssociation(remote_nodeID) != nil {
		return "", nil, errNoEstablishedAssociation
	}
	return remote_nodeID, fseid, nil
}
