package core

import (
	"encoding/binary"
	"fmt"
	"github.com/edgecomllc/eupf/cmd/ebpf"
	"log"
	"net"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
	"golang.org/x/exp/slices"
)

var errMandatoryIeMissing = fmt.Errorf("mandatory IE missing")
var errNoEstablishedAssociation = fmt.Errorf("no established association")

// #TODO: Extract Create/Update/Delete IE to separate functions
// #TODO: Research how to merge UplinkPDRs and DownlinkPDRs

func HandlePfcpSessionEstablishmentRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	req := msg.(*message.SessionEstablishmentRequest)
	log.Printf("Got Session Establishment Request from: %s.", addr)
	remoteSEID, err := validateRequest(req.NodeID, req.CPFSEID)
	if err != nil {
		log.Printf("Rejecting Session Establishment Request from: %s (missing NodeID or F-SEID)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEMissing)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, convertErrorToIeCause(err)), nil
	}

	association, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Printf("Rejecting Session Establishment Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}

	localSEID := association.NewLocalSEID()

	session := NewSession(localSEID, remoteSEID.SEID)

	printSessionEstablishmentRequest(req)
	// #TODO: Implement rollback on error
	err = func() error {
		mapOperations := conn.mapOperations
		for _, far := range req.CreateFAR {
			farInfo, err := composeFarInfo(far, conn.n3Address.To4(), ebpf.FarInfo{})
			if err != nil {
				log.Printf("Error extracting FAR info: %s", err.Error())
				continue
			}

			farid, _ := far.FARID()
			log.Printf("Saving FAR info to session: %d, %+v", farid, farInfo)
			if internalId, err := mapOperations.NewFar(farInfo); err == nil {
				session.NewFar(farid, internalId, farInfo)
			} else {
				log.Printf("Can't put FAR: %s", err.Error())
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
			log.Printf("Saving QER info to session: %d, %+v", qerId, qerInfo)
			if internalId, err := mapOperations.NewQer(qerInfo); err == nil {
				session.NewQer(qerId, internalId, qerInfo)
			} else {
				log.Printf("Can't put QER: %s", err.Error())
				return err
			}
		}

		for _, pdr := range req.CreatePDR {
			// PDR should be created last, because we need to reference FARs and QERs global id
			spdrInfo := SPDRInfo{}
			pdrId, err := pdr.PDRID()
			if err != nil {
				return fmt.Errorf("PDR ID missing")
			}
			updateSPDRInfo(pdr, &spdrInfo, session)
			pdi, err := pdr.PDI()
			if err != nil {
				return err
			}
			srcIfacePdiId := findIEindex(pdi, 20) // IE Type source interface
			srcInterface, _ := pdi[srcIfacePdiId].SourceInterface()
			switch srcInterface {
			case ie.SrcInterfaceAccess, ie.SrcInterfaceCPFunction:
				{
					sdfFilterId := findIEindex(pdi, 23) // IE Type SDF Filter
					if sdfFilterId != -1 {
						log.Printf("WARN: SDF Filter is not supported yet. Ignore PDR")
						continue
					}

					if err := applyUplinkPDR(pdi, spdrInfo, pdrId, session, mapOperations); err != nil {
						log.Printf("Errored while applying PDR: %s", err.Error())
						return err
					}
				}
			case ie.SrcInterfaceCore, ie.SrcInterfaceSGiLANN6LAN:
				{
					err := applyDownlinkPDR(pdi, spdrInfo, pdrId, session, mapOperations)
					if err == fmt.Errorf("IPv6 not supported") {
						continue
					}
					if err != nil {
						log.Printf("Errored[ while applying PDR: %s", err.Error())
						return err
					}
				}
			default:
				log.Printf("WARN: Unsupported Source Interface type: %d", srcInterface)
			}
		}
		return nil
	}()

	if err != nil {
		log.Printf("Rejecting Session Establishment Request from: %s (error in applying IEs)", err)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, remoteSEID.SEID, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), nil
	}

	// Reassigning is the best I can think of for now
	association.Sessions[localSEID] = session
	conn.NodeAssociations[addr] = association

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
	PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRequestAccepted)).Inc()
	return estResp, nil
}

func HandlePfcpSessionDeletionRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	req := msg.(*message.SessionDeletionRequest)
	log.Printf("Got Session Deletion Request from: %s. \n", addr)
	association, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Printf("Rejecting Session Deletion Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}
	printSessionDeleteRequest(req)

	session, ok := association.Sessions[req.SEID()]
	if !ok {
		log.Printf("Rejecting Session Deletion Request from: %s (unknown SEID)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseSessionContextNotFound)).Inc()
		return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseSessionContextNotFound)), nil
	}
	mapOperations := conn.mapOperations
	for _, pdrInfo := range session.UplinkPDRs {
		if err := mapOperations.DeletePdrUpLink(pdrInfo.Teid); err != nil {
			PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
			return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), err
		}
	}
	for _, pdrInfo := range session.DownlinkPDRs {
		if err := mapOperations.DeletePdrDownLink(pdrInfo.Ipv4); err != nil {
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
	log.Printf("Deleting session: %d", req.SEID())
	delete(association.Sessions, req.SEID())

	PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRequestAccepted)).Inc()
	return message.NewSessionDeletionResponse(0, 0, session.RemoteSEID, req.Sequence(), 0, ie.NewCause(ie.CauseRequestAccepted)), nil
}

func HandlePfcpSessionModificationRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	req := msg.(*message.SessionModificationRequest)
	log.Printf("Got Session Modification Request from: %s. \n", addr)

	log.Printf("Finding association for %s", addr)
	association, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Printf("Rejecting Session Modification Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionModificationResponse(0, 0, req.SEID(), req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}

	log.Printf("Finding session %d", req.SEID())
	session, ok := association.Sessions[req.SEID()]
	if !ok {
		log.Printf("Rejecting Session Modification Request from: %s (unknown SEID)", addr)
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
				log.Printf("Error extracting FAR info: %s", err.Error())
				continue
			}

			farid, _ := far.FARID()
			log.Printf("Saving FAR info to session: %d, %+v", farid, farInfo)
			if internalId, err := mapOperations.NewFar(farInfo); err == nil {
				session.NewFar(farid, internalId, farInfo)
			} else {
				log.Printf("Can't put FAR: %s", err.Error())
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
				log.Printf("Error extracting FAR info: %s", err.Error())
				continue
			}
			log.Printf("Updating FAR info: %d, %+v", farid, sFarInfo)
			session.UpdateFar(farid, sFarInfo.FarInfo)
			if err := mapOperations.UpdateFar(sFarInfo.GlobalId, sFarInfo.FarInfo); err != nil {
				log.Printf("Can't update FAR: %s", err.Error())
			}
		}

		for _, removeFar := range req.RemoveFAR {
			farid, _ := removeFar.FARID()
			log.Printf("Removing FAR: %d", farid)
			sFarInfo := session.RemoveFar(farid)
			if err := mapOperations.DeleteFar(sFarInfo.GlobalId); err != nil {
				log.Printf("Can't remove FAR: %s", err.Error())
			}
		}

		for _, qer := range req.CreateQER {
			qerInfo := ebpf.QerInfo{}
			qerId, err := qer.QERID()
			if err != nil {
				return fmt.Errorf("QER ID missing")
			}
			updateQer(&qerInfo, qer)
			log.Printf("Saving QER info to session: %d, %+v", qerId, qerInfo)
			if internalId, err := mapOperations.NewQer(qerInfo); err == nil {
				session.NewQer(qerId, internalId, qerInfo)
			} else {
				log.Printf("Can't put QER: %s", err.Error())
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
			log.Printf("Updating QER ID: %d, QER Info: %+v", qerId, sQerInfo)
			session.UpdateQer(qerId, sQerInfo.QerInfo)
			if err := mapOperations.UpdateQer(sQerInfo.GlobalId, sQerInfo.QerInfo); err != nil {
				log.Printf("Can't update QER: %s", err.Error())
				return err
			}
		}

		for _, qer := range req.RemoveQER {
			qerId, err := qer.QERID()
			if err != nil {
				return fmt.Errorf("QER ID missing")
			}
			log.Printf("Removing QER ID: %d", qerId)
			sQerInfo := session.RemoveQer(qerId)
			log.Printf("Removing QER ID: %d", qerId)
			if err := mapOperations.DeleteQer(sQerInfo.GlobalId); err != nil {
				log.Printf("Can't remove QER: %s", err.Error())
				return err
			}
		}

		for _, pdr := range req.CreatePDR {
			// PDR should be created last, because we need to reference FARs and QERs global id
			spdrInfo := SPDRInfo{}
			pdrId, err := pdr.PDRID()
			if err != nil {
				return fmt.Errorf("PDR ID missing")
			}
			updateSPDRInfo(pdr, &spdrInfo, session)
			pdi, err := pdr.PDI()
			if err != nil {
				return err
			}
			srcIfacePdiId := findIEindex(pdi, 20) // IE Type source interface
			srcInterface, _ := pdi[srcIfacePdiId].SourceInterface()
			switch srcInterface {
			case ie.SrcInterfaceAccess, ie.SrcInterfaceCPFunction:
				{
					sdfFilterId := findIEindex(pdi, 23) // IE Type SDF Filter
					if sdfFilterId != -1 {
						log.Printf("WARN: SDF Filter is not supported yet. Ignore PDR")
						continue
					}

					if err := applyUplinkPDR(pdi, spdrInfo, pdrId, session, mapOperations); err != nil {
						log.Printf("Errored while applying PDR: %s", err.Error())
						return err
					}
				}
			case ie.SrcInterfaceCore, ie.SrcInterfaceSGiLANN6LAN:
				{
					err := applyDownlinkPDR(pdi, spdrInfo, pdrId, session, mapOperations)
					if err == fmt.Errorf("IPv6 not supported") {
						continue
					}
					if err != nil {
						log.Printf("Errored[ while applying PDR: %s", err.Error())
						return err
					}
				}
			default:
				log.Printf("WARN: Unsupported Source Interface type: %d", srcInterface)
			}
		}

		for _, pdr := range req.UpdatePDR {
			pdrId, err := pdr.PDRID()
			if err != nil {
				return fmt.Errorf("PDR ID missing")
			}
			pdi, err := pdr.PDI()
			if err != nil {
				return err
			}
			srcIfacePdiId := findIEindex(pdi, 20) // IE Type source interface
			srcInterface, _ := pdi[srcIfacePdiId].SourceInterface()
			switch srcInterface {
			case ie.SrcInterfaceAccess, ie.SrcInterfaceCPFunction:
				{
					spdrInfo := session.GetUplinkPDR(pdrId)
					updateSPDRInfo(pdr, &spdrInfo, session)
					if err := applyUplinkPDR(pdi, spdrInfo, pdrId, session, mapOperations); err != nil {
						log.Printf("Errored while applying PDR: %s", err.Error())
						return err
					}
				}
			case ie.SrcInterfaceCore, ie.SrcInterfaceSGiLANN6LAN:
				{
					spdrInfo := session.GetDownlinkPDR(pdrId)
					updateSPDRInfo(pdr, &spdrInfo, session)
					err = applyDownlinkPDR(pdi, spdrInfo, pdrId, session, mapOperations)
					if err == fmt.Errorf("IPv6 not supported") {
						continue
					}
					if err != nil {
						log.Printf("Errored while applying PDR: %s", err.Error())
						return err
					}
				}
			default:
				log.Printf("WARN: Unsupported Source Interface type: %d", srcInterface)
			}
		}

		for _, pdr := range req.RemovePDR {
			pdrId, _ := pdr.PDRID()
			if _, ok := session.UplinkPDRs[uint32(pdrId)]; ok {
				log.Printf("Removing uplink PDR: %d", pdrId)
				session.RemoveUplinkPDR(uint32(pdrId))
				if err := mapOperations.DeletePdrUpLink(session.UplinkPDRs[uint32(pdrId)].Teid); err != nil {
					log.Printf("Failed to remove uplink PDR: %v", err)
				}
			}
			if _, ok := session.DownlinkPDRs[uint32(pdrId)]; ok {
				log.Printf("Removing downlink PDR: %d", pdrId)
				session.RemoveDownlinkPDR(uint32(pdrId))
				if err := mapOperations.DeletePdrDownLink(session.DownlinkPDRs[uint32(pdrId)].Ipv4); err != nil {
					log.Printf("Failed to remove downlink PDR: %v", err)
				}
			}
		}

		return nil
	}()
	if err != nil {
		log.Printf("Rejecting Session Modification Request from: %s (failed to apply rules)", err)
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

func updateSPDRInfo(pdr *ie.IE, spdrInfo *SPDRInfo, session *Session) {
	if outerHeaderRemoval, err := pdr.OuterHeaderRemovalDescription(); err == nil {
		spdrInfo.PdrInfo.OuterHeaderRemoval = outerHeaderRemoval
	}
	if farid, err := pdr.FARID(); err == nil {
		spdrInfo.PdrInfo.FarId = session.GetFar(farid).GlobalId
	}
	if qerid, err := pdr.QERID(); err == nil {
		spdrInfo.PdrInfo.QerId = session.GetQer(qerid).GlobalId
	}
}

func convertErrorToIeCause(err error) *ie.IE {
	switch err {
	case errMandatoryIeMissing:
		return ie.NewCause(ie.CauseMandatoryIEMissing)
	case errNoEstablishedAssociation:
		return ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)
	default:
		log.Printf("Unknown error: %s", err.Error())
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

func applyUplinkPDR(pdi []*ie.IE, spdrInfo SPDRInfo, pdrId uint16, session *Session, mapOperations ebpf.ForwardingPlaneController) error {
	// IE Type F-TEID
	if teidPdiId := findIEindex(pdi, 21); teidPdiId != -1 {
		if fteid, err := pdi[teidPdiId].FTEID(); err == nil {
			spdrInfo.Teid = fteid.TEID
			session.PutUplinkPDR(uint32(pdrId), spdrInfo)
			if err := mapOperations.PutPdrUpLink(spdrInfo.Teid, spdrInfo.PdrInfo); err != nil {
				log.Printf("Can't put uplink PDR: %s", err.Error())
			}
		} else {
			log.Println(err)
			return err
		}
	} else {
		log.Println("F-TEID IE missing")
	}
	return nil
}

func cloneIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func applyDownlinkPDR(pdi []*ie.IE, spdrInfo SPDRInfo, pdrId uint16, session *Session, mapOperations ebpf.ForwardingPlaneController) error {
	// IE Type UE IP Address
	if ueipPdiId := findIEindex(pdi, 93); ueipPdiId != -1 {
		ueIp, _ := pdi[ueipPdiId].UEIPAddress()
		if ueIp.IPv4Address == nil && ueIp.IPv6Address == nil {
			return fmt.Errorf("UE IP Address IE missing")
		}
		if ueIp.IPv4Address != nil {
			// net.IP is a trap, it needs to be copied, otherwise it will be overwritten by next packet.
			spdrInfo.Ipv4 = cloneIP(ueIp.IPv4Address)
			session.PutDownlinkPDR(uint32(pdrId), spdrInfo)
			if err := mapOperations.PutPdrDownLink(spdrInfo.Ipv4, spdrInfo.PdrInfo); err != nil {
				log.Printf("Can't put downlink PDR: %s", err.Error())
			}
		}
		if ueIp.IPv6Address != nil {
			spdrInfo.Ipv6 = cloneIP(ueIp.IPv6Address)
			session.PutDownlinkPDR(uint32(pdrId), spdrInfo)
			if err := mapOperations.PutDownlinkPdrIp6(spdrInfo.Ipv6, spdrInfo.PdrInfo); err != nil {
				log.Printf("Can't put downlink PDR: %s", err.Error())
			}
		}
	} else {
		log.Println("UE IP Address IE missing")
	}
	return nil
}

func composeFarInfo(far *ie.IE, localIp net.IP, farInfo ebpf.FarInfo) (ebpf.FarInfo, error) {
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
			log.Println("WARN: No OuterHeaderCreation")
		} else {
			outerHeaderCreation, _ := forward[outerHeaderCreationIndex].OuterHeaderCreation()
			farInfo.OuterHeaderCreation = uint8(outerHeaderCreation.OuterHeaderCreationDescription >> 8)
			farInfo.Teid = outerHeaderCreation.TEID
			if outerHeaderCreation.HasIPv4() {
				farInfo.RemoteIP = binary.LittleEndian.Uint32(outerHeaderCreation.IPv4Address)
				farInfo.LocalIP = binary.LittleEndian.Uint32(localIp)
			}
			if outerHeaderCreation.HasIPv6() {
				log.Print("WARN: IPv6 not supported yet, ignoring")
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
