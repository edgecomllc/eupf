package core

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

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
		log.Warn().Msgf("Rejecting Session Establishment Request from: %s (missing NodeID or F-SEID)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEMissing)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, newIeNodeID(conn.nodeId), convertErrorToIeCause(err)), nil
	}

	association, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Warn().Msgf("Rejecting Session Establishment Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, newIeNodeID(conn.nodeId), ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}

	localSEID := association.NewLocalSEID()

	session := NewSession(localSEID, remoteSEID.SEID)

	printSessionEstablishmentRequest(req)
	// #TODO: Implement rollback on error
	createdPDRs := []SPDRInfo{}
	pdrContext := NewPDRCreationContext(session, conn.ResourceManager)

	err = func() error {
		mapOperations := conn.mapOperations
		for _, far := range req.CreateFAR {
			farInfo, err := composeFarInfo(far, ebpf.FarInfo{})
			if err != nil {
				log.Error().Err(err).Msg("Error extracting FAR info")
				return err
			}

			farid, _ := far.FARID()
			log.Info().Msgf("Saving FAR info to session: %d, %+v", farid, farInfo)
			if internalId, err := mapOperations.NewFar(farInfo); err == nil {
				session.NewFar(farid, internalId, farInfo)
			} else {
				log.Error().Err(err).Msg("Can't put FAR")
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
				log.Error().Err(err).Msg("Can't put QER")
				return err
			}
		}

		for _, urr := range req.CreateURR {
			urrInfo := ebpf.UrrInfo{}
			urrId, err := urr.URRID()
			if err != nil {
				return fmt.Errorf("URR ID missing")
			}
			updateUrr(&urrInfo, urr)
			log.Info().Msgf("Saving URR info to session: %d, %+v", urrId, urrInfo)
			if internalId, err := mapOperations.NewUrr(urrInfo); err == nil {
				session.NewUrr(urrId, internalId, urrInfo)
			} else {
				log.Error().Err(err).Msg("Can't put URR")
				return err
			}
		}

		for _, pdr := range req.CreatePDR {
			// PDR should be created last, because we need to reference FARs and QERs global id
			pdrId, err := pdr.PDRID()
			if err != nil {
				return fmt.Errorf("PDR ID missing")
			}

			spdrInfo := SPDRInfo{PdrID: uint32(pdrId)}

			if err := pdrContext.extractPDR(pdr, &spdrInfo); err == nil {
				session.PutPDR(spdrInfo.PdrID, spdrInfo)
				if err := applyPDR(spdrInfo, mapOperations); err == nil {
					createdPDRs = append(createdPDRs, spdrInfo)
				} else {
					return err
				}
			} else {
				log.Error().Err(err).Msg("error extracting PDR info")
				return err
			}
		}

		return nil
	}()

	if err != nil {
		log.Warn().Msgf("Rejecting Session Establishment Request from: %s (error in applying IEs)", err)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, remoteSEID.SEID, req.Sequence(), 0, newIeNodeID(conn.nodeId), ie.NewCause(ie.CauseRuleCreationModificationFailure)), nil
	}

	// Reassigning is the best I can think of for now
	association.Sessions[localSEID] = session
	conn.NodeAssociations[addr] = association

	additionalIEs := []*ie.IE{
		newIeNodeID(conn.nodeId),
		ie.NewCause(ie.CauseRequestAccepted),
		ie.NewFSEID(localSEID, cloneIP(conn.nodeAddrV4), nil),
	}

	pdrIEs := processCreatedPDRs(createdPDRs, cloneIP(conn.n3Address))
	additionalIEs = append(additionalIEs, pdrIEs...)

	// Send SessionEstablishmentResponse
	estResp := message.NewSessionEstablishmentResponse(0, 0, remoteSEID.SEID, req.Sequence(), 0, additionalIEs...)
	PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRequestAccepted)).Inc()
	log.Info().Msgf("Session Establishment Request from %s accepted.", addr)
	return estResp, nil
}

func HandlePfcpSessionDeletionRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	req := msg.(*message.SessionDeletionRequest)
	log.Info().Msgf("Got Session Deletion Request from: %s. \n", addr)
	association, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Warn().Msgf("Rejecting Session Deletion Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}
	printSessionDeleteRequest(req)

	session, ok := association.Sessions[req.SEID()]
	if !ok {
		log.Warn().Msgf("Rejecting Session Deletion Request from: %s (unknown SEID)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseSessionContextNotFound)).Inc()
		return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseSessionContextNotFound)), nil
	}
	deletedURRs := make([]*ie.IE, 0, len(session.URRs))
	mapOperations := conn.mapOperations
	pdrContext := NewPDRCreationContext(session, conn.ResourceManager)
	for _, pdrInfo := range session.PDRs {
		if err := pdrContext.deletePDR(pdrInfo, mapOperations); err != nil {
			PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
			return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), err
		}
	}
	for _, farInfo := range session.FARs {
		if err := mapOperations.DeleteFar(farInfo.GlobalId); err != nil {
			PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
			return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), err
		}
	}
	for _, qerInfo := range session.QERs {
		if err := mapOperations.DeleteQer(qerInfo.GlobalId); err != nil {
			PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
			return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), err
		}
	}
	for id, urr := range session.URRs {
		err, urrInfo := mapOperations.DeleteUrr(urr.GlobalId)
		if err != nil {
			log.Error().Msgf("WARN: mapOperations failed to delete URR: %d, %s", id, err.Error())
			continue
		}
		urr.ReportSeqNumber = urr.ReportSeqNumber + 1
		deletedURRs = append(deletedURRs, ie.NewUsageReportWithinSessionDeletionResponse(
			ie.NewURRID(id),
			ie.NewURSEQN(urr.ReportSeqNumber),
			ie.NewUsageReportTrigger([]uint8{0, 1 << 3, 0}...),
			ie.NewEndTime(time.Now()),
			ie.NewVolumeMeasurement(0x7, urrInfo.UplinkVolume+urrInfo.DownlinkVolume, urrInfo.UplinkVolume, urrInfo.DownlinkVolume, 0, 0, 0),
		))
	}

	additionalIEs := []*ie.IE{
		ie.NewCause(ie.CauseRequestAccepted),
	}
	if len(deletedURRs) != 0 {
		additionalIEs = append(additionalIEs, deletedURRs...)
	}

	log.Info().Msgf("Deleting session: %d", req.SEID())
	delete(association.Sessions, req.SEID())

	conn.ReleaseResources(req.SEID())

	PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRequestAccepted)).Inc()
	delResp := message.NewSessionDeletionResponse(0, 0, session.RemoteSEID, req.Sequence(), 0, additionalIEs...)
	return delResp, nil
}

func HandlePfcpSessionModificationRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	req := msg.(*message.SessionModificationRequest)
	log.Info().Msgf("Got Session Modification Request from: %s. \n", addr)

	log.Info().Msgf("Finding association for %s", addr)
	association, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Warn().Msgf("Rejecting Session Modification Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionModificationResponse(0, 0, req.SEID(), req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}

	log.Info().Msgf("Finding session %d", req.SEID())
	session, ok := association.Sessions[req.SEID()]
	if !ok {
		log.Warn().Msgf("Rejecting Session Modification Request from: %s (unknown SEID)", addr)
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
	createdPDRs := []SPDRInfo{}
	removedURRs := make([]*ie.IE, 0, len(req.RemoveURR))
	pdrContext := NewPDRCreationContext(session, conn.ResourceManager)

	err := func() error {
		mapOperations := conn.mapOperations

		for _, far := range req.CreateFAR {
			farInfo, err := composeFarInfo(far, ebpf.FarInfo{})
			if err != nil {
				log.Warn().Err(err).Msg("Error extracting FAR info")
				return err
			}

			farid, _ := far.FARID()
			log.Info().Msgf("Saving FAR info to session: %d, %+v", farid, farInfo)
			if internalId, err := mapOperations.NewFar(farInfo); err == nil {
				session.NewFar(farid, internalId, farInfo)
			} else {
				log.Error().Err(err).Msg("Can't put FAR")
				return err
			}
		}

		for _, far := range req.UpdateFAR {
			farid, err := far.FARID()
			if err != nil {
				return err
			}
			sFarInfo := session.GetFar(farid)
			sFarInfo.FarInfo, err = composeFarInfo(far, sFarInfo.FarInfo)
			if err != nil {
				log.Warn().Err(err).Msg("Error extracting FAR info")
				return err
			}
			log.Info().Msgf("Updating FAR info: %d, %+v", farid, sFarInfo)
			session.UpdateFar(farid, sFarInfo.FarInfo)
			if err := mapOperations.UpdateFar(sFarInfo.GlobalId, sFarInfo.FarInfo); err != nil {
				log.Error().Err(err).Msg("Can't update FAR")
				return err
			}
		}

		for _, far := range req.RemoveFAR {
			farid, _ := far.FARID()
			log.Info().Msgf("Removing FAR: %d", farid)
			sFarInfo := session.RemoveFar(farid)
			if err := mapOperations.DeleteFar(sFarInfo.GlobalId); err != nil {
				log.Error().Err(err).Msg("Can't remove FAR")
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
				log.Error().Err(err).Msg("Can't put QER")
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
				log.Error().Err(err).Msg("Can't update QER")
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
				log.Error().Err(err).Msg("Can't remove QER")
				return err
			}
		}

		for _, urr := range req.CreateURR {
			urrInfo := ebpf.UrrInfo{}
			urrId, err := urr.URRID()
			if err != nil {
				return fmt.Errorf("URR ID missing")
			}
			updateUrr(&urrInfo, urr)
			log.Info().Msgf("Saving URR info to session: %d, %+v", urrId, urrInfo)
			if internalId, err := mapOperations.NewUrr(urrInfo); err == nil {
				session.NewUrr(urrId, internalId, urrInfo)
			} else {
				log.Error().Err(err).Msg("Can't put URR")
				return err
			}
		}

		for _, urr := range req.UpdateURR {
			urrId, err := urr.URRID()
			if err != nil {
				return fmt.Errorf("URR ID missing")
			}
			sUrrInfo := session.GetUrr(urrId)
			updateUrr(&sUrrInfo.UrrInfo, urr)
			log.Info().Msgf("Updating URR ID: %d, URR Info: %+v", urrId, sUrrInfo)
			session.UpdateUrr(urrId, sUrrInfo.UrrInfo)
			if err := mapOperations.UpdateUrr(sUrrInfo.GlobalId, sUrrInfo.UrrInfo); err != nil {
				log.Error().Err(err).Msg("Can't update URR")
				return err
			}
		}

		for _, urr := range req.RemoveURR {
			urrId, err := urr.URRID()
			if err != nil {
				return fmt.Errorf("URR ID missing")
			}
			log.Info().Msgf("Removing URR ID: %d", urrId)
			sUrrInfo := session.RemoveUrr(urrId)

			err, urrInfo := mapOperations.DeleteUrr(sUrrInfo.GlobalId)
			if err != nil {
				log.Error().Err(err).Msg("Can't remove URR")
				return err
			}

			sUrrInfo.ReportSeqNumber = sUrrInfo.ReportSeqNumber + 1
			removedURRs = append(removedURRs, ie.NewUsageReportWithinSessionModificationResponse(
				ie.NewURRID(urrId),
				ie.NewURSEQN(sUrrInfo.ReportSeqNumber),
				ie.NewUsageReportTrigger([]uint8{0, 1 << 3, 0}...),
				ie.NewEndTime(time.Now()),
				ie.NewVolumeMeasurement(0x7, urrInfo.UplinkVolume+urrInfo.DownlinkVolume, urrInfo.UplinkVolume, urrInfo.DownlinkVolume, 0, 0, 0),
			))
		}

		for _, pdr := range req.CreatePDR {
			// PDR should be created last, because we need to reference FARs and QERs global id
			pdrId, err := pdr.PDRID()
			if err != nil {
				log.Warn().Err(err).Msg("PDR ID missing")
				return err
			}

			spdrInfo := SPDRInfo{PdrID: uint32(pdrId)}

			if err := pdrContext.extractPDR(pdr, &spdrInfo); err == nil {
				session.PutPDR(spdrInfo.PdrID, spdrInfo)
				if err := applyPDR(spdrInfo, mapOperations); err == nil {
					createdPDRs = append(createdPDRs, spdrInfo)
				} else {
					return err
				}
			} else {
				log.Info().Err(err).Msg("Error extracting PDR info")
				return err
			}
		}

		for _, pdr := range req.UpdatePDR {
			pdrId, err := pdr.PDRID()
			if err != nil {
				return fmt.Errorf("PDR ID missing")
			}

			spdrInfo := session.GetPDR(pdrId)
			if err := pdrContext.extractPDR(pdr, &spdrInfo); err == nil {
				session.PutPDR(uint32(pdrId), spdrInfo)
				if err := applyPDR(spdrInfo, mapOperations); err != nil {
					return err
				}
			} else {
				log.Warn().Err(err).Msg("Error extracting PDR info")
				return err
			}
		}

		for _, pdr := range req.RemovePDR {
			pdrId, _ := pdr.PDRID()
			if _, ok := session.PDRs[uint32(pdrId)]; ok {
				log.Info().Msgf("Removing uplink PDR: %d", pdrId)
				sPDRInfo := session.RemovePDR(uint32(pdrId))

				if err := pdrContext.deletePDR(sPDRInfo, mapOperations); err != nil {
					log.Error().Err(err).Msg("Failed to remove uplink PDR")
					return err
				}
			}
		}

		return nil
	}()
	if err != nil {
		log.Warn().Msgf("Rejecting Session Modification Request from: %s (failed to apply rules)", err)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
		return message.NewSessionModificationResponse(0, 0, session.RemoteSEID, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), nil
	}

	association.Sessions[req.SEID()] = session

	additionalIEs := []*ie.IE{
		ie.NewCause(ie.CauseRequestAccepted),
	}

	pdrIEs := processCreatedPDRs(createdPDRs, conn.n3Address)
	additionalIEs = append(additionalIEs, pdrIEs...)
	if len(removedURRs) != 0 {
		additionalIEs = append(additionalIEs, removedURRs...)
	}

	// Send SessionEstablishmentResponse
	modResp := message.NewSessionModificationResponse(0, 0, session.RemoteSEID, req.Sequence(), 0, additionalIEs...)
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
		log.Warn().Err(err).Msg("Unknown error")
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

func composeFarInfo(far *ie.IE, farInfo ebpf.FarInfo) (ebpf.FarInfo, error) {
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
			log.Warn().Msg("No OuterHeaderCreation")
		} else {
			outerHeaderCreation, _ := forward[outerHeaderCreationIndex].OuterHeaderCreation()
			farInfo.OuterHeaderCreation = uint8(outerHeaderCreation.OuterHeaderCreationDescription >> 8)

			farInfo.Teid = outerHeaderCreation.TEID
			if outerHeaderCreation.HasIPv4() {
				farInfo.RemoteIP = binary.LittleEndian.Uint32(outerHeaderCreation.IPv4Address)
			}
			if outerHeaderCreation.HasIPv6() {
				log.Warn().Msg("IPv6 not supported yet, ignoring")
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

// TODO: add making or updating UrrInfo
func updateUrr(urrInfo *ebpf.UrrInfo, urr *ie.IE) {

	// if urr.HasVOLUM() {
	// }

	// if volumeThreshold, err := urr.VolumeThreshold(); err == nil {
	// }
}
