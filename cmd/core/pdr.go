package core

import (
	"net"

	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
)

const flagPresentIPv4 = 2

func applyPDR(spdrInfo SPDRInfo, mapOperations ebpf.ForwardingPlaneController) {
	if spdrInfo.Ipv4 != nil {
		if err := mapOperations.PutPdrDownlink(spdrInfo.Ipv4, spdrInfo.PdrInfo); err != nil {
			log.Info().Msgf("Can't apply IPv4 PDR: %s", err.Error())
		}
	} else if spdrInfo.Ipv6 != nil {
		if err := mapOperations.PutDownlinkPdrIp6(spdrInfo.Ipv6, spdrInfo.PdrInfo); err != nil {
			log.Info().Msgf("Can't apply IPv6 PDR: %s", err.Error())
		}
	} else {
		if err := mapOperations.PutPdrUplink(spdrInfo.Teid, spdrInfo.PdrInfo); err != nil {
			log.Info().Msgf("Can't apply GTP PDR: %s", err.Error())
		}
	}
}

func processCreatedPDRs(createdPDRs []SPDRInfo, n3Address net.IP) []*ie.IE {
	var additionalIEs []*ie.IE
	for _, pdr := range createdPDRs {
		if pdr.Allocated {
			if pdr.Ipv4 != nil {
				additionalIEs = append(additionalIEs, ie.NewCreatedPDR(ie.NewPDRID(uint16(pdr.PdrID)), ie.NewUEIPAddress(flagPresentIPv4, pdr.Ipv4.String(), "", 0, 0)))
			} else if pdr.Ipv6 != nil {

			} else {
				additionalIEs = append(additionalIEs, ie.NewCreatedPDR(ie.NewPDRID(uint16(pdr.PdrID)), ie.NewFTEID(0x01, pdr.Teid, cloneIP(n3Address), nil, 0)))
			}
		}
	}
	return additionalIEs
}
