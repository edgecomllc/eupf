package core

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func writeLineTabbed(sb *strings.Builder, s string, tab int) {
	sb.WriteString(strings.Repeat("  ", tab))
	sb.WriteString(s)
	sb.WriteString("\n")
}

func printAssociationSetupRequest(req *message.AssociationSetupRequest) {
	var sb strings.Builder
	sb.WriteString("\n")
	writeLineTabbed(&sb, "Association Setup Request:", 0)
	nodeId, err := req.NodeID.NodeID()
	if err == nil {
		writeLineTabbed(&sb, fmt.Sprintf("Node ID: %s", nodeId), 1)
	}
	if req.RecoveryTimeStamp != nil {
		recoveryTime, err := req.RecoveryTimeStamp.RecoveryTimeStamp()
		if err == nil {
			writeLineTabbed(&sb, fmt.Sprintf("Recovery Time: %s", recoveryTime.String()), 1)
		}
	}
	log.Info().Msg(sb.String())
}

func printAssociationUpdateRequest(req *message.AssociationUpdateRequest) {
	var sb strings.Builder
	sb.WriteString("\n")
	writeLineTabbed(&sb, "Association Update Request:", 0)
	nodeId, err := req.NodeID.NodeID()
	if err == nil {
		writeLineTabbed(&sb, fmt.Sprintf("Node ID: %s", nodeId), 1)
	}
	log.Info().Msg(sb.String())
}

func GetFSEID(CPFSEID *ie.IE) uint64 {
	if CPFSEID != nil {
		if fseid, err := CPFSEID.FSEID(); err == nil {
			return fseid.SEID
		}
	}
	return 0
}

func printSessionEstablishmentRequest(req *message.SessionEstablishmentRequest) {
	var sb strings.Builder
	sb.WriteString("\n")
	writeLineTabbed(&sb, fmt.Sprintf("%s( SEID: %#016x, F-SEID: %#016x ):", req.MessageTypeName(), req.SEID(), GetFSEID(req.CPFSEID)), 0)

	for _, pdr := range req.CreatePDR {
		sb.WriteString("  Create")
		displayPdr(&sb, pdr)
	}

	for _, far := range req.CreateFAR {
		sb.WriteString("  Create")
		displayFar(&sb, far)
	}

	for _, qer := range req.CreateQER {
		sb.WriteString("  Create")
		displayQer(&sb, qer)
	}

	for _, urr := range req.CreateURR {
		sb.WriteString("  Create")
		displayUrr(&sb, urr)
	}

	if req.CreateBAR != nil {
		sb.WriteString("  Create")
		displayBar(&sb, req.CreateBAR)
	}

	if imsiId := findEnterpriseSpecificIEindex(req.IEs, 32769, 2011); imsiId != -1 { // IE Huawei IMSI
		imsiEncoded := req.IEs[imsiId].Payload
		writeLineTabbed(&sb, fmt.Sprintf("IMSI: %s ", DecodeDigitsFromBytes(imsiEncoded)), 1)
	}

	if msisdnId := findEnterpriseSpecificIEindex(req.IEs, 32770, 2011); msisdnId != -1 { // IE Huawei MSISDN
		msisdnEncoded := req.IEs[msisdnId].Payload
		writeLineTabbed(&sb, fmt.Sprintf("MSISDN: %s ", DecodeDigitsFromBytes(msisdnEncoded)), 1)
	}

	if req.UserID != nil {
		if userID, err := req.UserID.UserID(); err == nil {
			if (userID.Flags & 0x01) == 0x01 {
				writeLineTabbed(&sb, fmt.Sprintf("IMSI: %s ", userID.IMSI), 1)
			}

			if (userID.Flags & 0x02) == 0x02 {
				writeLineTabbed(&sb, fmt.Sprintf("IMEI: %s ", userID.IMEI), 1)
			}

			if (userID.Flags & 0x04) == 0x04 {
				writeLineTabbed(&sb, fmt.Sprintf("MSISDN: %s ", userID.MSISDN), 1)
			}
		}
	}

	if req.APNDNN != nil {
		if apn, err := req.APNDNN.APNDNN(); err == nil {
			writeLineTabbed(&sb, fmt.Sprintf("APN/DNN: %s ", apn), 1)
		}
	}

	if req.SNSSAI != nil {
		if snssai, err := req.SNSSAI.SNSSAI(); err == nil && len(snssai) == 4 {
			writeLineTabbed(&sb, fmt.Sprintf("S-NSSAI: SST: %d SD: %x ", snssai[0], snssai[1:3]), 1)
		}
	}

	log.Debug().Msg(sb.String())
}

// IE Contents of Create/Update/Remove are mostly the same
func printSessionModificationRequest(req *message.SessionModificationRequest) {
	var sb strings.Builder
	sb.WriteString("\n")
	writeLineTabbed(&sb, fmt.Sprintf("%s( SEID: %#016x, F-SEID: %#016x ):", req.MessageTypeName(), req.SEID(), GetFSEID(req.CPFSEID)), 0)
	for _, pdr := range req.CreatePDR {
		sb.WriteString("  Create")
		displayPdr(&sb, pdr)
	}

	for _, far := range req.CreateFAR {
		sb.WriteString("  Create")
		displayFar(&sb, far)
	}

	for _, qer := range req.CreateQER {
		sb.WriteString("  Create")
		displayQer(&sb, qer)
	}

	for _, urr := range req.CreateURR {
		sb.WriteString("  Create")
		displayUrr(&sb, urr)
	}

	if req.CreateBAR != nil {
		sb.WriteString("  Create")
		displayBar(&sb, req.CreateBAR)
	}

	for _, pdr := range req.UpdatePDR {
		sb.WriteString("  Update")
		displayPdr(&sb, pdr)
	}

	for _, far := range req.UpdateFAR {
		sb.WriteString("  Update")
		displayFar(&sb, far)
	}

	for _, qer := range req.UpdateQER {
		sb.WriteString("  Update")
		displayQer(&sb, qer)
	}

	for _, urr := range req.UpdateURR {
		sb.WriteString("  Update")
		displayUrr(&sb, urr)
	}

	if req.UpdateBAR != nil {
		writeLineTabbed(&sb, "Update BAR:", 1)
		barId, err := req.UpdateBAR.BARID()
		if err == nil {
			writeLineTabbed(&sb, fmt.Sprintf("BAR ID: %d ", barId), 2)
		}
		downlink, err := req.UpdateBAR.DownlinkDataNotificationDelay()
		if err == nil {
			writeLineTabbed(&sb, fmt.Sprintf("Downlink Data Notification Delay: %s ", downlink), 2)
		}
		suggestedBufferingPackets, err := req.UpdateBAR.SuggestedBufferingPacketsCount()
		if err == nil {
			writeLineTabbed(&sb, fmt.Sprintf("Suggested Buffering Packets Count: %d ", suggestedBufferingPackets), 2)
		}
		mtEdtControl, err := req.UpdateBAR.MTEDTControlInformation()
		if err == nil {
			writeLineTabbed(&sb, fmt.Sprintf("MT EDI: %d ", mtEdtControl), 2)
		}
	}

	//log.Println("------ Remove:")
	for _, pdr := range req.RemovePDR {
		sb.WriteString("  Remove")
		displayPdr(&sb, pdr)
	}

	for _, far := range req.RemoveFAR {
		sb.WriteString("  Remove")
		displayFar(&sb, far)
	}

	for _, qer := range req.RemoveQER {
		sb.WriteString("  Remove")
		displayQer(&sb, qer)
	}

	for _, urr := range req.RemoveURR {
		sb.WriteString("  Remove")
		displayUrr(&sb, urr)
	}

	if req.RemoveBAR != nil {
		writeLineTabbed(&sb, "Remove BAR:", 1)
		barId, err := req.RemoveBAR.BARID()
		if err == nil {
			writeLineTabbed(&sb, fmt.Sprintf("BAR ID: %d ", barId), 2)
		}
	}

	if imsiId := findEnterpriseSpecificIEindex(req.IEs, 32769, 2011); imsiId != -1 { // IE Huawei IMSI
		imsiEncoded := req.IEs[imsiId].Payload
		writeLineTabbed(&sb, fmt.Sprintf("IMSI: %s ", DecodeDigitsFromBytes(imsiEncoded)), 1)
	}

	if msisdnId := findEnterpriseSpecificIEindex(req.IEs, 32770, 2011); msisdnId != -1 { // IE Huawei MSISDN
		msisdnEncoded := req.IEs[msisdnId].Payload
		writeLineTabbed(&sb, fmt.Sprintf("MSISDN: %s ", DecodeDigitsFromBytes(msisdnEncoded)), 1)
	}

	log.Debug().Msg(sb.String())
}

func printSessionDeleteRequest(req *message.SessionDeletionRequest) {
	var sb strings.Builder
	sb.WriteString("\n")
	writeLineTabbed(&sb, fmt.Sprintf("%s( SEID: %#016x, F-SEID: %#016x ):", req.MessageTypeName(), req.SEID(), 0), 0)
	log.Debug().Msg(sb.String())
}

func displayBar(sb *strings.Builder, bar *ie.IE) {
	barId, _ := bar.BARID()
	sb.WriteString(fmt.Sprintf("BAR ID: %d\n", barId))

	downlink, err := bar.DownlinkDataNotificationDelay()
	if err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Downlink Data Notification Delay: %s ", downlink), 2)
	}
	suggestedBufferingPackets, err := bar.SuggestedBufferingPacketsCount()
	if err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Suggested Buffering Packets Count: %d ", suggestedBufferingPackets), 2)
	}
	mtEdtControl, err := bar.MTEDTControlInformation()
	if err == nil {
		writeLineTabbed(sb, fmt.Sprintf("MT EDI: %d ", mtEdtControl), 2)
	}
}

func displayUrr(sb *strings.Builder, urr *ie.IE) {
	urrId, _ := urr.URRID()
	sb.WriteString(fmt.Sprintf("URR ID: %d \n", urrId))

	if measurementMethod, err := urr.MeasurementMethod(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Measurement Method: %d ", measurementMethod), 2)
	}
	if volumeThreshold, err := urr.VolumeThreshold(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Volume Threshold: %+v ", volumeThreshold), 2)
	}
	if timeThreshold, err := urr.TimeThreshold(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Time Threshold: %f s", timeThreshold.Seconds()), 2)
	}
	if monitoringTime, err := urr.MonitoringTime(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Monitoring Time: %s ", monitoringTime.Format(time.RFC3339)), 2)
	}
}

func displayQer(sb *strings.Builder, qer *ie.IE) {
	qerId, _ := qer.QERID()
	sb.WriteString(fmt.Sprintf("QER ID: %d \n", qerId))

	if gateStatusDL, err := qer.GateStatusDL(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Gate Status DL: %d ", gateStatusDL), 2)
	}
	if gateStatusUL, err := qer.GateStatusUL(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Gate Status UL: %d ", gateStatusUL), 2)
	}
	if maxBitrateDL, err := qer.MBRDL(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Max Bitrate DL: %d ", uint32(maxBitrateDL)), 2)
	}
	if maxBitrateUL, err := qer.MBRUL(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Max Bitrate UL: %d ", uint32(maxBitrateUL)), 2)
	}
	if qfi, err := qer.QFI(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("QFI: %d ", qfi), 2)
	} else if qciId := findEnterpriseSpecificIEindex(qer.ChildIEs, 32785, 2011); qciId != -1 { // IE Huawei QCI
		qfi := qer.ChildIEs[qciId].Payload[0]
		writeLineTabbed(sb, fmt.Sprintf("Huawei QFI: %d ", qfi), 2)
	}
}

func displayFar(sb *strings.Builder, far *ie.IE) {
	farId, _ := far.FARID()
	sb.WriteString(fmt.Sprintf("FAR ID: %d \n", farId))

	applyAction, err := far.ApplyAction()
	if err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Apply Action: %+v ", applyAction), 2)
	}
	if forwardingParameters, err := far.ForwardingParameters(); err == nil {
		writeLineTabbed(sb, "Forwarding Parameters:", 2)
		for _, forwardingParameter := range forwardingParameters {
			networkInstance, err := forwardingParameter.NetworkInstance()
			if err == nil {
				writeLineTabbed(sb, fmt.Sprintf("Network Instance: %s ", networkInstance), 3)
			}
			//outerHeaderCreation, err := forwardingParameter.OuterHeaderCreation()
			outerHeaderCreation, err := HuaweiOuterHeaderCreation(forwardingParameter)
			if err == nil {
				writeLineTabbed(sb, fmt.Sprintf("Outer Header Creation: %+v ", outerHeaderCreation), 3)
			}
			redirectInformation, err := forwardingParameter.RedirectInformation()
			if err == nil {
				writeLineTabbed(sb, fmt.Sprintf("Redirect Information, server address: %s ", redirectInformation.RedirectServerAddress), 3)
				writeLineTabbed(sb, fmt.Sprintf("Redirect Information, other server address: %s ", redirectInformation.OtherRedirectServerAddress), 3)
			}
			headerEnrichment, err := forwardingParameter.HeaderEnrichment()
			if err == nil {
				writeLineTabbed(sb, fmt.Sprintf("Header Enrichment: %s : %s ", headerEnrichment.HeaderFieldName, headerEnrichment.HeaderFieldValue), 3)
			}
		}
	}
	if updateForwardingParameters, err := far.UpdateForwardingParameters(); err == nil {
		writeLineTabbed(sb, "Update forwarding Parameters:", 2)
		for _, updateForwardingParameter := range updateForwardingParameters {
			if networkInstance, err := updateForwardingParameter.NetworkInstance(); err == nil {
				writeLineTabbed(sb, fmt.Sprintf("Network Instance: %s ", networkInstance), 3)
			}

			//outerHeaderCreation, err := updateForwardingParameter.OuterHeaderCreation()
			if outerHeaderCreation, err := HuaweiOuterHeaderCreation(updateForwardingParameter); err == nil {
				writeLineTabbed(sb, "Outer Header Creation:", 3)
				writeLineTabbed(sb, fmt.Sprintf("Outer Header Creation Description: %+v ", outerHeaderCreation.OuterHeaderCreationDescription), 4)
				writeLineTabbed(sb, fmt.Sprintf("TEID: %+v ", outerHeaderCreation.TEID), 4)
				writeLineTabbed(sb, fmt.Sprintf("IPv4Address: %+v ", outerHeaderCreation.IPv4Address), 4)
				writeLineTabbed(sb, fmt.Sprintf("IPv6Address: %+v ", outerHeaderCreation.IPv6Address), 4)
				writeLineTabbed(sb, fmt.Sprintf("PortNumber: %+v ", outerHeaderCreation.PortNumber), 4)
				writeLineTabbed(sb, fmt.Sprintf("CTag: %+v ", outerHeaderCreation.CTag), 4)
				writeLineTabbed(sb, fmt.Sprintf("STag: %+v ", outerHeaderCreation.STag), 4)
			}

			if redirectInformation, err := updateForwardingParameter.RedirectInformation(); err == nil {
				writeLineTabbed(sb, fmt.Sprintf("Redirect Information, server address: %s ", redirectInformation.RedirectServerAddress), 3)
				writeLineTabbed(sb, fmt.Sprintf("Redirect Information, other server address: %s ", redirectInformation.OtherRedirectServerAddress), 3)
			}

			if headerEnrichment, err := updateForwardingParameter.HeaderEnrichment(); err == nil {
				writeLineTabbed(sb, fmt.Sprintf("Header Enrichment: %s : %s ", headerEnrichment.HeaderFieldName, headerEnrichment.HeaderFieldValue), 3)
			}

			if transportLevelMarking, err := updateForwardingParameter.TransportLevelMarking(); err == nil {
				writeLineTabbed(sb, fmt.Sprintf("Transport Level Marking: %+v", transportLevelMarking), 3)
			}
		}
	}

	duplicatingParameters, err := far.DuplicatingParameters()
	if err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Duplicating Parameters: %+v ", duplicatingParameters), 2)
	}
	barId, err := far.BARID()
	if err == nil {
		writeLineTabbed(sb, fmt.Sprintf("BAR ID: %d ", barId), 2)
	}
	transportLevelMarking, err := GetTransportLevelMarking(far)
	if err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Transport Level Marking: %d", transportLevelMarking), 2)
		// DSCP (first octet) and ToS or Traffic Class mask (second octet)
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, transportLevelMarking)
		writeLineTabbed(sb, fmt.Sprintf("DSCP: %x", buf[0]), 3)
		writeLineTabbed(sb, fmt.Sprintf("ToS or Traffic Class mask: %x", buf[1]), 3)
	}
}

func displayPdr(sb *strings.Builder, pdr *ie.IE) {
	pdrId, _ := pdr.PDRID()
	sb.WriteString(fmt.Sprintf("PDR ID: %d \n", pdrId))

	if outerHeaderRemoval, err := pdr.OuterHeaderRemovalDescription(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Outer Header Removal: %d ", outerHeaderRemoval), 2)
	}

	if farid, err := pdr.FARID(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("FAR ID: %d ", farid), 2)
	}

	// No method to get several IEs in go-pfcp. So go through all child IEs
	for _, x := range pdr.ChildIEs {
		if x.Type == ie.QERID {
			if qerid, err := x.QERID(); err == nil {
				writeLineTabbed(sb, fmt.Sprintf("QER ID: %d ", qerid), 2)
			}
		}
	}

	if urrid, err := pdr.URRID(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("URR ID: %d ", urrid), 2)
	}

	if barid, err := pdr.BARID(); err == nil {
		writeLineTabbed(sb, fmt.Sprintf("BAR ID: %d ", barid), 2)
	}

	if pdi, err := pdr.PDI(); err == nil {
		for _, x := range pdi {
			switch x.Type {
			case 20: // IE Type source interface
				srcInterface, _ := x.SourceInterface()
				writeLineTabbed(sb, fmt.Sprintf("Source Interface: %d ", srcInterface), 2)
			case 21: // IE Type F-TEID
				if fteid, err := x.FTEID(); err == nil {
					writeLineTabbed(sb, fmt.Sprintf("F-TEID: TEID: %d, IPv4: %+v, IPv6: %+v ", fteid.TEID, fteid.IPv4Address, fteid.IPv6Address), 2)
				}
			case 93: // IE Type UE IP Address
				if ueIp, _ := x.UEIPAddress(); ueIp != nil {
					if ueIp.IPv4Address != nil {
						writeLineTabbed(sb, fmt.Sprintf("UE IPv4 Address: %s ", ueIp.IPv4Address), 2)
					}
					if ueIp.IPv6Address != nil {
						writeLineTabbed(sb, fmt.Sprintf("UE IPv6 Address: %s ", ueIp.IPv6Address), 2)
					}
				}
			case 22: // IE Type Network Instance
				if ne, err := x.NetworkInstance(); err == nil {
					writeLineTabbed(sb, fmt.Sprintf("Network Instance: %s ", ne), 2)
				}

			case 23: // IE Type SDF Filter
				if sdfFilter, err := x.SDFFilter(); err == nil {
					writeLineTabbed(sb, fmt.Sprintf("SDF Filter: %s ", sdfFilter.FlowDescription), 2)
				}
			}
		}
	}
}

func printSessionReportResponse(req *message.SessionReportResponse) {
	var sb strings.Builder
	sb.WriteString("\n")
	writeLineTabbed(&sb, fmt.Sprintf("%s( SEID: %#016x, F-SEID: %#016x ):", req.MessageTypeName(), req.SEID(), GetFSEID(req.CPFSEID)), 0)

	log.Debug().Msg(sb.String())
}
