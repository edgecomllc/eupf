package core

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"reflect"

	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/edgecomllc/eupf/cmd/ebpf"

	"golang.org/x/exp/slices"

	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

var errMandatoryIeMissing = fmt.Errorf("mandatory IE missing")
var errNoEstablishedAssociation = fmt.Errorf("no established association")

func getNetworkInstances(req *message.SessionEstablishmentRequest) []string {
	networkInstances := []string{}
	for _, pdr := range req.CreatePDR {
		if pdi, err := pdr.PDI(); err == nil {
			for _, x := range pdi {
				if ne, err := x.NetworkInstance(); err == nil {
					networkInstances = append(networkInstances, ne)
				}
			}
		}

	}
	return networkInstances
}

func HandlePfcpSessionEstablishmentRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, bool, error) {
	req := msg.(*message.SessionEstablishmentRequest)
	log.Debug().Msgf("Got Session Establishment Request from: %s.", addr)
	printSessionEstablishmentRequest(req)

	imsi, msisdn := getSubscriberData(req.IEs)
	isTraced := conn.NeedSessionTrace(imsi, msisdn)

	remoteSEID, err := validateRequest(req.NodeID, req.CPFSEID)
	if err != nil {
		log.Warn().Msgf("Rejecting Session Establishment Request from: %s (missing NodeID or F-SEID)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEMissing)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, newIeNodeID(conn.nodeId), convertErrorToIeCause(err)), isTraced, nil
	}

	association, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Warn().Msgf("Rejecting Session Establishment Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, newIeNodeID(conn.nodeId), ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), isTraced, nil
	}

	localSEID := association.NewLocalSEID()
	session := NewSession(localSEID, remoteSEID.SEID, imsi, msisdn, isTraced)
	// #TODO: Implement rollback on error

	// While rollback is still missing
	networkInstances := getNetworkInstances(req)
	for _, ne := range networkInstances {
		if !conn.ValidateNetworkInstance(ne) {
			log.Warn().Msgf("Rejecting Session Establishment Request from: %s (Not allowed network instance)", addr)
			PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
			return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, newIeNodeID(conn.nodeId), ie.NewCause(ie.CauseRuleCreationModificationFailure)), isTraced, nil
		}
	}

	createdPDRs := []SPDRInfo{}
	pdrContext := NewPDRCreationContext(session, conn.ResourceManager)

	err = func() error {
		mapOperations := conn.mapOperations
		for _, far := range req.CreateFAR {
			farInfo, err := composeFarInfo(far, conn.n3Address.To4(), conn.n9Address.To4(), ebpf.FarInfo{})
			if err != nil {
				log.Warn().Msgf("Error extracting FAR info: %s", err.Error())
				continue
			}

			farid, _ := far.FARID()
			log.Debug().Msgf("Saving FAR info to session: %d, %+v", farid, farInfo)
			if internalId, err := mapOperations.NewFar(farInfo); err == nil {
				session.NewFar(farid, internalId, farInfo)
			} else {
				log.Warn().Msgf("Can't put FAR: %s", err.Error())
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
			log.Debug().Msgf("Saving QER info to session: %d, %+v", qerId, qerInfo)
			if internalId, err := mapOperations.NewQer(qerInfo); err == nil {
				session.NewQer(qerId, internalId, qerInfo)
			} else {
				log.Warn().Msgf("Can't put QER: %s", err.Error())
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
			log.Debug().Msgf("Saving URR info to session: %d, %+v", urrId, urrInfo)
			if internalId, err := mapOperations.NewUrr(urrInfo); err == nil {
				session.NewUrr(urrId, internalId, urrInfo)
			} else {
				log.Warn().Msgf("Can't put URR: %s", err.Error())
				return err
			}
		}

		sdfFilters := make([]ebpf.SdfFilter, 0)

		for _, pdr := range req.CreatePDR {
			// PDR should be created last, because we need to reference FARs and QERs global id
			pdrId, err := pdr.PDRID()
			if err != nil {
				continue
			}

			spdrInfo := SPDRInfo{
				PdrID:   uint32(pdrId),
				PdrInfo: ebpf.PdrInfo{TraceFlag: isTraced},
			}

			if err := pdrContext.extractPDR(pdr, &spdrInfo); err == nil {
				if spdrInfo.PCCInfo != nil {
					rule, ok := conn.PCCRules[spdrInfo.PCCInfo.PCCName]
					if !ok {
						log.Warn().
							Str("pcc rule name", spdrInfo.PCCInfo.PCCName).
							Uint16("pdrID", pdrId).
							Str("request address", addr).
							Msgf("PCC rule not found")

						continue
					}

					spdrInfo.PCCInfo = &rule
					sdfFilters = append(sdfFilters, spdrInfo.PCCInfo.SDFFilter)
					log.Debug().
						Str("pcc rule name", spdrInfo.PCCInfo.PCCName).
						Uint16("pdrID", pdrId).
						Str("request address", addr).
						Msgf("PCC rule found: %+v", spdrInfo.PCCInfo)
				}

				session.PutPDR(spdrInfo.PdrID, spdrInfo)
				applyPDR(spdrInfo, mapOperations)
				createdPDRs = append(createdPDRs, spdrInfo)
			} else {
				log.Error().Msgf("error extracting PDR info: %s", err.Error())
			}
		}

		applySdfFiltersToSession(session, sdfFilters, mapOperations)

		return nil
	}()

	if err != nil {
		log.Warn().Msgf("Rejecting Session Establishment Request from: %s (error in applying IEs)", err)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, remoteSEID.SEID, req.Sequence(), 0, newIeNodeID(conn.nodeId), ie.NewCause(ie.CauseRuleCreationModificationFailure)), isTraced, nil
	}

	// Reassigning is the best I can think of for now
	association.Sessions[localSEID] = session
	conn.NodeAssociations[addr] = association

	additionalIEs := []*ie.IE{
		newIeNodeIDHuawei(conn.nodeId),
		ie.NewCause(ie.CauseRequestAccepted),
		ie.NewFSEID(localSEID, cloneIP(conn.nodeAddrV4.Addr().AsSlice()), nil),
	}

	pdrIEs := processCreatedPDRs(createdPDRs, cloneIP(conn.n3Address))
	additionalIEs = append(additionalIEs, pdrIEs...)

	// Send SessionEstablishmentResponse
	estResp := message.NewSessionEstablishmentResponse(0, 0, remoteSEID.SEID, req.Sequence(), 0, additionalIEs...)

	//fakeIP := cloneIP(conn.nodeAddrV4)
	//fakeIP[3] = fakeIP[3] - 2
	estResp.IEs = append(estResp.IEs, ie.NewFSEID(localSEID, net.IPv4(10, 169, 26, 130), nil)) //FIXME
	estResp.SetLength()

	PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRequestAccepted)).Inc()
	log.Debug().Msgf("Session Establishment Request from %s accepted. F-SEID: %#016x", addr, localSEID)
	return estResp, isTraced, nil
}

func HandlePfcpSessionDeletionRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, bool, error) {
	req := msg.(*message.SessionDeletionRequest)
	log.Debug().Msgf("Got Session Deletion Request from: %s. \n", addr)
	association, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Warn().Msgf("Rejecting Session Deletion Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), false, nil
	}
	printSessionDeleteRequest(req)

	session, ok := association.Sessions[req.SEID()]
	if !ok {
		log.Warn().Msgf("Rejecting Session Deletion Request from: %s (unknown SEID)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseSessionContextNotFound)).Inc()
		return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseSessionContextNotFound)), false, nil
	}

	traced := session.IsSessionTraced()

	deletedURRs := make([]*ie.IE, 0, len(session.URRs))
	mapOperations := conn.mapOperations
	pdrContext := NewPDRCreationContext(session, conn.ResourceManager)

	for _, pdrInfo := range session.PDRs {
		if err := pdrContext.deletePDR(pdrInfo, mapOperations); err != nil {
			PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
			return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), traced, err
		}
	}
	for _, far := range session.FARs {
		if err := mapOperations.DeleteFar(far.GlobalId); err != nil {
			PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
			return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), traced, err
		}
	}
	for _, qer := range session.QERs {
		if err := mapOperations.DeleteQer(qer.GlobalId); err != nil {
			PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
			return message.NewSessionDeletionResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), traced, err
		}
	}
	session.URRSequence += 1
	for id, urr := range session.URRs {
		prevReport := urr.UrrInfo
		err, newReport := mapOperations.DeleteUrr(urr.GlobalId)
		if err != nil {
			log.Warn().Msgf("WARN: mapOperations failed to delete URR: %d, %s", id, err.Error())
			continue
		}
		//urr.ReportSeqNumber += 1 //Huawei
		uplink := newReport.UplinkVolume - prevReport.UplinkVolume
		downlink := newReport.DownlinkVolume - prevReport.DownlinkVolume
		deletedURRs = append(deletedURRs, ie.NewUsageReportWithinSessionDeletionResponse(
			ie.NewURRID(id),
			//ie.NewURSEQN(urr.ReportSeqNumber),
			ie.NewURSEQN(session.URRSequence), //Huawei
			ie.NewUsageReportTrigger(0, 1<<3, 0),
			ie.NewEndTime(time.Now()),
			ie.NewVolumeMeasurement(0x7, uplink+downlink, uplink, downlink, 0, 0, 0),
			// CHOICE
			// urr-type
			//    enterprise-id: ---- 0x7db(2011)
			//    urr-level-type: ---- bearer(2)
			//    urr-function-type: ---- charging(1)
			//    urr-charging-type: ---- offlinepgw(3)
			ie.NewVendorSpecificIE(34000, 2011, []byte{0x02, 0x01, 0x03}),
			// CHOICE
			// bearer-sequence
			//    enterprise-id: ---- 0x7db(2011)
			//    bearer-sequence-value: ---- 0x1(1)
			ie.NewVendorSpecificIE(32843, 2011, []byte{1}),
			// CHOICE
			// private-stop-time
			//    enterprise-id: ---- 0x7db(2011)
			//    private-stop-time-value: ---- 0x000001917EEC1EEC
			ie.NewVendorSpecificIE(34010, 2011, []byte{0x00, 0x00, 0x01, 0x92, 0x1d, 0xca, 0x24, 0x8c}),
		))
	}

	additionalIEs := []*ie.IE{
		ie.NewCause(ie.CauseRequestAccepted),
	}
	if len(deletedURRs) != 0 {
		additionalIEs = append(additionalIEs, deletedURRs...)
	}

	log.Debug().Msgf("Deleting session: %d", req.SEID())
	delete(association.Sessions, req.SEID())

	conn.ReleaseResources(req.SEID())

	PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRequestAccepted)).Inc()

	delResp := message.NewSessionDeletionResponse(0, 0, session.RemoteSEID, req.Sequence(), 0, additionalIEs...)
	return delResp, traced, nil
}

func HandlePfcpSessionModificationRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, bool, error) {
	req := msg.(*message.SessionModificationRequest)
	log.Debug().Msgf("Got Session Modification Request from: %s. \n", addr)

	log.Debug().Msgf("Finding association for %s", addr)
	association, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Warn().Msgf("Rejecting Session Modification Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionModificationResponse(0, 0, req.SEID(), req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), false, nil
	}

	log.Debug().Msgf("Finding session %d", req.SEID())
	session, ok := association.Sessions[req.SEID()]
	if !ok {
		log.Warn().Msgf("Rejecting Session Modification Request from: %s (unknown SEID)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseSessionContextNotFound)).Inc()
		return message.NewSessionModificationResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseSessionContextNotFound)), false, nil
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

	traced := session.IsSessionTraced()
	imsi, msisdn := getSubscriberData(req.IEs)

	printSessionModificationRequest(req)
	// #TODO: Implement rollback on error
	createdPDRs := []SPDRInfo{}
	removedURRs := make([]*ie.IE, 0, len(req.RemoveURR))
	pdrContext := NewPDRCreationContext(session, conn.ResourceManager)

	err := func() error {
		mapOperations := conn.mapOperations

		for _, far := range req.CreateFAR {
			farInfo, err := composeFarInfo(far, conn.n3Address.To4(), conn.n9Address.To4(), ebpf.FarInfo{})
			if err != nil {
				log.Warn().Msgf("Error extracting FAR info: %s", err.Error())
				continue
			}

			farid, _ := far.FARID()
			log.Debug().Msgf("Saving FAR info to session: %d, %+v", farid, farInfo)
			if internalId, err := mapOperations.NewFar(farInfo); err == nil {
				session.NewFar(farid, internalId, farInfo)
			} else {
				log.Warn().Msgf("Can't put FAR: %s", err.Error())
				return err
			}
		}

		for _, far := range req.UpdateFAR {
			farid, err := far.FARID()
			if err != nil {
				return err
			}
			sFarInfo := session.GetFar(farid)
			sFarInfo.FarInfo, err = composeFarInfo(far, conn.n3Address.To4(), conn.n9Address.To4(), sFarInfo.FarInfo)
			if err != nil {
				log.Warn().Msgf("Error extracting FAR info: %s", err.Error())
				continue
			}
			log.Debug().Msgf("Updating FAR info: %d, %+v", farid, sFarInfo)
			session.UpdateFar(farid, sFarInfo.FarInfo)
			if err := mapOperations.UpdateFar(sFarInfo.GlobalId, sFarInfo.FarInfo); err != nil {
				log.Info().Msgf("Can't update FAR: %s", err.Error())
			}
		}

		for _, far := range req.RemoveFAR {
			farid, _ := far.FARID()
			log.Debug().Msgf("Removing FAR: %d", farid)
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
			log.Debug().Msgf("Saving QER info to session: %d, %+v", qerId, qerInfo)
			if internalId, err := mapOperations.NewQer(qerInfo); err == nil {
				session.NewQer(qerId, internalId, qerInfo)
			} else {
				log.Warn().Msgf("Can't put QER: %s", err.Error())
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
			log.Debug().Msgf("Updating QER ID: %d, QER Info: %+v", qerId, sQerInfo)
			session.UpdateQer(qerId, sQerInfo.QerInfo)
			if err := mapOperations.UpdateQer(sQerInfo.GlobalId, sQerInfo.QerInfo); err != nil {
				log.Warn().Msgf("Can't update QER: %s", err.Error())
				return err
			}
		}

		for _, qer := range req.RemoveQER {
			qerId, err := qer.QERID()
			if err != nil {
				return fmt.Errorf("QER ID missing")
			}
			log.Debug().Msgf("Removing QER ID: %d", qerId)
			sQerInfo := session.RemoveQer(qerId)
			if err := mapOperations.DeleteQer(sQerInfo.GlobalId); err != nil {
				log.Warn().Msgf("Can't remove QER: %s", err.Error())
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
			log.Debug().Msgf("Saving URR info to session: %d, %+v", urrId, urrInfo)
			if internalId, err := mapOperations.NewUrr(urrInfo); err == nil {
				session.NewUrr(urrId, internalId, urrInfo)
			} else {
				log.Warn().Msgf("Can't put URR: %s", err.Error())
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
			log.Debug().Msgf("Updating URR ID: %d, URR Info: %+v", urrId, sUrrInfo)
			session.UpdateUrr(urrId, sUrrInfo.UrrInfo)
			if err := mapOperations.UpdateUrr(sUrrInfo.GlobalId, sUrrInfo.UrrInfo); err != nil {
				log.Warn().Msgf("Can't update URR: %s", err.Error())
				return err
			}
		}

		for _, urr := range req.RemoveURR {
			urrId, err := urr.URRID()
			if err != nil {
				return fmt.Errorf("URR ID missing")
			}
			log.Debug().Msgf("Removing URR ID: %d", urrId)
			sUrrInfo := session.RemoveUrr(urrId)
			prevReport := sUrrInfo.UrrInfo
			err, newReport := mapOperations.DeleteUrr(sUrrInfo.GlobalId)
			if err != nil {
				log.Warn().Msgf("Can't remove URR: %s", err.Error())
				continue
			}

			sUrrInfo.ReportSeqNumber += 1
			uplink := newReport.UplinkVolume - prevReport.UplinkVolume
			downlink := newReport.DownlinkVolume - prevReport.DownlinkVolume
			removedURRs = append(removedURRs, ie.NewUsageReportWithinSessionModificationResponse(
				ie.NewURRID(urrId),
				ie.NewURSEQN(sUrrInfo.ReportSeqNumber),
				ie.NewUsageReportTrigger(0, 1<<3, 0),
				ie.NewEndTime(time.Now()),
				ie.NewVolumeMeasurement(0x7, uplink+downlink, uplink, downlink, 0, 0, 0),
			))
		}

		// obtain tracing flag from storage by IMSI or MSISDN
		isTraced := conn.NeedSessionTrace(imsi, msisdn)

		sdfFilters := make([]ebpf.SdfFilter, 0)

		for _, pdr := range req.CreatePDR {
			// PDR should be created last, because we need to reference FARs and QERs global id
			pdrId, err := pdr.PDRID()
			if err != nil {
				log.Warn().Msgf("PDR ID missing")
				continue
			}

			spdrInfo := SPDRInfo{
				PdrID:   uint32(pdrId),
				PdrInfo: ebpf.PdrInfo{TraceFlag: isTraced},
			}

			if err := pdrContext.extractPDR(pdr, &spdrInfo); err == nil {
				if spdrInfo.PCCInfo != nil {
					rule, ok := conn.PCCRules[spdrInfo.PCCInfo.PCCName]
					if !ok {
						log.Warn().
							Str("pcc rule name", spdrInfo.PCCInfo.PCCName).
							Uint16("pdrID", pdrId).
							Str("request address", addr).
							Msgf("PCC rule not found")

						continue
					}

					spdrInfo.PCCInfo = &rule
					sdfFilters = append(sdfFilters, spdrInfo.PCCInfo.SDFFilter)
					log.Debug().
						Str("pcc rule name", spdrInfo.PCCInfo.PCCName).
						Uint16("pdrID", pdrId).
						Str("request address", addr).
						Msgf("PCC rule found: %+v", spdrInfo.PCCInfo)
				}

				session.PutPDR(spdrInfo.PdrID, spdrInfo)
				applyPDR(spdrInfo, mapOperations)
				createdPDRs = append(createdPDRs, spdrInfo)
			} else {
				log.Info().Msgf("Error extracting PDR info: %s", err.Error())
			}
		}

		applySdfFiltersToSession(session, sdfFilters, mapOperations)

		delSDFFilters := make([]ebpf.SdfFilter, 0)
		sdfFilters = make([]ebpf.SdfFilter, 0)

		for _, pdr := range req.UpdatePDR {
			pdrId, err := pdr.PDRID()
			if err != nil {
				return fmt.Errorf("PDR ID missing")
			}

			spdrInfo := session.GetPDR(pdrId)
			spdrInfo.PdrInfo.TraceFlag = isTraced

			spdrInfoOld := spdrInfo

			if err := pdrContext.extractPDR(pdr, &spdrInfo); err == nil {
				if spdrInfoOld.PCCInfo != nil && spdrInfo.PCCInfo == nil {
					if err := pdrContext.deletePDR(spdrInfoOld, mapOperations); err != nil {
						log.Info().Msgf("Failed to remove uplink PDR: %v", err)
					}

					delSDFFilters = append(delSDFFilters, spdrInfoOld.PCCInfo.SDFFilter)

					continue
				} else if spdrInfoOld.PCCInfo == nil && spdrInfo.PCCInfo != nil {
					rule, ok := conn.PCCRules[spdrInfo.PCCInfo.PCCName]
					if !ok {
						log.Warn().
							Str("pcc rule name", spdrInfo.PCCInfo.PCCName).
							Uint16("pdrID", pdrId).
							Str("request address", addr).
							Msgf("PCC rule not found")

						continue
					}

					spdrInfo.PCCInfo = &rule
					sdfFilters = append(sdfFilters, spdrInfo.PCCInfo.SDFFilter)
					log.Debug().
						Str("pcc rule name", spdrInfo.PCCInfo.PCCName).
						Uint16("pdrID", pdrId).
						Str("request address", addr).
						Msgf("PCC rule found: %+v", spdrInfo.PCCInfo)
				} else if spdrInfoOld.PCCInfo != nil {
					continue
				}

				session.PutPDR(uint32(pdrId), spdrInfo)
				applyPDR(spdrInfo, mapOperations)
			} else {
				log.Info().Msgf("Error extracting PDR info: %s", err.Error())
			}
		}

		removeSdfFiltersFromSession(session, delSDFFilters, mapOperations)
		applySdfFiltersToSession(session, sdfFilters, mapOperations)

		for _, pdr := range req.RemovePDR {
			pdrId, _ := pdr.PDRID()
			if _, ok := session.PDRs[uint32(pdrId)]; ok {
				log.Info().Msgf("Removing uplink PDR: %d", pdrId)
				sPDRInfo := session.RemovePDR(uint32(pdrId))

				if err := pdrContext.deletePDR(sPDRInfo, mapOperations); err != nil {
					log.Info().Msgf("Failed to remove uplink PDR: %v", err)
				}
			}
		}

		return nil
	}()
	if err != nil {
		log.Warn().Msgf("Rejecting Session Modification Request from: %s (failed to apply rules)", err)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRuleCreationModificationFailure)).Inc()
		return message.NewSessionModificationResponse(0, 0, session.RemoteSEID, req.Sequence(), 0, ie.NewCause(ie.CauseRuleCreationModificationFailure)), traced, nil
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
	return modResp, traced, nil
}

func applySdfFiltersToSession(session *Session, sdfFilters []ebpf.SdfFilter, mapOps ebpf.ForwardingPlaneController) {
	if len(sdfFilters) == 0 {
		return
	}

	for _, spdrInfo := range session.PDRs {
		log.Info().Msgf("Applying SDF Filters: %+v", spdrInfo)
		if spdrInfo.PCCInfo != nil || (spdrInfo.Teid == 0 && spdrInfo.Ipv4 == nil && spdrInfo.Ipv6 == nil) {
			continue
		}

		if spdrInfo.PdrInfo.SdfFilter == nil {
			spdrInfo.PdrInfo.SdfFilter = make([]ebpf.SdfFilter, 0)
		}

		for _, filter := range sdfFilters {
			if len(spdrInfo.PdrInfo.SdfFilter) >= 2 {
				log.Warn().Uint32("pdrID", spdrInfo.PdrID).Msg("sdfFilter is full")
				break
			}
			spdrInfo.PdrInfo.SdfFilter = append(spdrInfo.PdrInfo.SdfFilter, filter)
			spdrInfo.PdrInfo.NotifyFlag = true //FIXME
		}

		session.PutPDR(spdrInfo.PdrID, spdrInfo)
		updatePdrInMap(spdrInfo, mapOps)
	}
}

func removeSdfFiltersFromSession(session *Session, delFilters []ebpf.SdfFilter, mapOps ebpf.ForwardingPlaneController) {
	if len(delFilters) == 0 {
		return
	}

	filterSet := make(map[string]struct{})
	for _, filter := range delFilters {
		filterSet[serializeSdfFilter(filter)] = struct{}{}
	}

	for _, spdrInfo := range session.PDRs {
		if spdrInfo.PCCInfo != nil ||
			(spdrInfo.Teid == 0 && spdrInfo.Ipv4 == nil && spdrInfo.Ipv6 == nil) ||
			spdrInfo.PdrInfo.SdfFilter == nil {
			continue
		}

		newFilters := make([]ebpf.SdfFilter, 0)
		for _, existing := range spdrInfo.PdrInfo.SdfFilter {
			if _, found := filterSet[serializeSdfFilter(existing)]; !found {
				newFilters = append(newFilters, existing)
			}
		}

		spdrInfo.PdrInfo.SdfFilter = newFilters
		session.PutPDR(spdrInfo.PdrID, spdrInfo)
		updatePdrInMap(spdrInfo, mapOps)
	}
}

func updatePdrInMap(spdrInfo SPDRInfo, mapOps ebpf.ForwardingPlaneController) {

	if spdrInfo.Ipv4 != nil {
		if err := mapOps.UpdatePdrDownlink(spdrInfo.Ipv4, spdrInfo.PdrInfo); err != nil {
			log.Info().Msgf("Can't update IPv4 PDR: %s", err)
		}
	}

	if spdrInfo.Ipv6 != nil {
		if err := mapOps.UpdateDownlinkPdrIp6(spdrInfo.Ipv6, spdrInfo.PdrInfo); err != nil {
			log.Info().Msgf("Can't update IPv6 PDR: %s", err)
		}
	}

	if spdrInfo.Teid > 0 {
		if err := mapOps.UpdatePdrUplink(spdrInfo.Teid, spdrInfo.PdrInfo); err != nil {
			log.Info().Msgf("Can't update GTP PDR: %s", err)
		}
	}
}

func serializeSdfFilter(filter ebpf.SdfFilter) string {
	t := reflect.TypeOf(filter)
	v := reflect.ValueOf(filter)

	result := ""
	for i := 0; i < t.NumField(); i++ {
		result += fmt.Sprintf("%v|", v.Field(i).Interface())
	}
	return result
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

func findEnterpriseSpecificIEindex(ieArr []*ie.IE, ieType uint16, ieEnterpriseID uint16) int {
	arrIndex := slices.IndexFunc(ieArr, func(ie *ie.IE) bool {
		return ie.Type == ieType && ie.EnterpriseID == ieEnterpriseID
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

func composeFarInfo(far *ie.IE, localN3Ip net.IP, localN9Ip net.IP, farInfo ebpf.FarInfo) (ebpf.FarInfo, error) {
	if applyAction, err := far.ApplyAction(); err == nil {
		farInfo.Action = applyAction[0]
	}
	var forward []*ie.IE
	var err error
	if far.Type == ie.CreateFAR {
		farInfo.LocalIP = binary.LittleEndian.Uint32(localN3Ip)
		forward, err = far.ForwardingParameters()
	} else if far.Type == ie.UpdateFAR {
		forward, err = far.UpdateForwardingParameters()
	} else {
		return ebpf.FarInfo{}, fmt.Errorf("unsupported IE type")
	}
	if err == nil {
		outerHeaderCreationIndex := findIEindex(forward, 84) // IE Type Outer Header Creation
		if outerHeaderCreationIndex == -1 {
			log.Debug().Msg("No OuterHeaderCreation")
		} else {

			if config.Conf.HuaweiSupport {
				huaweiOuterHeaderCreation, err := HuaweiOuterHeaderCreation(forward[outerHeaderCreationIndex]) //Huawei
				if err != nil {
					log.Error().Msgf("Error creating OuterHeaderCreation: %s", err.Error())
					return ebpf.FarInfo{}, err
				}

				farInfo.OuterHeaderCreation = uint8(1 << huaweiOuterHeaderCreation.OuterHeaderCreationDescription) //Huawei
				farInfo.Teid = huaweiOuterHeaderCreation.TEID
				if huaweiOuterHeaderCreation.HasIPv4() {
					farInfo.RemoteIP = binary.LittleEndian.Uint32(huaweiOuterHeaderCreation.IPv4Address)
				}
				if huaweiOuterHeaderCreation.HasIPv6() {
					log.Warn().Msg("IPv6 not supported yet, ignoring")
					return ebpf.FarInfo{}, fmt.Errorf("IPv6 not supported yet")
				}
			} else {
				outerHeaderCreation, err := forward[outerHeaderCreationIndex].OuterHeaderCreation()
				if err != nil {
					log.Error().Msgf("Error creating OuterHeaderCreation: %s", err.Error())

					return ebpf.FarInfo{}, err
				}

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

			destInterfaceIndex := findIEindex(forward, 42) // IE Destination Interface
			if destInterfaceIndex == -1 {
				log.Warn().Msg("No Destination Interface IE")
			} else {
				destInterface, _ := forward[destInterfaceIndex].DestinationInterface()
				// OuterHeaderCreation == GTP-U/UDP/IPv4 && DestinationInterface == Core:
				if (farInfo.OuterHeaderCreation&0x01) == 0x01 && (destInterface == 0x01) {
					farInfo.LocalIP = binary.LittleEndian.Uint32(localN9Ip)
				} else {
					farInfo.LocalIP = binary.LittleEndian.Uint32(localN3Ip)
				}
			}
		}
	}

	if transportLevelMarking, err := GetTransportLevelMarking(far); err == nil {
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
		qerInfo.MaxBitrateDL = maxBitrateDL * 1000
	}
	maxBitrateUL, err := qer.MBRUL()
	if err == nil {
		qerInfo.MaxBitrateUL = maxBitrateUL * 1000
	}
	if qfi, err := qer.QFI(); err == nil {
		qerInfo.Qfi = qfi
	} else if qfiId := findEnterpriseSpecificIEindex(qer.ChildIEs, 32785, 2011); qfiId != -1 { // IE Huawei QCI
		qfi := qer.ChildIEs[qfiId].Payload[0]
		qerInfo.Qfi = qfi
	}

	qerInfo.Dscp = config.Conf.GetDscpMarkByQci(qerInfo.Qfi)
}

func GetTransportLevelMarking(far *ie.IE) (uint16, error) {
	for _, informationalElement := range far.ChildIEs {
		if informationalElement.Type == ie.TransportLevelMarking {
			return informationalElement.TransportLevelMarking()
		}
	}
	return 0, fmt.Errorf("no TransportLevelMarking found")
}

func updateUrr(urrInfo *ebpf.UrrInfo, urr *ie.IE) {

	if urr.HasVOLUM() {
		if volumeThreshold, err := urr.VolumeThreshold(); err == nil {
			if volumeThreshold.HasTOVOL() {
				urrInfo.VolumeThreshold = volumeThreshold.TotalVolume
			}
		}
	}

	if urr.HasDURAT() {
		if timeThreshold, err := urr.TimeThreshold(); err == nil {
			urrInfo.TimeThreshold = timeThreshold.Seconds()
		}
	}

}

func getSubscriberData(ieArr []*ie.IE) (string, string) {
	var imsi, msisdn string

	imsiIdx := findEnterpriseSpecificIEindex(ieArr, 32769, 2011) // IE Huawei IMSI
	if imsiIdx != -1 {
		imsiEncoded := ieArr[imsiIdx].Payload
		imsi = DecodeDigitsFromBytes(imsiEncoded)
	}

	msisdnIdx := findEnterpriseSpecificIEindex(ieArr, 32770, 2011) // IE Huawei MSISDN
	if msisdnIdx != -1 {
		msisdnEncoded := ieArr[msisdnIdx].Payload
		msisdn = DecodeDigitsFromBytes(msisdnEncoded)
	}

	return imsi, msisdn
}
