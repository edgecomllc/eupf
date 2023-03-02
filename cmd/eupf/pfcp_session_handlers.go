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
	remoteNodeID, fseid, err := validateRequest(conn, addr, req.NodeID, req.CPFSEID)
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
	if _, ok := conn.nodeAssociations[remoteNodeID].Sessions[fseid.SEID]; ok {
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
	conn.nodeAssociations[remoteNodeID].Sessions[fseid.SEID] = Session{
		SEID: fseid.SEID,
	}

	// #TODO: Actually apply rules to the dataplane
	// #TODO: Handle failed applies and return error
	printSessionEstablishmentRequest(req)

	// #TODO: support v6
	var v6 net.IP
	// Send SessionEstablishmentResponse
	estResp := message.NewSessionEstablishmentResponse(
		0, 0,
		fseid.SEID,
		req.SequenceNumber,
		0,
		ie.NewCause(ie.CauseRequestAccepted),
		newIeNodeID(conn.nodeId),
		ie.NewFSEID(fseid.SEID, conn.nodeAddrV4, v6),
	)
	if err := conn.SendMessage(estResp, addr); err != nil {
		return err
	}
	SerSuccess.Inc()
	return nil
}

func validateRequest(conn *PfcpConnection, addr *net.UDPAddr, nodeId *ie.IE, cpfseid *ie.IE) (string, *ie.FSEIDFields, error) {
	if nodeId == nil || cpfseid == nil {
		return "", nil, fmt.Errorf("mandatory IE is missing")
	}
	_, err := nodeId.NodeID()
	if err != nil {
		return "", nil, fmt.Errorf("NodeId is corrupted")
	}
	_, err = cpfseid.FSEID()
	if err != nil {
		return "", nil, fmt.Errorf("FSEID is corrupted")
	}
	
	remoteNodeID, _ := nodeId.NodeID()
	fseid, _ := cpfseid.FSEID()
	// Check if the PFCP Session Establishment Request contains a Node ID for which a PFCP association was already established
	if conn.checkNodeAssociation(remoteNodeID) != nil {
		return "", nil, errNoEstablishedAssociation
	}
	return remoteNodeID, fseid, nil
}
