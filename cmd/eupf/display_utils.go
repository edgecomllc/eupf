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

	for _, pdr := range req.CreatePDR {
		displayCreatePdr(pdr)
	}

	for _, far := range req.CreateFAR {
		displayCreateFar(far)
	}

	for _, qer := range req.CreateQER {
		displayCreateQer(qer)
	}

	for _, urr := range req.CreateURR {
		displayCreateUrr(urr)
	}

	if req.CreateBAR != nil {
		displayCreateBar(req)
	}
}

func displayCreateBar(req *message.SessionEstablishmentRequest) {
	log.Printf("------ Create BAR: %+v", req.CreateBAR)
	var sb strings.Builder
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

func displayCreateUrr(urr *ie.IE) {
	log.Printf("------ Create URR: %+v", urr)
	var sb strings.Builder
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

func displayCreateQer(qer *ie.IE) {
	log.Printf("------ Create QER: %+v", qer)
	var sb strings.Builder
	qerId, err := qer.QERID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("QER ID: %d \n", qerId))
	}
	qfi, err := qer.QFI()
	if err == nil {
		sb.WriteString(fmt.Sprintf("QFI: %d \n", qfi))
	}
	gateStatus, err := qer.GateStatus()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Gate Status: %d \n", gateStatus))
	}
	mbr, err := qer.MBR()
	if err == nil {
		sb.WriteString(fmt.Sprintf("MBR: %+v \n", mbr))
	}
	gbr, err := qer.GBR()
	if err == nil {
		sb.WriteString(fmt.Sprintf("GBR: %+v \n", gbr))
	}
	packetRate, err := qer.PacketRate()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Packet Rate: %+v \n", packetRate))
	}
}

func displayCreateFar(far *ie.IE) {
	log.Printf("------ Create FAR: %+v", far)
	var sb strings.Builder
	farId, err := far.FARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("FAR ID: %d \n", farId))
	}
	applyAction, err := far.ApplyAction()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Apply Action: %+v \n", applyAction))
	}
	forwardingParameters, err := far.ForwardingParameters()
	if err == nil {
		for _, forwardingParameter := range forwardingParameters {
			networkInstance, err := forwardingParameter.NetworkInstance()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Network Instance: %s \n", networkInstance))
			}
			redirectInformation, err := forwardingParameter.RedirectInformation()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Redirect Information, server address: %s \n", redirectInformation.RedirectServerAddress))
				sb.WriteString(fmt.Sprintf("Redirect Information, other server address: %s \n", redirectInformation.OtherRedirectServerAddress))
			}
			headerEnrichment, err := forwardingParameter.HeaderEnrichment()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Header Enrichment: %s : %s \n", headerEnrichment.HeaderFieldName, headerEnrichment.HeaderFieldValue))
			}
		}
	}
	duplicatingParameters, err := far.DuplicatingParameters()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Duplicating Parameters: %+v \n", duplicatingParameters))
	}
	barId, err := far.BARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("BAR ID: %d \n", barId))
	}
	outerHeaderCreation, err := far.OuterHeaderCreation()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Outer Header Creation: %+v \n", outerHeaderCreation))
	}
	log.Println(sb.String())
}

func displayCreatePdr(pdr *ie.IE) {
	log.Printf("------ Create PDR: %+v", pdr)
	var sb strings.Builder
	pdrId, err := pdr.PDRID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("PDR ID: %d \n", pdrId))
	}
	precedence, err := pdr.Precedence()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Precedence: %d \n", precedence))
	}
	pdiIes, err := pdr.PDI()
	if err == nil {
		for _, pdi := range pdiIes {
			sourceInterface, err := pdi.SourceInterface()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Source Interface: %d \n", sourceInterface))
			}
			fTeid, err := pdi.FTEID()
			if err == nil {
				sb.WriteString(fmt.Sprintf("F-TEID: %+v \n", fTeid))
			}
			networkInstance, err := pdi.NetworkInstance()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Network Instance: %s \n", networkInstance))
			}
			redurantTransmissionParameters, err := pdi.RedundantTransmissionParameters()
			if err == nil {
				for _, rtp := range redurantTransmissionParameters {
					localFTeid, err := rtp.FTEID()
					if err == nil {
						sb.WriteString(fmt.Sprintf("Local F-TEID: %+v \n", localFTeid))
					}
					networkInstance, err := rtp.NetworkInstance()
					if err == nil {
						sb.WriteString(fmt.Sprintf("Network Instance: %s \n", networkInstance))
					}
				}
			}
		}
	}
	outerHeaderRemoval, err := pdr.OuterHeaderRemoval()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Outer Header Removal: %+v \n", outerHeaderRemoval))
	}
	farId, err := pdr.FARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("FAR ID: %d \n", farId))
	}
	urrId, err := pdr.URRID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("URR ID: %d \n", urrId))
	}
	qerId, err := pdr.QERID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("QER ID: %d \n", qerId))
	}
	activatePredefinedRules, err := pdr.ActivatePredefinedRules()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Activate Predefined Rules: %s \n", activatePredefinedRules))
	}
	activationTime, err := pdr.ActivationTime()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Activation Time: %s \n", activationTime.Format(time.RFC3339)))
	}
	deactivationTime, err := pdr.DeactivationTime()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Deactivation Time: %s \n", deactivationTime.Format(time.RFC3339)))
	}
	marId, err := pdr.MARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("MAR ID: %d \n", marId))
	}
	packetReplicationAndDetectionCarryOnInformation, err := pdr.PacketReplicationAndDetectionCarryOnInformation()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Packet Replication and Detection Carry On Information: %+v \n", packetReplicationAndDetectionCarryOnInformation))
	}
	ipMulticastAddressingInfo, err := pdr.IPMulticastAddressingInfo()
	if err == nil {
		for _, ipma := range ipMulticastAddressingInfo {
			ipMulticastAddress, err := ipma.IPMulticastAddress()
			if err == nil {
				sb.WriteString(fmt.Sprintf("IP Multicast Address: %+v \n", ipMulticastAddress))
			}
			sourceIpAddress, err := ipma.SourceIPAddress()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Source IP Address: %+v \n", sourceIpAddress))
			}
		}
	}
	ueIpAddressPoolIdentity, err := pdr.UEIPAddressPoolIdentity()
	if err == nil {
		sb.WriteString(fmt.Sprintf("UE IP Address Pool Identity: %+vd \n", ueIpAddressPoolIdentity))
	}
	mptcpApplicableIndication, err := pdr.MPTCPApplicableIndication()
	if err == nil {
		sb.WriteString(fmt.Sprintf("MPTCP Applicable Indication: %d \n", mptcpApplicableIndication))
	}

	log.Println(sb.String())
}
