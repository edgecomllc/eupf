package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func printSessionEstablishmentRequest(req *message.SessionEstablishmentRequest) {
	log.Printf("Session Establishment Request: \n")

	log.Println("------ Create:")
	for _, pdr := range req.CreatePDR {
		displayPdr(pdr)
	}

	for _, far := range req.CreateFAR {
		displayFar(far)
	}

	for _, qer := range req.CreateQER {
		displayQer(qer)
	}

	for _, urr := range req.CreateURR {
		displayUrr(urr)
	}

	if req.CreateBAR != nil {
		displayBar(req)
	}
}

// IE Contents of Create/Update/Remove are mostly the same
func printSessionModificationRequest(req *message.SessionModificationRequest) {
	log.Printf("Session Modification Request:")
	log.Println("------ Update:")
	for _, pdr := range req.UpdatePDR {
		displayPdr(pdr)
	}

	for _, far := range req.UpdateFAR {
		displayFar(far)
	}

	for _, qer := range req.UpdateQER {
		displayQer(qer)
	}

	for _, urr := range req.UpdateURR {
		displayUrr(urr)
	}

	if req.UpdateBAR != nil {
		var sb strings.Builder
		sb.WriteString("------ BAR")
		barId, err := req.UpdateBAR.BARID()
		if err == nil {
			sb.WriteString(fmt.Sprintf("BAR ID: %d \n", barId))
		}
		downlink, err := req.UpdateBAR.DownlinkDataNotificationDelay()
		if err == nil {
			sb.WriteString(fmt.Sprintf("Downlink Data Notification Delay: %s \n", downlink))
		}
		suggestedBufferingPackets, err := req.UpdateBAR.SuggestedBufferingPacketsCount()
		if err == nil {
			sb.WriteString(fmt.Sprintf("Suggested Buffering Packets Count: %d \n", suggestedBufferingPackets))
		}
		mtEdtControl, err := req.UpdateBAR.MTEDTControlInformation()
		if err == nil {
			sb.WriteString(fmt.Sprintf("MT EDI: %d \n", mtEdtControl))
		}
	}

	log.Println("------ Remove:")
	for _, pdr := range req.RemovePDR {
		displayPdr(pdr)
	}

	for _, far := range req.RemoveFAR {
		displayFar(far)
	}

	for _, qer := range req.RemoveQER {
		displayQer(qer)
	}

	for _, urr := range req.RemoveURR {
		displayUrr(urr)
	}

	if req.RemoveBAR != nil {
		log.Print("------ BAR:")
		var sb strings.Builder
		barId, err := req.RemoveBAR.BARID()
		if err == nil {
			sb.WriteString(fmt.Sprintf("BAR ID: %d \n", barId))
		}
	}
}

func displayBar(req *message.SessionEstablishmentRequest) {
	var sb strings.Builder
	sb.WriteString("------ BAR:")
	barId, err := req.CreateBAR.BARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("BAR ID: %d \n", barId))
	}
	downlink, err := req.CreateBAR.DownlinkDataNotificationDelay()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Downlink Data Notification Delay: %s \n", downlink))
	}
	suggestedBufferingPackets, err := req.CreateBAR.SuggestedBufferingPacketsCount()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Suggested Buffering Packets Count: %d \n", suggestedBufferingPackets))
	}
	mtEdtControl, err := req.CreateBAR.MTEDTControlInformation()
	if err == nil {
		sb.WriteString(fmt.Sprintf("MT EDI: %d \n", mtEdtControl))
	}
}

func displayUrr(urr *ie.IE) {
	var sb strings.Builder
	sb.WriteString("------ URR:")
	urrId, err := urr.URRID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("URR ID: %d \n", urrId))
	}
	measurementMethod, err := urr.MeasurementMethod()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Measurement Method: %d \n", measurementMethod))
	}
	volumeThreshold, err := urr.VolumeThreshold()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Volume Threshold: %+v \n", volumeThreshold))
	}
	timeThreshold, err := urr.TimeThreshold()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Time Threshold: %d \n", timeThreshold))
	}
	monitoringTime, err := urr.MonitoringTime()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Monitoring Time: %s \n", monitoringTime.Format(time.RFC3339)))
	}
	log.Println(sb.String())
}

func displayQer(qer *ie.IE) {
	var sb strings.Builder
	sb.WriteString("------ QER:")

	qerId, err := qer.QERID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tQER ID: %d \n", qerId))
	}
	gateStatusDL, err := qer.GateStatusDL()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tGate Status DL: %d \n", gateStatusDL))
	}
	gateStatusUL, err := qer.GateStatusUL()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tGate Status UL: %d \n", gateStatusUL))
	}
	maxBitrateDL, err := qer.MBRDL()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tMax Bitrate DL: %d \n", uint32(maxBitrateDL)))
	}
	maxBitrateUL, err := qer.MBRUL()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tMax Bitrate UL: %d \n", uint32(maxBitrateUL)))
	}
	qfi, err := qer.QFI()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tQFI: %d \n", qfi))
	}
	log.Println(sb.String())
}

func displayFar(far *ie.IE) {
	var sb strings.Builder
	sb.WriteString("------ FAR:")
	farId, err := far.FARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tFAR ID: %d \n", farId))
	}
	applyAction, err := far.ApplyAction()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tApply Action: %+v \n", applyAction))
	}
	forwardingParameters, err := far.ForwardingParameters()
	if err == nil {
		for _, forwardingParameter := range forwardingParameters {
			networkInstance, err := forwardingParameter.NetworkInstance()
			if err == nil {
				sb.WriteString(fmt.Sprintf("\tNetwork Instance: %s \n", networkInstance))
			}
			redirectInformation, err := forwardingParameter.RedirectInformation()
			if err == nil {
				sb.WriteString(fmt.Sprintf("\tRedirect Information, server address: %s \n", redirectInformation.RedirectServerAddress))
				sb.WriteString(fmt.Sprintf("\tRedirect Information, other server address: %s \n", redirectInformation.OtherRedirectServerAddress))
			}
			headerEnrichment, err := forwardingParameter.HeaderEnrichment()
			if err == nil {
				sb.WriteString(fmt.Sprintf("\tHeader Enrichment: %s : %s \n", headerEnrichment.HeaderFieldName, headerEnrichment.HeaderFieldValue))
			}
		}
	}
	duplicatingParameters, err := far.DuplicatingParameters()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tDuplicating Parameters: %+v \n", duplicatingParameters))
	}
	barId, err := far.BARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tBAR ID: %d \n", barId))
	}
	outerHeaderCreation, err := far.OuterHeaderCreation()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tOuter Header Creation: %+v \n", outerHeaderCreation))
	}
	log.Println(sb.String())
}

func displayPdr(pdr *ie.IE) {
	var sb strings.Builder
	sb.WriteString("------ PDR:")
	pdrId, err := pdr.PDRID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tPDR ID: %d \n", pdrId))
	}

	outerHeaderRemoval, err := pdr.OuterHeaderRemovalDescription()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tOuter Header Removal: %d \n", outerHeaderRemoval))
	}

	farid, err := pdr.FARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("\tFAR ID: %d \n", uint32(farid)))
	}

	pdi, err := pdr.PDI()
	if err == nil {
		srcIfacePdiId := findIEindex(pdi, 20) // IE Type source interface
		srcInterface, _ := pdi[srcIfacePdiId].SourceInterface()

		if srcInterface == ie.SrcInterfaceAccess {
			teidPdiId := findIEindex(pdi, 21) // IE Type F-TEID

			if teidPdiId != -1 {
				fteid, err := pdi[teidPdiId].FTEID()
				if err == nil {
					sb.WriteString(fmt.Sprintf("\tTEID: %d \n", fteid.TEID))
				}
			}
		} else {
			ueipPdiId := findIEindex(pdi, 93) // IE Type UE IP Address
			if ueipPdiId != -1 {
				ue_ip, _ := pdi[ueipPdiId].UEIPAddress()
				sb.WriteString(fmt.Sprintf("\tUE IP Address: %s \n", ue_ip.IPv4Address))
			}
		}
	}

	log.Println(sb.String())
}
