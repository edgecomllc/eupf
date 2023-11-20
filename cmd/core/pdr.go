package core

import (
	"fmt"
	"net"

	"github.com/edgecomllc/eupf/cmd/core/service"
	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
)

func deletePDR(spdrInfo SPDRInfo, mapOperations ebpf.ForwardingPlaneController, resourceManager *service.ResourceManager, seID uint64) error {
	if spdrInfo.Ipv4 != nil {
		if err := mapOperations.DeletePdrDownLink(spdrInfo.Ipv4); err != nil {
			return fmt.Errorf("Can't delete IPv4 PDR: %s", err.Error())
		}
	} else if spdrInfo.Ipv6 != nil {
		if err := mapOperations.DeleteDownlinkPdrIp6(spdrInfo.Ipv6); err != nil {
			return fmt.Errorf("Can't delete IPv6 PDR: %s", err.Error())
		}
	} else {
		if err := mapOperations.DeletePdrUpLink(spdrInfo.Teid); err != nil {
			return fmt.Errorf("Can't delete GTP PDR: %s", err.Error())
		}
	}
	if spdrInfo.Teid != 0 {
		if resourceManager.FTEIDM != nil {
			resourceManager.FTEIDM.ReleaseTEID(seID)
		}
	}
	return nil
}

// func extractPDR(pdr *ie.IE, session *Session, spdrInfo *SPDRInfo, resourceManager *service.ResourceManager, teidCache map[uint8]uint32) error {
func extractPDR(pdr *ie.IE, spdrInfo *SPDRInfo, pdrContext *PDRCreationContext) error {
	if outerHeaderRemoval, err := pdr.OuterHeaderRemovalDescription(); err == nil {
		spdrInfo.PdrInfo.OuterHeaderRemoval = outerHeaderRemoval
	}
	if farid, err := pdr.FARID(); err == nil {
		spdrInfo.PdrInfo.FarId = pdrContext.getFARID(farid)
	}
	if qerid, err := pdr.QERID(); err == nil {
		spdrInfo.PdrInfo.QerId = pdrContext.getQERID(qerid)
	}

	pdi, err := pdr.PDI()
	if err != nil {
		return fmt.Errorf("PDI IE is missing")
	}

	if sdfFilter, err := pdr.SDFFilter(); err == nil {
		if sdfFilterParsed, err := ParseSdfFilter(sdfFilter.FlowDescription); err == nil {
			spdrInfo.PdrInfo.SdfFilter = &sdfFilterParsed
			// log.Printf("Sdf Filter Parsed: %+v", sdfFilterParsed)
		} else {
			return err
		}
	}

	//Bug in go-pfcp:
	//if fteid, err := pdr.FTEID(); err == nil {
	if teidPdiId := findIEindex(pdi, 21); teidPdiId != -1 { // IE Type F-TEID
		if fteid, err := pdi[teidPdiId].FTEID(); err == nil {
			var teid = fteid.TEID
			if pdrContext.ResourceManager.FTEIDM != nil {
				if fteid.HasCh() {
					var allocate = true
					if fteid.HasChID() {
						if teidFromCache, ok := pdrContext.hasTEIDCache(fteid.ChooseID); ok {
							allocate = false
							teid = teidFromCache
							spdrInfo.Allocated = true
						}
					}
					if allocate {
						allocatedTeid, err := pdrContext.getFTEID(pdrContext.Session.RemoteSEID, spdrInfo.PdrID)
						if err != nil {
							log.Info().Msgf("[ERROR] AllocateTEID err: %v", err)
							return fmt.Errorf("Can't allocate TEID: %s", causeToString(ie.CauseNoResourcesAvailable))
						}
						teid = allocatedTeid
						spdrInfo.Allocated = true
						if fteid.HasChID() {
							pdrContext.setTEIDCache(fteid.ChooseID, teid)
						}
					}
				}
			}
			spdrInfo.Teid = teid
			return nil
		}
		return fmt.Errorf("F-TEID IE is missing")
	} else if ueIP, err := pdr.UEIPAddress(); err == nil {
		if ueIP.IPv4Address != nil {
			spdrInfo.Ipv4 = cloneIP(ueIP.IPv4Address)
		} else if ueIP.IPv6Address != nil {
			spdrInfo.Ipv6 = cloneIP(ueIP.IPv6Address)
		} else {
			return fmt.Errorf("UE IP Address IE is missing")
		}

		return nil
	} else {
		log.Info().Msg("Both F-TEID IE and UE IP Address IE are missing")
		return err
	}
}

func applyPDR(spdrInfo SPDRInfo, mapOperations ebpf.ForwardingPlaneController) {
	if spdrInfo.Ipv4 != nil {
		if err := mapOperations.PutPdrDownLink(spdrInfo.Ipv4, spdrInfo.PdrInfo); err != nil {
			log.Info().Msgf("Can't apply IPv4 PDR: %s", err.Error())
		}
	} else if spdrInfo.Ipv6 != nil {
		if err := mapOperations.PutDownlinkPdrIp6(spdrInfo.Ipv6, spdrInfo.PdrInfo); err != nil {
			log.Info().Msgf("Can't apply IPv6 PDR: %s", err.Error())
		}
	} else {
		if err := mapOperations.PutPdrUpLink(spdrInfo.Teid, spdrInfo.PdrInfo); err != nil {
			log.Info().Msgf("Can't apply GTP PDR: %s", err.Error())
		}
	}
}

func processCreatedPDRs(createdPDRs []SPDRInfo, n3Address net.IP) []*ie.IE {

	var additionalIEs []*ie.IE

	allocatedPDRs := []uint32{}
	for _, pdr := range createdPDRs {
		if pdr.Allocated {
			allocatedPDRs = append(allocatedPDRs, pdr.PdrID)
			if pdr.Ipv4 != nil {
				additionalIEs = append(additionalIEs, ie.NewCreatedPDR(ie.NewPDRID(uint16(pdr.PdrID)), ie.NewUEIPAddress(0, pdr.Ipv4.String(), "", 0, 0)))
			} else if pdr.Ipv6 != nil {

			} else {
				additionalIEs = append(additionalIEs, ie.NewCreatedPDR(ie.NewPDRID(uint16(pdr.PdrID)), ie.NewFTEID(1, pdr.Teid, n3Address, nil, 0)))
			}
		}
	}
	log.Info().Msgf("*********************ALLOCATED PDRS COUNT: %d, PDRIDS: %v", len(allocatedPDRs), allocatedPDRs)
	return additionalIEs
}
