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
	_, remoneSEID, err := validateRequest(conn, addr, req.NodeID, req.CPFSEID)
	if err != nil {
		log.Printf("Rejecting Session Establishment Request from: %s (missing NodeID or F-SEID)", addr)
		SerReject.Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, convertErrorToIeCause(err)), nil
	}

	association, ok := conn.nodeAssociations[addr.String()]
	if !ok {
		log.Printf("Rejecting Session Establishment Request from: %s (no association)", addr)
		SerReject.Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}

	localSEID := association.NewLocalSEID()
	// // if session already exists, return error
	// if _, ok := association.Sessions[fseid.SEID]; ok {
	// 	log.Printf("Rejecting Session Establishment Request from: %s (unknown SEID)", addr)
	// 	SerReject.Inc()
	// 	return message.NewSessionEstablishmentResponse(0, 0, fseid.SEID, req.Sequence(), 0, ie.NewCause(ie.CauseRequestRejected)), nil
	// }

	association.Sessions[localSEID] = Session{
		LocalSEID:  localSEID,
		RemoteSEID: remoneSEID.SEID,
	}

	// FIXME
	conn.nodeAssociations[addr.String()] = association

	// #TODO: Actually apply rules to the dataplane
	// #TODO: Handle failed applies and return error
	printSessionEstablishmentRequest(req)

	// #TODO: support v6
	var v6 net.IP
	// Send SessionEstablishmentResponse
	estResp := message.NewSessionEstablishmentResponse(
		0, 0,
		remoneSEID.SEID,
		req.Sequence(),
		0,
		ie.NewCause(ie.CauseRequestAccepted),
		newIeNodeID(conn.nodeId),
		ie.NewFSEID(localSEID, conn.nodeAddrV4, v6),
	)
	SerSuccess.Inc()
	return estResp, nil
}

func handlePfcpSessionDeletionRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) (message.Message, error) {
	req := msg.(*message.SessionDeletionRequest)
	log.Printf("Got Session Deletion Request from: %s. \n %s", addr, req)
	association, ok := conn.nodeAssociations[addr.String()]
	if !ok {
		log.Printf("Rejecting Session Deletion Request from: %s (no association)", addr)
		SdrReject.Inc()
		return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}
	// #TODO: Explore how Sessions should be stored, perform actual deletion of session when session storage API stabilizes
	session, ok := association.Sessions[req.SEID()]
	if !ok {
		log.Printf("Rejecting Session Deletion Request from: %s (unknown SEID)", addr)
		SdrReject.Inc()
		return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseSessionContextNotFound)), nil
	}

	SdrSuccess.Inc()
	return message.NewSessionDeletionResponse(0, 0, session.RemoteSEID, req.Sequence(), 0, ie.NewCause(ie.CauseRequestAccepted)), nil
}

func handlePfcpSessionModificationRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) (message.Message, error) {
	req := msg.(*message.SessionModificationRequest)
	log.Printf("Got Session Modification Request from: %s. \n %s", addr, req)

	association, ok := conn.nodeAssociations[addr.String()]
	if !ok {
		log.Printf("Rejecting Session Modification Request from: %s (no association)", addr)
		SmrReject.Inc()
		return message.NewSessionModificationResponse(0, 0, req.SEID(), req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}

	session, ok := association.Sessions[req.SEID()]
	if !ok {
		log.Printf("Rejecting Session Modification Request from: %s (unknown SEID)", addr)
		SmrReject.Inc()
		return message.NewSessionModificationResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseSessionContextNotFound)), nil
	}

	// This IE shall be present if the CP function decides to change its F-SEID for the PFCP session. The UP function
	// shall use the new CP F-SEID for subsequent PFCP Session related messages for this PFCP Session
	if req.CPFSEID != nil {
		remoteSEID, err := req.CPFSEID.FSEID()
		if err == nil {
			session.RemoteSEID = remoteSEID.SEID

			association.Sessions[req.SEID()] = session         // FIXME
			conn.nodeAssociations[addr.String()] = association // FIXME
		}
	}

	// #TODO: Actually apply rules to the dataplane
	// #TODO: Handle failed applies and return error
	printSessionModificationRequest(req)

	// Send SessionEstablishmentResponse
	modResp := message.NewSessionModificationResponse(
		0, 0,
		session.RemoteSEID,
		req.Sequence(),
		0,
		ie.NewCause(ie.CauseRequestAccepted),
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
