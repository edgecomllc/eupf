package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/wmnsk/go-pfcp/message"
)

// Это можно будет разобрать на отдельные функции, когда будем делать вывод для остальных сообщений.
func printSessionEstablishmentRequest(req *message.SessionEstablishmentRequest) {
	fmt.Printf("Session Establishment Request: \n")

	for _, pdr := range req.CreatePDR {
		log.Printf("------ Create PDR: %+v", pdr)
		var sb strings.Builder
		pdr_id, err := pdr.PDRID()
		if err == nil {
			sb.WriteString(fmt.Sprintf("PDR ID: %d \t", pdr_id))
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
		precedence, err := pdr.Precedence()
		if err == nil {
			sb.WriteString(fmt.Sprintf("Precedence: %d", precedence))
		}
		log.Println(sb.String())
	}

	for _, far := range req.CreateFAR {
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
				destination_interface, err := forwarding_parameter.DestinationInterface()
				if err == nil {
					sb.WriteString(fmt.Sprintf("Destination Interface: %d \t", destination_interface))
				}
				network_instance, err := forwarding_parameter.NetworkInstance()
				if err == nil {
					sb.WriteString(fmt.Sprintf("Network Instance: %s \t", network_instance))
				}
				outer_header_creation, err := forwarding_parameter.OuterHeaderCreation()
				if err == nil {
					sb.WriteString(fmt.Sprintf("Outer Header Creation: %+v \t", outer_header_creation))
				}
				transport_level_marking, err := forwarding_parameter.TransportLevelMarking()
				if err == nil {
					sb.WriteString(fmt.Sprintf("Transport Level Marking: %d \t", transport_level_marking))
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

	for _, qer := range req.CreateQER {
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

	for _, urr := range req.CreateURR {
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
			sb.WriteString(fmt.Sprintf("Monitoring Time: %s \t", monitoring_time))
		}
		log.Println(sb.String())
	}
}
