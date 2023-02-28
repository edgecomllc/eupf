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
		sb.WriteString(fmt.Sprintf("BAR ID: %d \n", bar_id))
	}
	downlink, err := req.CreateBAR.DownlinkDataNotificationDelay()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Downlink Data Notification Delay: %s \n", downlink))
	}
	suggested_buffering_packets, err := req.CreateBAR.SuggestedBufferingPacketsCount()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Suggested Buffering Packets Count: %d \n", suggested_buffering_packets))
	}
	mt_edt_control, err := req.CreateBAR.MTEDTControlInformation()
	if err == nil {
		sb.WriteString(fmt.Sprintf("MT EDI: %d \n", mt_edt_control))
	}
}

func display_create_urr(urr *ie.IE) {
	log.Printf("------ Create URR: %+v", urr)
	var sb strings.Builder
	urr_id, err := urr.URRID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("URR ID: %d \n", urr_id))
	}
	measurement_method, err := urr.MeasurementMethod()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Measurement Method: %d \n", measurement_method))
	}
	volume_threshold, err := urr.VolumeThreshold()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Volume Threshold: %+v \n", volume_threshold))
	}
	time_threshold, err := urr.TimeThreshold()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Time Threshold: %d \n", time_threshold))
	}
	monitoring_time, err := urr.MonitoringTime()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Monitoring Time: %s \n", monitoring_time.Format(time.RFC3339)))
	}
	log.Println(sb.String())
}

func display_create_qer(qer *ie.IE) {
	log.Printf("------ Create QER: %+v", qer)
	var sb strings.Builder
	qer_id, err := qer.QERID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("QER ID: %d \n", qer_id))
	}
	qfi, err := qer.QFI()
	if err == nil {
		sb.WriteString(fmt.Sprintf("QFI: %d \n", qfi))
	}
	gate_status, err := qer.GateStatus()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Gate Status: %d \n", gate_status))
	}
	mbr, err := qer.MBR()
	if err == nil {
		sb.WriteString(fmt.Sprintf("MBR: %+v \n", mbr))
	}
	gbr, err := qer.GBR()
	if err == nil {
		sb.WriteString(fmt.Sprintf("GBR: %+v \n", gbr))
	}
	packet_rate, err := qer.PacketRate()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Packet Rate: %+v \n", packet_rate))
	}
}

func display_create_far(far *ie.IE) {
	log.Printf("------ Create FAR: %+v", far)
	var sb strings.Builder
	far_id, err := far.FARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("FAR ID: %d \n", far_id))
	}
	apply_action, err := far.ApplyAction()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Apply Action: %+v \n", apply_action))
	}
	forwarding_parameters, err := far.ForwardingParameters()
	if err == nil {
		for _, forwarding_parameter := range forwarding_parameters {
			network_instance, err := forwarding_parameter.NetworkInstance()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Network Instance: %s \n", network_instance))
			}
			redirect_information, err := forwarding_parameter.RedirectInformation()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Redirect Information, server address: %s \n", redirect_information.RedirectServerAddress))
				sb.WriteString(fmt.Sprintf("Redirect Information, other server address: %s \n", redirect_information.OtherRedirectServerAddress))
			}
			header_enrichment, err := forwarding_parameter.HeaderEnrichment()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Header Enrichment: %s : %s \n", header_enrichment.HeaderFieldName, header_enrichment.HeaderFieldValue))
			}
		}
	}
	duplicating_parameters, err := far.DuplicatingParameters()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Duplicating Parameters: %+v \n", duplicating_parameters))
	}
	bar_id, err := far.BARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("BAR ID: %d \n", bar_id))
	}
	outer_header_creation, err := far.OuterHeaderCreation()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Outer Header Creation: %+v \n", outer_header_creation))
	}
	log.Println(sb.String())
}

func display_create_pdr(pdr *ie.IE) {
	log.Printf("------ Create PDR: %+v", pdr)
	var sb strings.Builder
	pdr_id, err := pdr.PDRID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("PDR ID: %d \n", pdr_id))
	}
	precedence, err := pdr.Precedence()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Precedence: %d \n", precedence))
	}
	pdi_ies, err := pdr.PDI()
	if err == nil {
		for _, pdi := range pdi_ies {
			source_interface, err := pdi.SourceInterface()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Source Interface: %d \n", source_interface))
			}
			f_teid, err := pdi.FTEID()
			if err == nil {
				sb.WriteString(fmt.Sprintf("F-TEID: %+v \n", f_teid))
			}
			network_instance, err := pdi.NetworkInstance()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Network Instance: %s \n", network_instance))
			}
			redurant_transmission_parameters, err := pdi.RedundantTransmissionParameters()
			if err == nil {
				for _, rtp := range redurant_transmission_parameters {
					local_f_teid, err := rtp.FTEID()
					if err == nil {
						sb.WriteString(fmt.Sprintf("Local F-TEID: %+v \n", local_f_teid))
					}
					network_instance, err := rtp.NetworkInstance()
					if err == nil {
						sb.WriteString(fmt.Sprintf("Network Instance: %s \n", network_instance))
					}
				}
			}
		}
	}
	outer_header_removal, err := pdr.OuterHeaderRemoval()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Outer Header Removal: %+v \n", outer_header_removal))
	}
	far_id, err := pdr.FARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("FAR ID: %d \n", far_id))
	}
	urr_id, err := pdr.URRID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("URR ID: %d \n", urr_id))
	}
	qer_id, err := pdr.QERID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("QER ID: %d \n", qer_id))
	}
	activate_predefined_rules, err := pdr.ActivatePredefinedRules()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Activate Predefined Rules: %s \n", activate_predefined_rules))
	}
	activation_time, err := pdr.ActivationTime()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Activation Time: %s \n", activation_time.Format(time.RFC3339)))
	}
	deactivation_time, err := pdr.DeactivationTime()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Deactivation Time: %s \n", deactivation_time.Format(time.RFC3339)))
	}
	mar_id, err := pdr.MARID()
	if err == nil {
		sb.WriteString(fmt.Sprintf("MAR ID: %d \n", mar_id))
	}
	packet_replication_and_detection_carry_on_information, err := pdr.PacketReplicationAndDetectionCarryOnInformation()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Packet Replication and Detection Carry On Information: %+v \n", packet_replication_and_detection_carry_on_information))
	}
	ip_multicast_addressing_info, err := pdr.IPMulticastAddressingInfo()
	if err == nil {
		for _, ipma := range ip_multicast_addressing_info {
			ip_multicast_address, err := ipma.IPMulticastAddress()
			if err == nil {
				sb.WriteString(fmt.Sprintf("IP Multicast Address: %+v \n", ip_multicast_address))
			}
			source_ip_address, err := ipma.SourceIPAddress()
			if err == nil {
				sb.WriteString(fmt.Sprintf("Source IP Address: %+v \n", source_ip_address))
			}
		}
	}
	ue_ip_address_pool_identity, err := pdr.UEIPAddressPoolIdentity()
	if err == nil {
		sb.WriteString(fmt.Sprintf("UE IP Address Pool Identity: %+vd \n", ue_ip_address_pool_identity))
	}
	mptcp_applicable_indication, err := pdr.MPTCPApplicableIndication()
	if err == nil {
		sb.WriteString(fmt.Sprintf("MPTCP Applicable Indication: %d \n", mptcp_applicable_indication))
	}

	log.Println(sb.String())
}
