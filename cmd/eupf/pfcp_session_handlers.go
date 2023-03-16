package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
	"golang.org/x/exp/slices"
)

var errMandatoryIeMissing = fmt.Errorf("mandatory IE missing")
var errNoEstablishedAssociation = fmt.Errorf("no established association")

func handlePfcpSessionEstablishmentRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) (message.Message, error) {
	req := msg.(*message.SessionEstablishmentRequest)
	log.Printf("Got Session Establishment Request from: %s. \n %s", addr, req)
	_, remoteSEID, err := validateRequest(conn, addr, req.NodeID, req.CPFSEID)
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
		LocalSEID:    localSEID,
		RemoteSEID:   remoteSEID.SEID,
		UplinkPDRs:   map[uint32]SPDRInfo{},
		DownlinkPDRs: map[uint32]SPDRInfo{},
		FARs:         map[uint32]FarInfo{},
	}

	printSessionEstablishmentRequest(req)
	// #TODO: Implement rollback on error
	err = func() error {
		bpfObjects := conn.bpfObjects
		for _, far := range req.CreateFAR {
			// #TODO: Extract to standalone function to avoid code duplication
			farInfo := FarInfo{}
			if applyAction, err := far.ApplyAction(); err == nil {
				farInfo.Action = applyAction[0]
			}
			if forward, err := far.ForwardingParameters(); err == nil {
				outerHeaderCreationIndex := findIEindex(forward, 84) // IE Type Outer Header Creation
				if outerHeaderCreationIndex == -1 {
					log.Println("WARN: No OuterHeaderCreation")
				} else {
					outerHeaderCreation, _ := forward[outerHeaderCreationIndex].OuterHeaderCreation()
					farInfo.OuterHeaderCreation = 1
					farInfo.Teid = outerHeaderCreation.TEID
					farInfo.Srcip = ip2int(outerHeaderCreation.IPv4Address)
				}
			}

			farid, _ := far.FARID()
			session.CreateFAR(bpfObjects, farid, farInfo)
		}

		//#TODO: Extract to standalone function to avoid code duplication
		for _, pdr := range req.CreatePDR {
			spdrInfo := SPDRInfo{}
			pdrId, err := pdr.PDRID()
			if err != nil {
				return fmt.Errorf("PDR ID missing")
			}
			if outerHeaderRemoval, err := pdr.OuterHeaderRemovalDescription(); err == nil {
				spdrInfo.PdrInfo.OuterHeaderRemoval = outerHeaderRemoval
			}
			if farid, err := pdr.FARID(); err == nil {
				spdrInfo.PdrInfo.FarId = uint16(farid)
			}
			pdi, err := pdr.PDI()
			if err != nil {
				log.Print(err)
				return err
			}
			srcIfacePdiId := findIEindex(pdi, 20) // IE Type source interface
			srcInterface, _ := pdi[srcIfacePdiId].SourceInterface()
			// #TODO: Rework Uplink/Downlink decesion making
			if srcInterface == ie.SrcInterfaceAccess {
				teidPdiId := findIEindex(pdi, 21) // IE Type F-TEID

				if teidPdiId == -1 {
					log.Println("F-TEID IE missing")
					return fmt.Errorf("F-TEID IE missing")
				}
				if fteid, err := pdi[teidPdiId].FTEID(); err == nil {
					spdrInfo.Teid = fteid.TEID
					session.CreateUpLinkPDR(bpfObjects, pdrId, spdrInfo)
				} else {
					log.Println(err)
					return err
				}
			} else {
				ueipPdiId := findIEindex(pdi, 93) // IE Type UE IP Address
				if ueipPdiId == -1 {
					log.Println("UE IP Address IE missing")
					return fmt.Errorf("UE IP Address IE missing")
				}
				ue_ip, _ := pdi[ueipPdiId].UEIPAddress()
				spdrInfo.Ipv4 = ue_ip.IPv4Address
				session.CreateDownLinkPDR(bpfObjects, pdrId, spdrInfo)
			}
		}
		return nil
	}()

	if err != nil {
		log.Printf("Rejecting Session Establishment Request from: %s (error in applying IEs)", err)
		SerReject.Inc()
		return message.NewSessionEstablishmentResponse(0, 0, remoteSEID.SEID, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), nil
	}

	// #TODO: Add cleanup if some of IEs cannot be applied

	// Reassigning is the best I can think of for now
	association.Sessions[localSEID] = session
	// FIXME
	conn.nodeAssociations[addr.String()] = association

	// #TODO: support v6
	var v6 net.IP
	// Send SessionEstablishmentResponse
	estResp := message.NewSessionEstablishmentResponse(
		0, 0,
		remoteSEID.SEID,
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

	if err := session.Cleanup(conn.bpfObjects); err != nil {
		log.Printf("Rejecting Session Deletion Request from: %s (cleanup failed)", addr)
		return nil, err // FIXME
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

	printSessionModificationRequest(req)

	// #TODO: Implement rollback on error
	err := func() error {
		bpfObjects := conn.bpfObjects
		// #TODO: Extract to standalone function to avoid code duplication
		for _, far := range req.UpdateFAR {
			farInfo := FarInfo{}
			if applyAction, err := far.ApplyAction(); err == nil {
				farInfo.Action = applyAction[0]
			}
			if forward, err := far.UpdateForwardingParameters(); err == nil {
				outerHeaderCreationIndex := findIEindex(forward, 84) // IE Type Outer Header Creation
				if outerHeaderCreationIndex == -1 {
					log.Println("WARN: No OuterHeaderCreation")
				} else {
					outerHeaderCreation, _ := forward[outerHeaderCreationIndex].OuterHeaderCreation()
					farInfo.OuterHeaderCreation = 1
					farInfo.Teid = outerHeaderCreation.TEID
					farInfo.Srcip = ip2int(outerHeaderCreation.IPv4Address)
				}
			} else {
				log.Println("WARN: No UpdateForwardingParameters")
			}

			farid, _ := far.FARID()
			session.UpdateFAR(bpfObjects, farid, farInfo)
		}

		for _, removeFar := range req.RemoveFAR {
			farid, _ := removeFar.FARID()
			session.RemoveFAR(bpfObjects, farid)
		}

		for _, pdr := range req.RemovePDR {
			pdrId, _ := pdr.PDRID()
			session.RemovePDR(bpfObjects, pdrId)
		}

		// #TODO: Extract to standalone function to avoid code duplication
		for _, pdr := range req.UpdatePDR {
			spdrInfo := SPDRInfo{}
			pdrId, err := pdr.PDRID()
			if err != nil {
				return fmt.Errorf("PDR ID missing")
			}
			if outerHeaderRemoval, err := pdr.OuterHeaderRemovalDescription(); err == nil {
				spdrInfo.PdrInfo.OuterHeaderRemoval = outerHeaderRemoval
			}
			if farid, err := pdr.FARID(); err == nil {
				spdrInfo.PdrInfo.FarId = uint16(farid)
			}
			pdi, err := pdr.PDI()
			if err != nil {
				log.Print(err)
				return err
			}
			srcIfacePdiId := findIEindex(pdi, 20) // IE Type source interface
			srcInterface, _ := pdi[srcIfacePdiId].SourceInterface()
			// #TODO: Rework Uplink/Downlink decesion making
			if srcInterface == ie.SrcInterfaceAccess {
				teidPdiId := findIEindex(pdi, 21) // IE Type F-TEID

				if teidPdiId == -1 {
					log.Println("F-TEID IE missing")
					return fmt.Errorf("F-TEID IE missing")
				}
				if fteid, err := pdi[teidPdiId].FTEID(); err == nil {
					spdrInfo.Teid = fteid.TEID
					session.UpdateUpLinkPDR(bpfObjects, pdrId, spdrInfo)
				} else {
					log.Println(err)
					return err
				}
			} else {
				ueipPdiId := findIEindex(pdi, 93) // IE Type UE IP Address
				if ueipPdiId == -1 {
					log.Println("UE IP Address IE missing")
					return fmt.Errorf("UE IP Address IE missing")
				}
				ue_ip, _ := pdi[ueipPdiId].UEIPAddress()
				spdrInfo.Ipv4 = ue_ip.IPv4Address
				session.UpdateDownLinkPDR(bpfObjects, pdrId, spdrInfo)
			}
		}
		return nil
	}()
	if err != nil {
		log.Printf("Rejecting Session Modification Request from: %s (failed to apply rules)", err)
		SmrReject.Inc()
		return message.NewSessionModificationResponse(0, 0, session.RemoteSEID, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), nil

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

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		panic("no sane way to convert ipv6 into uint32")
	}
	return binary.BigEndian.Uint32(ip.To4())
}

func findIEindex(ieArr []*ie.IE, ieType uint16) int {
	arrIndex := slices.IndexFunc(ieArr, func(ie *ie.IE) bool {
		return ie.Type == ieType
	})
	return arrIndex
}
