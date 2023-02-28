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
	fmt.Printf("Session Establishment Request: \n")

	for _, pdr := range req.CreatePDR {
		display_create_pdr(pdr)
	}

	for _, far := range req.CreateFAR {
		display_create_far(far)
	}

	for _, qer := range req.CreateQER {
		display_create_qer(qer)
	}

	for _, urr := range req.CreateURR {
		display_create_urr(urr)
	}

	if req.CreateBAR != nil {
		display_create_bar(req)
	}
}

func display_create_bar(req *message.SessionEstablishmentRequest) {
	log.Printf("------ Create BAR: %+v", req.CreateBAR)
	var sb strings.Builder
	bar_id, err := req.CreateBAR.BARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("BAR ID: %d \t", bar_id))
	}
	downlink, err := req.CreateBAR.DownlinkDataNotificationDelay()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Downlink Data Notification Delay: %s \t", downlink))
	}
	suggested_buffering_packets, err := req.CreateBAR.SuggestedBufferingPacketsCount()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Suggested Buffering Packets Count: %d \t", suggested_buffering_packets))
	}
	mt_edt_control, err := req.CreateBAR.MTEDTControlInformation()
	if err == nil {
		sb.WriteString(fmt.Sprintf("MT EDI: %d \t", mt_edt_control))
	}
}

func display_create_urr(urr *ie.IE) {
	log.Printf("------ Create URR: %+v", urr)
	var sb strings.Builder
	urr_id, err := urr.URRID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("URR ID: %d \t", urr_id))
	}
	measurement_method, err := urr.MeasurementMethod()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Measurement Method: %d \t", measurement_method))
	}
	volume_threshold, err := urr.VolumeThreshold()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Volume Threshold: %+v \t", volume_threshold))
	}
	time_threshold, err := urr.TimeThreshold()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Time Threshold: %d \t", time_threshold))
	}
	monitoring_time, err := urr.MonitoringTime()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Monitoring Time: %s \t", monitoring_time.Format(time.RFC3339)))
	}
	log.Println(sb.String())
}

func display_create_qer(qer *ie.IE) {
	log.Printf("------ Create QER: %+v", qer)
	var sb strings.Builder
	qer_id, err := qer.QERID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("QER ID: %d \t", qer_id))
	}
	qfi, err := qer.QFI()
	if err == nil {
		sb.WriteString(fmt.Sprintf("QFI: %d \t", qfi))
	}
	gate_status, err := qer.GateStatus()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Gate Status: %d \t", gate_status))
	}
	mbr, err := qer.MBR()
	if err == nil {
		sb.WriteString(fmt.Sprintf("MBR: %+v \t", mbr))
	}
	gbr, err := qer.GBR()
	if err == nil {
		sb.WriteString(fmt.Sprintf("GBR: %+v \t", gbr))
	}
	packet_rate, err := qer.PacketRate()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Packet Rate: %+v \t", packet_rate))
	}
}

func display_create_far(far *ie.IE) {
	log.Printf("------ Create FAR: %+v", far)
	var sb strings.Builder
	far_id, err := far.FARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("FAR ID: %d \t", far_id))
	}
	apply_action, err := far.ApplyAction()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Apply Action: %+v \t", apply_action))
	}
	forwarding_parameters, err := far.ForwardingParameters()
	if err == nil {
		for _, forwarding_parameter := range forwarding_parameters {
			network_instance, err := forwarding_parameter.NetworkInstance()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Network Instance: %s \t", network_instance))
			}
			redirect_information, err := forwarding_parameter.RedirectInformation()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Redirect Information, server address: %s \t", redirect_information.RedirectServerAddress))
				sb.WriteString(fmt.Sprintf("Redirect Information, other server address: %s \t", redirect_information.OtherRedirectServerAddress))
			}
			header_enrichment, err := forwarding_parameter.HeaderEnrichment()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Header Enrichment: %s : %s \t", header_enrichment.HeaderFieldName, header_enrichment.HeaderFieldValue))
			}
		}
	}
	duplicating_parameters, err := far.DuplicatingParameters()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Duplicating Parameters: %+v \t", duplicating_parameters))
	}
	bar_id, err := far.BARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("BAR ID: %d \t", bar_id))
	}
	outer_header_creation, err := far.OuterHeaderCreation()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Outer Header Creation: %+v \t", outer_header_creation))
	}
	log.Println(sb.String())
}

func display_create_pdr(pdr *ie.IE) {
	log.Printf("------ Create PDR: %+v", pdr)
	var sb strings.Builder
	pdr_id, err := pdr.PDRID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("PDR ID: %d \t", pdr_id))
	}
	precedence, err := pdr.Precedence()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Precedence: %d", precedence))
	}
	pdi_ies, err := pdr.PDI()
	if err == nil {
		for _, pdi := range pdi_ies {
			source_interface, err := pdi.SourceInterface()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Source Interface: %d \t", source_interface))
			}
			f_teid, err := pdi.FTEID()
			if err == nil {
				sb.WriteString(fmt.Sprintf("F-TEID: %+v \t", f_teid))
			}
			network_instance, err := pdi.NetworkInstance()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Network Instance: %s \t", network_instance))
			}
			redurant_transmission_parameters, err := pdi.RedundantTransmissionParameters()
			if err == nil {
				for _, rtp := range redurant_transmission_parameters {
					local_f_teid, err := rtp.FTEID()
					if err == nil {
						sb.WriteString(fmt.Sprintf("Local F-TEID: %+v \t", local_f_teid))
					}
					network_instance, err := rtp.NetworkInstance()
					if err == nil {
						sb.WriteString(fmt.Sprintf("Network Instance: %s \t", network_instance))
					}
				}
			}
		}
	}
	outer_header_removal, err := pdr.OuterHeaderRemoval()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Outer Header Removal: %+v \t", outer_header_removal))
	}
	far_id, err := pdr.FARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("FAR ID: %d \t", far_id))
	}
	urr_id, err := pdr.URRID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("URR ID: %d \t", urr_id))
	}
	qer_id, err := pdr.QERID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("QER ID: %d \t", qer_id))
	}
	activate_predefined_rules, err := pdr.ActivatePredefinedRules()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Activate Predefined Rules: %s \t", activate_predefined_rules))
	}
	activation_time, err := pdr.ActivationTime()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Activation Time: %s \t", activation_time.Format(time.RFC3339)))
	}
	deactivation_time, err := pdr.DeactivationTime()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Deactivation Time: %s \t", deactivation_time.Format(time.RFC3339)))
	}
	mar_id, err := pdr.MARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("MAR ID: %d \t", mar_id))
	}
	packet_replication_and_detection_carry_on_information, err := pdr.PacketReplicationAndDetectionCarryOnInformation()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Packet Replication and Detection Carry On Information: %+v \t", packet_replication_and_detection_carry_on_information))
	}
	ip_multicast_addressing_info, err := pdr.IPMulticastAddressingInfo()
	if err == nil {
		for _, ipma := range ip_multicast_addressing_info {
			ip_multicast_address, err := ipma.IPMulticastAddress()
			if err == nil {
				sb.WriteString(fmt.Sprintf("IP Multicast Address: %+v \t", ip_multicast_address))
			}
			source_ip_address, err := ipma.SourceIPAddress()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Source IP Address: %+v \t", source_ip_address))
			}
		}
	}
	ue_ip_address_pool_identity, err := pdr.UEIPAddressPoolIdentity()
	if err == nil {
		sb.WriteString(fmt.Sprintf("UE IP Address Pool Identity: %+vd \t", ue_ip_address_pool_identity))
	}
	mptcp_applicable_indication, err := pdr.MPTCPApplicableIndication()
	if err == nil {
		sb.WriteString(fmt.Sprintf("MPTCP Applicable Indication: %d \t", mptcp_applicable_indication))
	}

	log.Println(sb.String())
}
