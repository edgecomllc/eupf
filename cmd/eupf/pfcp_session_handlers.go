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
	log.Printf("Got Session Establishment Request from: %s.", addr)
	_, remoteSEID, err := validateRequest(req.NodeID, req.CPFSEID)
	if err != nil {
		log.Printf("Rejecting Session Establishment Request from: %s (missing NodeID or F-SEID)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEMissing)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, convertErrorToIeCause(err)), nil
	}

	association, ok := conn.nodeAssociations[addr.String()]
	if !ok {
		log.Printf("Rejecting Session Establishment Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return message.NewSessionEstablishmentResponse(0, 0, 0, req.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation)), nil
	}

	localSEID := association.NewLocalSEID()

	session := Session{
		LocalSEID:    localSEID,
		RemoteSEID:   remoteSEID.SEID,
		UplinkPDRs:   map[uint32]SPDRInfo{},
		DownlinkPDRs: map[uint32]SPDRInfo{},
		FARs:         map[uint32]FarInfo{},
		QERs:         map[uint32]QerInfo{},
	}

	printSessionEstablishmentRequest(req)
	// #TODO: Implement rollback on error
	err = func() error {
		mapOperations := conn.mapOperations
		for _, far := range req.CreateFAR {
			farInfo, err := composeFarInfo(far, conn.n3Address.To4())
			if err != nil {
				log.Printf("Error extracting FAR info: %s", err)
				continue
			}

			farid, _ := far.FARID()
			log.Printf("Saving FAR info to session: %d, %+v", farid, farInfo)
			session.PutFAR(farid, farInfo)
			if err := mapOperations.PutFar(farid, farInfo); err != nil {
				log.Printf("Can't put FAR: %s", err)
			}
		}

		for _, pdr := range req.CreatePDR {
			spdrInfo, pdrId, pdi, err := composeSPDRInfo(pdr)
			if err != nil {
				return err
			}
			srcIfacePdiId := findIEindex(pdi, 20) // IE Type source interface
			srcInterface, _ := pdi[srcIfacePdiId].SourceInterface()
			switch srcInterface {
			case ie.SrcInterfaceAccess, ie.SrcInterfaceCPFunction:
				{
					if err := applyUplinkPDR(pdi, spdrInfo, pdrId, session, mapOperations); err != nil {
						log.Printf("Errored while applying PDR: %s", err)
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
						log.Printf("Errored[ while applying PDR: %s", err)
						return err
					}
				}
			default:
				log.Printf("WARN: Unsupported Source Interface type: %d", srcInterface)
			}
		}

		for _, qer := range req.CreateQER {
			qerInfo := QerInfo{}
			qerId, err := qer.QERID() // Probably will be used as ebpf map key
			if err != nil {
				return fmt.Errorf("QER ID missing")
			}

			gateStatusDL, err := qer.GateStatusDL()
			if err != nil {
				return fmt.Errorf("gate Status DL missing")
			}
			qerInfo.GateStatusDL = gateStatusDL

			gateStatusUL, err := qer.GateStatusUL()
			if err != nil {
				return fmt.Errorf("Gate Status UL missing")
			}
			qerInfo.GateStatusUL = gateStatusUL

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
			log.Printf("Saving QER info to session: %d, %+v", qerId, qerInfo)
			session.PutQER(qerId, qerInfo)
			log.Printf("Creating QER ID: %d, QER Info: %+v", qerId, qerInfo)
			if err := mapOperations.PutQer(qerId, qerInfo); err != nil {
				log.Printf("Can't put QER: %s", err)
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
	PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRequestAccepted)).Inc()
	return estResp, nil
}

func handlePfcpSessionDeletionRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) (message.Message, error) {
	req := msg.(*message.SessionDeletionRequest)
	log.Printf("Got Session Deletion Request from: %s. \n", addr)
	association, ok := conn.nodeAssociations[addr.String()]
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

func handlePfcpSessionModificationRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) (message.Message, error) {
	req := msg.(*message.SessionModificationRequest)
	log.Printf("Got Session Modification Request from: %s. \n", addr)

	log.Printf("Finding association for %s", addr)
	association, ok := conn.nodeAssociations[addr.String()]
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

			association.Sessions[req.SEID()] = session         // FIXME
			conn.nodeAssociations[addr.String()] = association // FIXME
		}
	}

	printSessionModificationRequest(req)

	// #TODO: Implement rollback on error
	err := func() error {
		mapOperations := conn.mapOperations
		for _, far := range req.UpdateFAR {
			farInfo, err := composeFarInfo(far, conn.n3Address.To4())
			if err != nil {
				log.Printf("Error extracting FAR info: %s", err)
				continue
			}

			farid, _ := far.FARID()
			log.Printf("Updating FAR info: %d, %+v", farid, farInfo)
			session.PutFAR(farid, farInfo)
			if err := mapOperations.UpdateFar(farid, farInfo); err != nil {
				log.Printf("Can't update FAR: %s", err)
			}
		}

		for _, removeFar := range req.RemoveFAR {
			farid, _ := removeFar.FARID()
			log.Printf("Removing FAR: %d", farid)
			session.RemoveFAR(farid)
			if err := mapOperations.DeleteFar(farid); err != nil {
				log.Printf("Can't remove FAR: %s", err)
			}
		}

		for _, pdr := range req.RemovePDR {
			pdrId, _ := pdr.PDRID()
			if _, ok := session.UplinkPDRs[uint32(pdrId)]; ok {
				log.Printf("Removing uplink PDR: %d", pdrId)
				session.RemoveUplinkPDR(pdrId)
				if err := mapOperations.DeletePdrUpLink(session.UplinkPDRs[uint32(pdrId)].Teid); err != nil {
					log.Printf("Failed to remove uplink PDR: %v", err)
				}
			}
			if _, ok := session.DownlinkPDRs[uint32(pdrId)]; ok {
				log.Printf("Removing downlink PDR: %d", pdrId)
				session.RemoveDownlinkPDR(pdrId)
				if err := mapOperations.DeletePdrDownLink(session.DownlinkPDRs[uint32(pdrId)].Ipv4); err != nil {
					log.Printf("Failed to remove downlink PDR: %v", err)
				}
			}
		}

		for _, qer := range req.RemoveQER {
			qerId, err := qer.QERID()
			if err != nil {
				return fmt.Errorf("QER ID missing")
			}
			log.Printf("Removing QER ID: %d", qerId)
			session.RemoveQER(qerId)
			log.Printf("Removing QER ID: %d", qerId)
			if err := mapOperations.DeleteQer(qerId); err != nil {
				log.Printf("Can't remove QER: %s", err)
			}
		}

		for _, pdr := range req.UpdatePDR {
			spdrInfo, pdrId, pdi, err := composeSPDRInfo(pdr)
			if err != nil {
				return err
			}
			srcIfacePdiId := findIEindex(pdi, 20) // IE Type source interface
			srcInterface, _ := pdi[srcIfacePdiId].SourceInterface()
			switch srcInterface {
			case ie.SrcInterfaceAccess, ie.SrcInterfaceCPFunction:
				{
					if err := applyUplinkPDR(pdi, spdrInfo, pdrId, session, mapOperations); err != nil {
						log.Printf("Errored while applying PDR: %s", err)
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
						log.Printf("Errored while applying PDR: %s", err)
						return err
					}
				}
			default:
				log.Printf("WARN: Unsupported Source Interface type: %d", srcInterface)
			}
		}

		for _, qer := range req.UpdateQER {
			qerInfo := QerInfo{}
			qerId, err := qer.QERID() // Probably will be used as ebpf map key
			if err != nil {
				return fmt.Errorf("QER ID missing")
			}

			gateStatusDL, err := qer.GateStatusDL()
			if err != nil {
				return fmt.Errorf("Gate Status DL missing")
			}
			qerInfo.GateStatusDL = gateStatusDL

			gateStatusUL, err := qer.GateStatusUL()
			if err != nil {
				return fmt.Errorf("Gate Status UL missing")
			}
			qerInfo.GateStatusUL = gateStatusUL

			maxBitrateDL, err := qer.MBRDL()
			if err != nil {
				return fmt.Errorf("Max Bitrate DL missing")
			}
			qerInfo.MaxBitrateDL = uint32(maxBitrateDL) * 1000

			maxBitrateUL, err := qer.MBRUL()
			if err != nil {
				return fmt.Errorf("Max Bitrate UL missing")
			}
			qerInfo.MaxBitrateUL = uint32(maxBitrateUL) * 1000

			qfi, err := qer.QFI()
			if err != nil {
				return fmt.Errorf("QFI missing")
			}
			qerInfo.Qfi = qfi
			qerInfo.StartUL = 0
			qerInfo.StartDL = 0

			log.Printf("Updating QER ID: %d, QER Info: %+v", qerId, qerInfo)
			session.PutQER(qerId, qerInfo)
			if err := mapOperations.UpdateQer(qerId, qerInfo); err != nil {
				log.Printf("Can't update QER: %s", err)
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

func validateRequest(nodeId *ie.IE, cpfseid *ie.IE) (string, *ie.FSEIDFields, error) {
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

func applyUplinkPDR(pdi []*ie.IE, spdrInfo SPDRInfo, pdrId uint16, session Session, mapOperations ForwardingPlaneController) error {
	// IE Type F-TEID
	if teidPdiId := findIEindex(pdi, 21); teidPdiId != -1 {
		if fteid, err := pdi[teidPdiId].FTEID(); err == nil {
			spdrInfo.Teid = fteid.TEID
			log.Printf("Saving uplink PDR info to session: %d, %+v", pdrId, spdrInfo)
			session.PutUplinkPDR(pdrId, spdrInfo)
			if err := mapOperations.PutPdrUpLink(spdrInfo.Teid, spdrInfo.PdrInfo); err != nil {
				log.Printf("Can't put uplink PDR: %s", err)
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

func applyDownlinkPDR(pdi []*ie.IE, spdrInfo SPDRInfo, pdrId uint16, session Session, mapOperations ForwardingPlaneController) error {
	// IE Type UE IP Address
	if ueipPdiId := findIEindex(pdi, 93); ueipPdiId != -1 {
		ueIp, _ := pdi[ueipPdiId].UEIPAddress()
		if ueIp.IPv4Address != nil {
			spdrInfo.Ipv4 = ueIp.IPv4Address
		} else {
			log.Print("WARN: No IPv4 address")
		}
		if ueIp.IPv6Address != nil {
			log.Print("WARN: UE IPv6 not supported yet, ignoring")
			return fmt.Errorf("IPv6 not supported")
		}
		log.Printf("Saving downlink PDR info to session: %d, %+v", pdrId, spdrInfo)
		session.PutDownlinkPDR(pdrId, spdrInfo)
		if err := mapOperations.PutPdrDownLink(spdrInfo.Ipv4, spdrInfo.PdrInfo); err != nil {
			log.Printf("Can't put uplink PDR: %s", err)
		}
	} else {
		log.Println("UE IP Address IE missing")
	}
	return nil
}

func composeFarInfo(far *ie.IE, localIp net.IP) (FarInfo, error) {
	farInfo := FarInfo{}
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
		return FarInfo{}, fmt.Errorf("unsupported IE type")
	}
	if err == nil {
		outerHeaderCreationIndex := findIEindex(forward, 84) // IE Type Outer Header Creation
		if outerHeaderCreationIndex == -1 {
			log.Println("WARN: No OuterHeaderCreation")
		} else {
			outerHeaderCreation, _ := forward[outerHeaderCreationIndex].OuterHeaderCreation()
			farInfo.OuterHeaderCreation = 1
			farInfo.Teid = outerHeaderCreation.TEID
			if outerHeaderCreation.HasIPv4() {
				farInfo.RemoteIP = binary.LittleEndian.Uint32(outerHeaderCreation.IPv4Address)
				farInfo.LocalIP = binary.LittleEndian.Uint32(localIp)
			}
			if outerHeaderCreation.HasIPv6() {
				log.Print("WARN: IPv6 not supported yet, ignoring")
				return FarInfo{}, fmt.Errorf("IPv6 not supported yet")
			}
		}
	}
	return farInfo, nil
}

func composeSPDRInfo(pdr *ie.IE) (SPDRInfo, uint16, []*ie.IE, error) {
	spdrInfo := SPDRInfo{}
	pdrId, err := pdr.PDRID()
	if err != nil {
		return SPDRInfo{}, 0, nil, fmt.Errorf("PDR ID missing")
	}
	if outerHeaderRemoval, err := pdr.OuterHeaderRemovalDescription(); err == nil {
		spdrInfo.PdrInfo.OuterHeaderRemoval = outerHeaderRemoval
	}
	if farid, err := pdr.FARID(); err == nil {
		spdrInfo.PdrInfo.FarId = farid
	}
	if qerid, err := pdr.QERID(); err == nil {
		spdrInfo.PdrInfo.QerId = qerid
	}
	pdi, err := pdr.PDI()
	if err != nil {
		log.Print(err)
		return SPDRInfo{}, 0, nil, err
	}
	return spdrInfo, pdrId, pdi, nil
}
