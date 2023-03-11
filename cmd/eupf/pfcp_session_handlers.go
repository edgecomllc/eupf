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

	session := Session{
		LocalSEID:  localSEID,
		RemoteSEID: remoneSEID.SEID,
		updrs:      map[uint32]PdrInfo{},
		dpdrs:      map[string]PdrInfo{},
		fars:       map[uint32]FarInfo{},
	}

	// #TODO: Actually apply rules to the dataplane
	// #TODO: Handle failed applies and return error
	printSessionEstablishmentRequest(req)

	bpfObjects := conn.bpfObjects
	for _, far := range req.CreateFAR {
		farInfo := FarInfo{}
		if applyAction, err := far.ApplyAction(); err == nil {
			farInfo.Action = applyAction[0]
		}
		if outerHeaderCreation, err := far.OuterHeaderCreation(); err == nil {
			farInfo.OuterHeaderCreation = 1 // FIXME
			farInfo.Teid = outerHeaderCreation.TEID
		}
		// SRC IP ???

		farid, _ := far.FARID()
		session.CreateFAR(bpfObjects, farid, farInfo)
	}

	for _, pdr := range req.CreatePDR {
		pdrInfo := PdrInfo{}
		if outerHeaderRemoval, err := pdr.OuterHeaderRemoval(); err == nil {
			pdrInfo.OuterHeaderRemoval = outerHeaderRemoval[0] // FIXME
		}
		if farid, err := pdr.FARID(); err == nil {
			pdrInfo.FarId = farid
		}
		pdi, err := pdr.PDI()
		if err != nil {
			log.Print(err)
			return nil, err
		}
		srcInterface, _ := pdi[0].SourceInterface()
		if srcInterface == ie.SrcInterfaceAccess {
			fteid, _ := pdi[0].FTEID()
			teid := fteid.TEID
			session.CreateUpLinkPDR(bpfObjects, teid, pdrInfo)
		} else {
			ue_ip, _ := pdi[0].UEIPAddress()
			ipv4 := ue_ip.IPv4Address
			session.CreateDownLinkPDR(bpfObjects, ipv4, pdrInfo)
		}
	}

	// Reassigning is the best I can think of for now
	association.Sessions[localSEID] = session
	// FIXME
	conn.nodeAssociations[addr.String()] = association

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

	delete(association.Sessions, req.SEID())

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

	bpfObjects := conn.bpfObjects
	for _, far := range req.UpdateFAR {
		farInfo := FarInfo{}
		if applyAction, err := far.ApplyAction(); err == nil {
			farInfo.Action = applyAction[0]
		}
		if outerHeaderCreation, err := far.OuterHeaderCreation(); err == nil {
			farInfo.OuterHeaderCreation = 1 // FIXME
			farInfo.Teid = outerHeaderCreation.TEID
		}
		// SRC IP ???

		farid, _ := far.FARID()
		session.UpdateFAR(bpfObjects, farid, farInfo)
	}

	for _, pdr := range req.UpdatePDR {
		pdrInfo := PdrInfo{}
		if outerHeaderRemoval, err := pdr.OuterHeaderRemoval(); err == nil {
			pdrInfo.OuterHeaderRemoval = outerHeaderRemoval[0] // FIXME
		}
		if farid, err := pdr.FARID(); err == nil {
			pdrInfo.FarId = farid
		}
		pdi, err := pdr.PDI()
		if err != nil {
			log.Print(err)
			return nil, err
		}
		srcInterface, _ := pdi[0].SourceInterface()
		if srcInterface == ie.SrcInterfaceAccess {
			fteid, _ := pdi[0].FTEID()
			teid := fteid.TEID
			session.UpdateUpLinkPDR(bpfObjects, teid, pdrInfo)
		} else {
			ue_ip, _ := pdi[0].UEIPAddress()
			ipv4 := ue_ip.IPv4Address
			session.UpdateDownLinkPDR(bpfObjects, ipv4, pdrInfo)
		}
	}

	association.Sessions[req.SEID()] = session

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
