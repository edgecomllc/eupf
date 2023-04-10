package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func writeLineTabbed(sb *strings.Builder, s string, tab int) {
	sb.WriteString(strings.Repeat("  ", tab))
	sb.WriteString(s)
	sb.WriteString("\n")
}

func printSessionEstablishmentRequest(req *message.SessionEstablishmentRequest) {
	var sb strings.Builder
	sb.WriteString("\n")
	writeLineTabbed(&sb, "Session Establishment Request:", 0)
	for _, pdr := range req.CreatePDR {
		displayPdr(&sb, pdr, "Create")
	}

	for _, far := range req.CreateFAR {
		displayFar(&sb, far, "Create")
	}

	for _, qer := range req.CreateQER {
		displayQer(&sb, qer, "Create")
	}

	for _, urr := range req.CreateURR {
		displayUrr(&sb, urr, "Create")
	}

	if req.CreateBAR != nil {
		displayBar(&sb, req, "Create")
	}
	log.Print(sb.String())
}

// IE Contents of Create/Update/Remove are mostly the same
func printSessionModificationRequest(req *message.SessionModificationRequest) {
	var sb strings.Builder
	sb.WriteString("\n")
	//log.Printf("Session Modification Request:")
	writeLineTabbed(&sb, "Session Modification Request:", 0)
	for _, pdr := range req.UpdatePDR {
		displayPdr(&sb, pdr, "Update")
	}

	for _, far := range req.UpdateFAR {
		displayFar(&sb, far, "Update")
	}

	for _, qer := range req.UpdateQER {
		displayQer(&sb, qer, "Update")
	}

	for _, urr := range req.UpdateURR {
		displayUrr(&sb, urr, "Update")
	}

	if req.UpdateBAR != nil {
		// sb.WriteString("------ BAR")
		writeLineTabbed(&sb, "Update BAR:", 1)
		barId, err := req.UpdateBAR.BARID()
		if err == nil {
			//sb.WriteString(fmt.Sprintf("BAR ID: %d ", barId))
			writeLineTabbed(&sb, fmt.Sprintf("BAR ID: %d ", barId), 2)
		}
		downlink, err := req.UpdateBAR.DownlinkDataNotificationDelay()
		if err == nil {
			//sb.WriteString(fmt.Sprintf("Downlink Data Notification Delay: %s ", downlink))
			writeLineTabbed(&sb, fmt.Sprintf("Downlink Data Notification Delay: %s ", downlink), 2)
		}
		suggestedBufferingPackets, err := req.UpdateBAR.SuggestedBufferingPacketsCount()
		if err == nil {
			//sb.WriteString(fmt.Sprintf("Suggested Buffering Packets Count: %d ", suggestedBufferingPackets))
			writeLineTabbed(&sb, fmt.Sprintf("Suggested Buffering Packets Count: %d ", suggestedBufferingPackets), 2)
		}
		mtEdtControl, err := req.UpdateBAR.MTEDTControlInformation()
		if err == nil {
			// sb.WriteString(fmt.Sprintf("MT EDI: %d ", mtEdtControl))
			writeLineTabbed(&sb, fmt.Sprintf("MT EDI: %d ", mtEdtControl), 2)
		}
	}

	//log.Println("------ Remove:")
	for _, pdr := range req.RemovePDR {
		displayPdr(&sb, pdr, "Remove")
	}

	for _, far := range req.RemoveFAR {
		displayFar(&sb, far, "Remove")
	}

	for _, qer := range req.RemoveQER {
		displayQer(&sb, qer, "Remove")
	}

	for _, urr := range req.RemoveURR {
		displayUrr(&sb, urr, "Remove")
	}

	if req.RemoveBAR != nil {
		writeLineTabbed(&sb, "Remove BAR:", 1)
		barId, err := req.RemoveBAR.BARID()
		if err == nil {
			writeLineTabbed(&sb, (fmt.Sprintf("BAR ID: %d ", barId)), 2)
		}
	}
	log.Print(sb.String())
}

func printSessionDeleteRequest(req *message.SessionDeletionRequest) {
	var sb strings.Builder
	sb.WriteString("\n")
	writeLineTabbed(&sb, "Session Deletion Request:", 0)
	writeLineTabbed(&sb, fmt.Sprintf("SEID: %d", req.SEID()), 1)
}

func displayBar(sb *strings.Builder, req *message.SessionEstablishmentRequest, prefix string) {
	writeLineTabbed(sb, prefix+" BAR:", 1)
	barId, err := req.CreateBAR.BARID()
	if err == nil {
		writeLineTabbed(sb, fmt.Sprintf("BAR ID: %d ", barId), 2)
	}
	downlink, err := req.CreateBAR.DownlinkDataNotificationDelay()
	if err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Downlink Data Notification Delay: %s ", downlink), 2)
	}
	suggestedBufferingPackets, err := req.CreateBAR.SuggestedBufferingPacketsCount()
	if err == nil {
		writeLineTabbed(sb, fmt.Sprintf("Suggested Buffering Packets Count: %d ", suggestedBufferingPackets), 2)
	}
	mtEdtControl, err := req.CreateBAR.MTEDTControlInformation()
	if err == nil {
		writeLineTabbed(sb, fmt.Sprintf("MT EDI: %d ", mtEdtControl), 2)
	}
}

func displayUrr(sb *strings.Builder, urr *ie.IE, prefix string) {
	writeLineTabbed(sb, prefix+" URR:", 1)
	urrId, err := urr.URRID()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("URR ID: %d ", urrId))
		writeLineTabbed(sb, fmt.Sprintf("URR ID: %d ", urrId), 2)
	}
	measurementMethod, err := urr.MeasurementMethod()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("Measurement Method: %d ", measurementMethod))
		writeLineTabbed(sb, fmt.Sprintf("Measurement Method: %d ", measurementMethod), 2)
	}
	volumeThreshold, err := urr.VolumeThreshold()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("Volume Threshold: %+v ", volumeThreshold))
		writeLineTabbed(sb, fmt.Sprintf("Volume Threshold: %+v ", volumeThreshold), 2)
	}
	timeThreshold, err := urr.TimeThreshold()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("Time Threshold: %d ", timeThreshold))
		writeLineTabbed(sb, fmt.Sprintf("Time Threshold: %d ", timeThreshold), 2)
	}
	monitoringTime, err := urr.MonitoringTime()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("Monitoring Time: %s ", monitoringTime.Format(time.RFC3339)))
		writeLineTabbed(sb, fmt.Sprintf("Monitoring Time: %s ", monitoringTime.Format(time.RFC3339)), 2)
	}
}

func displayQer(sb *strings.Builder, qer *ie.IE, prefix string) {
	writeLineTabbed(sb, prefix+" QER:", 1)

	qerId, err := qer.QERID()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("QER ID: %d ", qerId))
		writeLineTabbed(sb, fmt.Sprintf("QER ID: %d ", qerId), 2)
	}
	gateStatusDL, err := qer.GateStatusDL()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("Gate Status DL: %d ", gateStatusDL))
		writeLineTabbed(sb, fmt.Sprintf("Gate Status DL: %d ", gateStatusDL), 2)
	}
	gateStatusUL, err := qer.GateStatusUL()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("Gate Status UL: %d ", gateStatusUL))
		writeLineTabbed(sb, fmt.Sprintf("Gate Status UL: %d ", gateStatusUL), 2)
	}
	maxBitrateDL, err := qer.MBRDL()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("Max Bitrate DL: %d ", uint32(maxBitrateDL)))
		writeLineTabbed(sb, fmt.Sprintf("Max Bitrate DL: %d ", uint32(maxBitrateDL)), 2)
	}
	maxBitrateUL, err := qer.MBRUL()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("Max Bitrate UL: %d ", uint32(maxBitrateUL)))
		writeLineTabbed(sb, fmt.Sprintf("Max Bitrate UL: %d ", uint32(maxBitrateUL)), 2)
	}
	qfi, err := qer.QFI()
	if err == nil {
		//sb.WriteString(fmt.Sprintf("QFI: %d ", qfi))
		writeLineTabbed(sb, fmt.Sprintf("QFI: %d ", qfi), 2)
	}
	log.Println(sb.String())
}

func displayFar(sb *strings.Builder, far *ie.IE, prefix string) {
	writeLineTabbed(sb, prefix+" FAR:", 1)
	farId, err := far.FARID()
	if err == nil {
		writeLineTabbed(sb, (fmt.Sprintf("FAR ID: %d ", farId)), 2)
	}
	applyAction, err := far.ApplyAction()
	if err == nil {
		writeLineTabbed(sb, (fmt.Sprintf("Apply Action: %+v ", applyAction)), 2)
	}
	forwardingParameters, err := far.ForwardingParameters()
	writeLineTabbed(sb, ("Forwarding Parameters:"), 2)
	if err == nil {
		for _, forwardingParameter := range forwardingParameters {
			networkInstance, err := forwardingParameter.NetworkInstance()
			if err == nil {
				//sb.WriteString(fmt.Sprintf("Network Instance: %s ", networkInstance))
				writeLineTabbed(sb, (fmt.Sprintf("Network Instance: %s ", networkInstance)), 3)
			}
			redirectInformation, err := forwardingParameter.RedirectInformation()
			if err == nil {
				// sb.WriteString(fmt.Sprintf("Redirect Information, server address: %s ", redirectInformation.RedirectServerAddress))
				// sb.WriteString(fmt.Sprintf("Redirect Information, other server address: %s ", redirectInformation.OtherRedirectServerAddress))
				writeLineTabbed(sb, (fmt.Sprintf("Redirect Information, server address: %s ", redirectInformation.RedirectServerAddress)), 3)
				writeLineTabbed(sb, (fmt.Sprintf("Redirect Information, other server address: %s ", redirectInformation.OtherRedirectServerAddress)), 3)
			}
			headerEnrichment, err := forwardingParameter.HeaderEnrichment()
			if err == nil {
				// sb.WriteString(fmt.Sprintf("Header Enrichment: %s : %s ", headerEnrichment.HeaderFieldName, headerEnrichment.HeaderFieldValue))
				writeLineTabbed(sb, (fmt.Sprintf("Header Enrichment: %s : %s ", headerEnrichment.HeaderFieldName, headerEnrichment.HeaderFieldValue)), 3)
			}
		}
	}
	duplicatingParameters, err := far.DuplicatingParameters()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("Duplicating Parameters: %+v ", duplicatingParameters))
		writeLineTabbed(sb, (fmt.Sprintf("Duplicating Parameters: %+v ", duplicatingParameters)), 2)
	}
	barId, err := far.BARID()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("BAR ID: %d ", barId))
		writeLineTabbed(sb, (fmt.Sprintf("BAR ID: %d ", barId)), 2)
	}
	outerHeaderCreation, err := far.OuterHeaderCreation()
	if err == nil {
		// sb.WriteString(fmt.Sprintf("Outer Header Creation: %+v ", outerHeaderCreation))
		writeLineTabbed(sb, (fmt.Sprintf("Outer Header Creation: %+v ", outerHeaderCreation)), 2)
	}
}

func displayPdr(sb *strings.Builder, pdr *ie.IE, prefix string) {
	writeLineTabbed(sb, prefix+" PDR:", 1)
	pdrId, err := pdr.PDRID()
	if err == nil {
		writeLineTabbed(sb, fmt.Sprintf("PDR ID: %d ", pdrId), 2)
	}

	outerHeaderRemoval, err := pdr.OuterHeaderRemovalDescription()
	if err == nil {
		writeLineTabbed(sb, (fmt.Sprintf("Outer Header Removal: %d ", outerHeaderRemoval)), 2)
	}

	farid, err := pdr.FARID()
	if err == nil {
		writeLineTabbed(sb, (fmt.Sprintf("FAR ID: %d ", uint32(farid))), 2)
	}

	pdi, err := pdr.PDI()
	if err == nil {
		srcIfacePdiId := findIEindex(pdi, 20) // IE Type source interface
		srcInterface, _ := pdi[srcIfacePdiId].SourceInterface()
		writeLineTabbed(sb, (fmt.Sprintf("Source Interface: %d ", srcInterface)), 2)
		if srcInterface == ie.SrcInterfaceAccess {
			teidPdiId := findIEindex(pdi, 21) // IE Type F-TEID

			if teidPdiId != -1 {
				fteid, err := pdi[teidPdiId].FTEID()
				if err == nil {
					writeLineTabbed(sb, (fmt.Sprintf("TEID: %d ", fteid.TEID)), 2)
				}
			}
		} else {
			ueipPdiId := findIEindex(pdi, 93) // IE Type UE IP Address
			if ueipPdiId != -1 {
				ue_ip, _ := pdi[ueipPdiId].UEIPAddress()
				writeLineTabbed(sb, (fmt.Sprintf("UE IP Address: %s ", ue_ip.IPv4Address)), 2)
			}
		}
	}
}
