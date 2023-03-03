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

func handlePfcpSessionEstablishmentRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) (message.Message, error) {
	req := msg.(*message.SessionEstablishmentRequest)
	log.Printf("Got Session Establishment Request from: %s. \n %s", addr, req)
	_, fseid, err := validateRequest(conn, addr, req.NodeID, req.CPFSEID)
	if err != nil {
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		SerReject.Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.SequenceNumber, 0, convertErrorToIeCause(err)), nil
	}

	association, ok := conn.nodeAssociations[addr.String()]
	if !ok {
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		SerReject.Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}

	// if session already exists, return error
	if _, ok := association.Sessions[fseid.SEID]; ok {
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		SerReject.Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseRequestRejected)), nil
	}
	// We are using same SEID as SMF
	association.Sessions[fseid.SEID] = Session{
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
	SerSuccess.Inc()
	return estResp, nil
}

func handlePfcpSessionDeletionRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) (message.Message, error) {
	req := msg.(*message.SessionDeletionRequest)
	log.Printf("Got Session Deletion Request from: %s. \n %s", addr, req)
	seid := req.SEID()
	association, ok := conn.nodeAssociations[addr.String()]
	if !ok {
		log.Printf("Rejecting Session Deletion Request from: %s", addr)
		SdrReject.Inc()
		return message.NewSessionDeletionResponse(0, 0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}
	// #TODO: Explore how Sessions should be stored, perform actual deletion of session when session storage API stabilizes
	_, ok = association.Sessions[seid]
	if !ok {
		log.Printf("Rejecting Session Deletion Request from: %s", addr)
		SdrReject.Inc()
		return message.NewSessionDeletionResponse(0, 0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseSessionContextNotFound)), nil
	}

	return message.NewSessionDeletionResponse(0, 0, seid, req.SequenceNumber, 0, ie.NewCause(ie.CauseRequestAccepted)), nil
}

func handlePfcpSessionModificationRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) (message.Message, error) {
	req := msg.(*message.SessionModificationRequest)
	log.Printf("Got Session Modification Request from: %s. \n %s", addr, req)
	_, fseid, err := validateRequest(conn, addr, req.NodeID, req.CPFSEID)
	if err != nil {
		log.Printf("Rejecting Session Modification Request from: %s", addr)
		SmrReject.Inc()
		return message.NewSessionModificationResponse(0, 0, 0, req.SequenceNumber, 0, convertErrorToIeCause(err)), nil
	}

	association, ok := conn.nodeAssociations[addr.String()]
	if !ok {
		log.Printf("Rejecting Session Modification Request from: %s", addr)
		SmrReject.Inc()
		return message.NewSessionModificationResponse(0, 0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}

	_, ok = association.Sessions[fseid.SEID]
	if !ok {
		log.Printf("Rejecting Session Modification Request from: %s", addr)
		SmrReject.Inc()
		return message.NewSessionModificationResponse(0, 0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseSessionContextNotFound)), nil
	}

	// #TODO: Actually apply rules to the dataplane
	// #TODO: Handle failed applies and return error
	printSessionModificationRequest(req)

	// #TODO: support v6
	var v6 net.IP
	// Send SessionEstablishmentResponse
	modResp := message.NewSessionModificationResponse(
		0, 0,
		fseid.SEID,
		req.SequenceNumber,
		0,
		ie.NewCause(ie.CauseRequestAccepted),
		newIeNodeID(conn.nodeId),
		ie.NewFSEID(fseid.SEID, conn.nodeAddrV4, v6),
	)
	SmrSuccess.Inc()
	return modResp, nil
}

func convertErrorToIeCause(err error) *ie.IE {
	switch err {
	case errMandatoryIeMissing:
		return ie.NewCause(ie.CauseMandatoryIEMissing)
	case errNoEstablishedAssociation:
		return ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)
	default:
		log.Printf("Unknown error: %s", err)
		return ie.NewCause(ie.CauseRequestRejected)
	}
}

func validateRequest(conn *PfcpConnection, addr *net.UDPAddr, nodeId *ie.IE, cpfseid *ie.IE) (string, *ie.FSEIDFields, error) {
	if nodeId == nil || cpfseid == nil {
		return "", nil, errMandatoryIeMissing
	}
	_, err := nodeId.NodeID()
	if err != nil {
		return "", nil, errMandatoryIeMissing
	}
	_, err = cpfseid.FSEID()
	if err != nil {
		return "", nil, errMandatoryIeMissing
	}

	remoteNodeID, _ := nodeId.NodeID()
	fseid, _ := cpfseid.FSEID()
	return remoteNodeID, fseid, nil
}
