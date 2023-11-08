package core

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/edgecomllc/eupf/cmd/core/service"

	"github.com/edgecomllc/eupf/cmd/ebpf"

	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
	"golang.org/x/exp/slices"
)

var errMandatoryIeMissing = fmt.Errorf("mandatory IE missing")
var errNoEstablishedAssociation = fmt.Errorf("no established association")

func HandlePfcpSessionEstablishmentRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	req := msg.(*message.SessionEstablishmentRequest)
	log.Info().Msgf("Got Session Establishment Request from: %s.", addr)
	remoteSEID, err := validateRequest(req.NodeID, req.CPFSEID)
	if err != nil {
		log.Info().Msgf("Rejecting Session Establishment Request from: %s (missing NodeID or F-SEID)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEMissing)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, convertErrorToIeCause(err)), nil
	}

	association, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Info().Msgf("Rejecting Session Establishment Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}

	localSEID := association.NewLocalSEID()

	session := NewSession(localSEID, remoteSEID.SEID)

	printSessionEstablishmentRequest(req)
	// #TODO: Implement rollback on error
	createdPDRs := []SPDRInfo{}

	err = func() error {
		mapOperations := conn.mapOperations
		for _, far := range req.CreateFAR {
			farInfo, err := composeFarInfo(far, conn.n3Address.To4(), ebpf.FarInfo{})
			if err != nil {
				log.Info().Msgf("Error extracting FAR info: %s", err.Error())
				continue
			}

			farid, _ := far.FARID()
			log.Info().Msgf("Saving FAR info to session: %d, %+v", farid, farInfo)
			if internalId, err := mapOperations.NewFar(farInfo); err == nil {
				session.NewFar(farid, internalId, farInfo)
			} else {
				log.Info().Msgf("Can't put FAR: %s", err.Error())
				return err
			}
		}

		for _, qer := range req.CreateQER {
			qerInfo := ebpf.QerInfo{}
			qerId, err := qer.QERID()
			if err != nil {
				return fmt.Errorf("QER ID missing")
			}
			updateQer(&qerInfo, qer)
			log.Info().Msgf("Saving QER info to session: %d, %+v", qerId, qerInfo)
			if internalId, err := mapOperations.NewQer(qerInfo); err == nil {
				session.NewQer(qerId, internalId, qerInfo)
			} else {
				log.Info().Msgf("Can't put QER: %s", err.Error())
				return err
			}
		}

		// TODO: create local_teid_map

		for _, pdr := range req.CreatePDR {
			// PDR should be created last, because we need to reference FARs and QERs global id
			pdrId, err := pdr.PDRID()
			if err != nil {
				log.Info().Msgf("PDR ID missing")
				continue
			}

			spdrInfo := SPDRInfo{PdrID: uint32(pdrId)}
			if err := extractPDR(pdr, session, &spdrInfo, conn.ipam, req.SEID()); err == nil {
				session.PutPDR(spdrInfo.PdrID, spdrInfo)
				applyPDR(spdrInfo, mapOperations)
				createdPDRs = append(createdPDRs, spdrInfo)
			} else {
				log.Printf("Error extracting PDR info: %s", err.Error())
			}
		}
		return nil
	}()

	if err != nil {
		log.Info().Msgf("Rejecting Session Establishment Request from: %s (error in applying IEs)", err)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, remoteSEID.SEID, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), nil
	}

	// Reassigning is the best I can think of for now
	association.Sessions[localSEID] = session
	conn.NodeAssociations[addr] = association

	additionalIEs := []*ie.IE{
		ie.NewCause(ie.CauseRequestAccepted),
		newIeNodeID(conn.nodeId),
		ie.NewFSEID(localSEID, conn.nodeAddrV4, nil),
	}

	//FIXME: process PDR with allocated UE/TEID only
	for _, pdr := range createdPDRs {
		if pdr.Ipv4 != nil {
			additionalIEs = append(additionalIEs, ie.NewCreatedPDR(ie.NewPDRID(0), ie.NewUEIPAddress(0, pdr.Ipv4.String(), "", 0, 0)))
		} else if pdr.Ipv6 != nil {

		} else {
			additionalIEs = append(additionalIEs, ie.NewCreatedPDR(ie.NewPDRID(0), ie.NewFTEID(0, pdr.Teid, conn.n3Address, nil, 0)))
		}
	}

	// Send SessionEstablishmentResponse
	estResp := message.NewSessionEstablishmentResponse(0, 0, remoteSEID.SEID, req.Sequence(), 0, additionalIEs...)
	PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRequestAccepted)).Inc()
	log.Info().Msgf("Session Establishment Request from %s accepted.", addr)
	return estResp, nil
}

func applyPDR(spdrInfo SPDRInfo, mapOperations ebpf.ForwardingPlaneController) {
	if spdrInfo.Ipv4 != nil {
		if err := mapOperations.PutPdrDownLink(spdrInfo.Ipv4, spdrInfo.PdrInfo); err != nil {
			log.Info().Msgf("Can't apply IPv4 PDR: %s", err.Error())
		}
	} else if spdrInfo.Ipv6 != nil {
		if err := mapOperations.PutDownlinkPdrIp6(spdrInfo.Ipv6, spdrInfo.PdrInfo); err != nil {
			log.Info().Msgf("Can't apply IPv6 PDR: %s", err.Error())
		}
	} else {
		if err := mapOperations.PutPdrUpLink(spdrInfo.Teid, spdrInfo.PdrInfo); err != nil {
			log.Info().Msgf("Can't apply GTP PDR: %s", err.Error())
		}
	}
}

func deletePDR(spdrInfo SPDRInfo, mapOperations ebpf.ForwardingPlaneController) error {
	if spdrInfo.Ipv4 != nil {
		if err := mapOperations.DeletePdrDownLink(spdrInfo.Ipv4); err != nil {
			return fmt.Errorf("Can't delete IPv4 PDR: %s", err.Error())
		}
	} else if spdrInfo.Ipv6 != nil {
		if err := mapOperations.DeleteDownlinkPdrIp6(spdrInfo.Ipv6); err != nil {
			return fmt.Errorf("Can't delete IPv6 PDR: %s", err.Error())
		}
	} else {
		if err := mapOperations.DeletePdrUpLink(spdrInfo.Teid); err != nil {
			return fmt.Errorf("Can't delete GTP PDR: %s", err.Error())
		}
	}
	return nil
}

func extractPDR(pdr *ie.IE, session *Session, spdrInfo *SPDRInfo, ipam *service.IPAM, seid uint64) error {

	if outerHeaderRemoval, err := pdr.OuterHeaderRemovalDescription(); err == nil {
		spdrInfo.PdrInfo.OuterHeaderRemoval = outerHeaderRemoval
	}
	if farid, err := pdr.FARID(); err == nil {
		spdrInfo.PdrInfo.FarId = session.GetFar(farid).GlobalId
	}
	if qerid, err := pdr.QERID(); err == nil {
		spdrInfo.PdrInfo.QerId = session.GetQer(qerid).GlobalId
	}

	pdi, err := pdr.PDI()
	if err != nil {
		return fmt.Errorf("PDI IE is missing")
	}

	if sdfFilter, err := pdr.SDFFilter(); err == nil {
		if sdfFilterParsed, err := ParseSdfFilter(sdfFilter.FlowDescription); err == nil {
			spdrInfo.PdrInfo.SdfFilter = &sdfFilterParsed
			// log.Printf("Sdf Filter Parsed: %+v", sdfFilterParsed)
		} else {
			return err
		}
	}

	//Bug in go-pfcp:
	//if fteid, err := pdr.FTEID(); err == nil {
	if teidPdiId := findIEindex(pdi, 21); teidPdiId != -1 { // IE Type F-TEID
		if fteid, err := pdi[teidPdiId].FTEID(); err == nil {
			if fteid.HasCh() {
				if fteid.HasChID() {
					//try to find teid previously allocated
					// if err, teid := teidCache.GetTEID(fteid.ChooseID); err == nil {
					// 	spdrInfo.Teid = teid
					// 	return nil
					// }
				}

				teid, err := ipam.AllocateTEID(seid)
				if err != nil {
					log.Error().Msgf("Allocate TEID error: %v", err)
					return fmt.Errorf("Can't allocate TEID: %s", causeToString(ie.CauseNoResourcesAvailable))
				}

				// if fteid.HasChID() {
				// 	teidCache.PutTEID(fteid.ChooseID, teid)
				// }

				spdrInfo.Teid = teid
				return nil
			} else {
				spdrInfo.Teid = fteid.TEID
				return nil
			}
		}
		return fmt.Errorf("F-TEID IE is missing")
	} else if ueIP, err := pdr.UEIPAddress(); err == nil {
		if ueIP.IPv4Address != nil {
			spdrInfo.Ipv4 = cloneIP(ueIP.IPv4Address)
		} else if ueIP.IPv6Address != nil {
			spdrInfo.Ipv6 = cloneIP(ueIP.IPv6Address)
		} else {
			return fmt.Errorf("UE IP Address IE is missing")
		}

		return nil
	} else {
		log.Info().Msg("Both F-TEID IE and UE IP Address IE are missing")
		return err
	}
}

func HandlePfcpSessionDeletionRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	req := msg.(*message.SessionDeletionRequest)
	log.Info().Msgf("Got Session Deletion Request from: %s. \n", addr)
	association, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Info().Msgf("Rejecting Session Deletion Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}
	printSessionDeleteRequest(req)

	session, ok := association.Sessions[req.SEID()]
	if !ok {
		log.Info().Msgf("Rejecting Session Deletion Request from: %s (unknown SEID)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseSessionContextNotFound)).Inc()
		return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseSessionContextNotFound)), nil
	}
	mapOperations := conn.mapOperations
	for _, pdrInfo := range session.PDRs {
		if err := deletePDR(pdrInfo, mapOperations); err != nil {
			PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
			return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), err
		}
	}
	for id := range session.FARs {
		if err := mapOperations.DeleteFar(id); err != nil {
			PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
			return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), err
		}
	}
	for id := range session.QERs {
		if err := mapOperations.DeleteQer(id); err != nil {
			PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
			return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), err
		}
	}
	log.Info().Msgf("Deleting session: %d", req.SEID())
	delete(association.Sessions, req.SEID())

	conn.ipam.ReleaseIP(req.SEID())

	PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRequestAccepted)).Inc()
	return message.NewSessionDeletionResponse(0, 0, session.RemoteSEID, req.Sequence(), 0, ie.NewCause(ie.CauseRequestAccepted)), nil
}

func HandlePfcpSessionModificationRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	req := msg.(*message.SessionModificationRequest)
	log.Info().Msgf("Got Session Modification Request from: %s. \n", addr)

	log.Info().Msgf("Finding association for %s", addr)
	association, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Info().Msgf("Rejecting Session Modification Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionModificationResponse(0, 0, req.SEID(), req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}

	log.Info().Msgf("Finding session %d", req.SEID())
	session, ok := association.Sessions[req.SEID()]
	if !ok {
		log.Info().Msgf("Rejecting Session Modification Request from: %s (unknown SEID)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseSessionContextNotFound)).Inc()
		return message.NewSessionModificationResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseSessionContextNotFound)), nil
	}

	// This IE shall be present if the CP function decides to change its F-SEID for the PFCP session. The UP function
	// shall use the new CP F-SEID for subsequent PFCP Session related messages for this PFCP Session
	if req.CPFSEID != nil {
		remoteSEID, err := req.CPFSEID.FSEID()
		if err == nil {
			session.RemoteSEID = remoteSEID.SEID

			association.Sessions[req.SEID()] = session // FIXME
			conn.NodeAssociations[addr] = association  // FIXME
		}
	}

	printSessionModificationRequest(req)

	// #TODO: Implement rollback on error
	err := func() error {
		mapOperations := conn.mapOperations

		for _, far := range req.CreateFAR {
			farInfo, err := composeFarInfo(far, conn.n3Address.To4(), ebpf.FarInfo{})
			if err != nil {
				log.Info().Msgf("Error extracting FAR info: %s", err.Error())
				continue
			}

			farid, _ := far.FARID()
			log.Info().Msgf("Saving FAR info to session: %d, %+v", farid, farInfo)
			if internalId, err := mapOperations.NewFar(farInfo); err == nil {
				session.NewFar(farid, internalId, farInfo)
			} else {
				log.Info().Msgf("Can't put FAR: %s", err.Error())
				return err
			}
		}

		for _, far := range req.UpdateFAR {
			farid, err := far.FARID()
			if err != nil {
				return err
			}
			sFarInfo := session.GetFar(farid)
			sFarInfo.FarInfo, err = composeFarInfo(far, conn.n3Address.To4(), sFarInfo.FarInfo)
			if err != nil {
				log.Info().Msgf("Error extracting FAR info: %s", err.Error())
				continue
			}
			log.Info().Msgf("Updating FAR info: %d, %+v", farid, sFarInfo)
			session.UpdateFar(farid, sFarInfo.FarInfo)
			if err := mapOperations.UpdateFar(sFarInfo.GlobalId, sFarInfo.FarInfo); err != nil {
				log.Info().Msgf("Can't update FAR: %s", err.Error())
			}
		}

		for _, far := range req.RemoveFAR {
			farid, _ := far.FARID()
			log.Info().Msgf("Removing FAR: %d", farid)
			sFarInfo := session.RemoveFar(farid)
			if err := mapOperations.DeleteFar(sFarInfo.GlobalId); err != nil {
				log.Info().Msgf("Can't remove FAR: %s", err.Error())
			}
		}

		for _, qer := range req.CreateQER {
			qerInfo := ebpf.QerInfo{}
			qerId, err := qer.QERID()
			if err != nil {
				return fmt.Errorf("QER ID missing")
			}
			updateQer(&qerInfo, qer)
			log.Info().Msgf("Saving QER info to session: %d, %+v", qerId, qerInfo)
			if internalId, err := mapOperations.NewQer(qerInfo); err == nil {
				session.NewQer(qerId, internalId, qerInfo)
			} else {
				log.Info().Msgf("Can't put QER: %s", err.Error())
				return err
			}
		}

		for _, qer := range req.UpdateQER {
			qerId, err := qer.QERID() // Probably will be used as ebpf map key
			if err != nil {
				return fmt.Errorf("QER ID missing")
			}
			sQerInfo := session.GetQer(qerId)
			updateQer(&sQerInfo.QerInfo, qer)
			log.Info().Msgf("Updating QER ID: %d, QER Info: %+v", qerId, sQerInfo)
			session.UpdateQer(qerId, sQerInfo.QerInfo)
			if err := mapOperations.UpdateQer(sQerInfo.GlobalId, sQerInfo.QerInfo); err != nil {
				log.Info().Msgf("Can't update QER: %s", err.Error())
				return err
			}
		}

		for _, qer := range req.RemoveQER {
			qerId, err := qer.QERID()
			if err != nil {
				return fmt.Errorf("QER ID missing")
			}
			log.Info().Msgf("Removing QER ID: %d", qerId)
			sQerInfo := session.RemoveQer(qerId)
			if err := mapOperations.DeleteQer(sQerInfo.GlobalId); err != nil {
				log.Info().Msgf("Can't remove QER: %s", err.Error())
				return err
			}
		}

		for _, pdr := range req.CreatePDR {
			// PDR should be created last, because we need to reference FARs and QERs global id
			pdrId, err := pdr.PDRID()
			if err != nil {
				log.Info().Msgf("PDR ID missing")
				continue
			}

			spdrInfo := SPDRInfo{PdrID: uint32(pdrId)}
			//if err := extractPDR(pdr, session, &spdrInfo, conn.ipam, req.SEID()); err == nil {
			if err := extractPDR(pdr, session, &spdrInfo, conn.ipam, req.SEID()); err == nil {
				//if err := extractPDR(pdr, session, &spdrInfo); err == nil {
				session.PutPDR(spdrInfo.PdrID, spdrInfo)
				applyPDR(spdrInfo, mapOperations)
			} else {
				log.Info().Msgf("Error extracting PDR info: %s", err.Error())
			}
		}

		for _, pdr := range req.UpdatePDR {
			pdrId, err := pdr.PDRID()
			if err != nil {
				return fmt.Errorf("PDR ID missing")
			}

			spdrInfo := session.GetPDR(pdrId)
			//if err := extractPDR(pdr, session, &spdrInfo, conn.ipam, req.SEID()); err == nil {
			if err := extractPDR(pdr, session, &spdrInfo, conn.ipam, req.SEID()); err == nil {
				session.PutPDR(uint32(pdrId), spdrInfo)
				applyPDR(spdrInfo, mapOperations)
			} else {
				log.Printf("Error extracting PDR info: %s", err.Error())
			}
		}

		for _, pdr := range req.RemovePDR {
			pdrId, _ := pdr.PDRID()
			if _, ok := session.PDRs[uint32(pdrId)]; ok {
				log.Info().Msgf("Removing uplink PDR: %d", pdrId)
				sPDRInfo := session.RemovePDR(uint32(pdrId))

				if err := deletePDR(sPDRInfo, mapOperations); err != nil {
					log.Info().Msgf("Failed to remove uplink PDR: %v", err)
				}
			}
		}

		return nil
	}()
	if err != nil {
		log.Info().Msgf("Rejecting Session Modification Request from: %s (failed to apply rules)", err)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
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
	PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRequestAccepted)).Inc()
	return modResp, nil
}

func convertErrorToIeCause(err error) *ie.IE {
	switch err {
	case errMandatoryIeMissing:
		return ie.NewCause(ie.CauseMandatoryIEMissing)
	case errNoEstablishedAssociation:
		return ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)
	default:
		log.Info().Msgf("Unknown error: %s", err.Error())
		return ie.NewCause(ie.CauseRequestRejected)
	}
}

func validateRequest(nodeId *ie.IE, cpfseid *ie.IE) (fseid *ie.FSEIDFields, err error) {
	if nodeId == nil || cpfseid == nil {
		return nil, errMandatoryIeMissing
	}

	_, err = nodeId.NodeID()
	if err != nil {
		return nil, errMandatoryIeMissing
	}

	fseid, err = cpfseid.FSEID()
	if err != nil {
		return nil, errMandatoryIeMissing
	}

	return fseid, nil
}

func findIEindex(ieArr []*ie.IE, ieType uint16) int {
	arrIndex := slices.IndexFunc(ieArr, func(ie *ie.IE) bool {
		return ie.Type == ieType
	})
	return arrIndex
}

func causeToString(cause uint8) string {
	switch cause {
	case ie.CauseRequestAccepted:
		return "RequestAccepted"
	case ie.CauseRequestRejected:
		return "RequestRejected"
	case ie.CauseSessionContextNotFound:
		return "SessionContextNotFound"
	case ie.CauseMandatoryIEMissing:
		return "MandatoryIEMissing"
	case ie.CauseConditionalIEMissing:
		return "ConditionalIEMissing"
	case ie.CauseInvalidLength:
		return "InvalidLength"
	case ie.CauseMandatoryIEIncorrect:
		return "MandatoryIEIncorrect"
	case ie.CauseInvalidForwardingPolicy:
		return "InvalidForwardingPolicy"
	case ie.CauseInvalidFTEIDAllocationOption:
		return "InvalidFTEIDAllocationOption"
	case ie.CauseNoEstablishedPFCPAssociation:
		return "NoEstablishedPFCPAssociation"
	case ie.CauseRuleCreationModificationFailure:
		return "RuleCreationModificationFailure"
	case ie.CausePFCPEntityInCongestion:
		return "PFCPEntityInCongestion"
	case ie.CauseNoResourcesAvailable:
		return "NoResourcesAvailable"
	case ie.CauseServiceNotSupported:
		return "ServiceNotSupported"
	case ie.CauseSystemFailure:
		return "SystemFailure"
	case ie.CauseRedirectionRequested:
		return "RedirectionRequested"
	default:
		return "UnknownCause"
	}
}

func cloneIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func composeFarInfo(far *ie.IE, localIp net.IP, farInfo ebpf.FarInfo) (ebpf.FarInfo, error) {
	farInfo.LocalIP = binary.LittleEndian.Uint32(localIp)
	if applyAction, err := far.ApplyAction(); err == nil {
		farInfo.Action = applyAction[0]
	}
	var forward []*ie.IE
	var err error
	if far.Type == ie.CreateFAR {
		forward, err = far.ForwardingParameters()
	} else if far.Type == ie.UpdateFAR {
		forward, err = far.UpdateForwardingParameters()
	} else {
		return ebpf.FarInfo{}, fmt.Errorf("unsupported IE type")
	}
	if err == nil {
		outerHeaderCreationIndex := findIEindex(forward, 84) // IE Type Outer Header Creation
		if outerHeaderCreationIndex == -1 {
			log.Info().Msg("WARN: No OuterHeaderCreation")
		} else {
			outerHeaderCreation, _ := forward[outerHeaderCreationIndex].OuterHeaderCreation()
			farInfo.OuterHeaderCreation = uint8(outerHeaderCreation.OuterHeaderCreationDescription >> 8)
			farInfo.Teid = outerHeaderCreation.TEID
			if outerHeaderCreation.HasIPv4() {
				farInfo.RemoteIP = binary.LittleEndian.Uint32(outerHeaderCreation.IPv4Address)
			}
			if outerHeaderCreation.HasIPv6() {
				log.Info().Msg("WARN: IPv6 not supported yet, ignoring")
				return ebpf.FarInfo{}, fmt.Errorf("IPv6 not supported yet")
			}
		}
	}
	transportLevelMarking, err := GetTransportLevelMarking(far)
	if err == nil {
		farInfo.TransportLevelMarking = transportLevelMarking
	}
	return farInfo, nil
}

func updateQer(qerInfo *ebpf.QerInfo, qer *ie.IE) {

	gateStatusDL, err := qer.GateStatusDL()
	if err == nil {
		qerInfo.GateStatusDL = gateStatusDL
	}
	gateStatusUL, err := qer.GateStatusUL()
	if err == nil {
		qerInfo.GateStatusUL = gateStatusUL
	}
	maxBitrateDL, err := qer.MBRDL()
	if err == nil {
		qerInfo.MaxBitrateDL = uint32(maxBitrateDL) * 1000
	}
	maxBitrateUL, err := qer.MBRUL()
	if err == nil {
		qerInfo.MaxBitrateUL = uint32(maxBitrateUL) * 1000
	}
	qfi, err := qer.QFI()
	if err == nil {
		qerInfo.Qfi = qfi
	}
	qerInfo.StartUL = 0
	qerInfo.StartDL = 0
}

func GetTransportLevelMarking(far *ie.IE) (uint16, error) {
	for _, informationalElement := range far.ChildIEs {
		if informationalElement.Type == ie.TransportLevelMarking {
			return informationalElement.TransportLevelMarking()
		}
	}
	return 0, fmt.Errorf("no TransportLevelMarking found")
}
