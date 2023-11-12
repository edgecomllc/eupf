package core

import (
	"fmt"
	"github.com/edgecomllc/eupf/cmd/core/service"
	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
)

func deletePDR(spdrInfo SPDRInfo, mapOperations ebpf.ForwardingPlaneController) error {
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
	return nil
}

func extractPDR(pdr *ie.IE, session *Session, spdrInfo *SPDRInfo, ipam *service.IPAM, seid uint64) error {

	if outerHeaderRemoval, err := pdr.OuterHeaderRemovalDescription(); err == nil {
		spdrInfo.PdrInfo.OuterHeaderRemoval = outerHeaderRemoval
	}
	if farid, err := pdr.FARID(); err == nil {
		spdrInfo.PdrInfo.FarId = session.GetFar(farid).GlobalId
	}
	if qerid, err := pdr.QERID(); err == nil {
		spdrInfo.PdrInfo.QerId = session.GetQer(qerid).GlobalId
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
			if ipam != nil {

				if fteid.HasCh() {
					pdrID, err := pdr.PDRID()
					if err != nil {
						log.Info().Msgf("parse PDRID err: %v", err)
					}

					if fteid.HasChID() {
						//try to find teid previously allocated
						// if err, teid := teidCache.GetTEID(fteid.ChooseID); err == nil {
						// 	spdrInfo.Teid = teid
						// 	return nil
						// }
						teid, ok := ipam.GetTEID(seid, pdrID, fteid.ChooseID)
						if !ok {
							teid, err = ipam.AllocateTEID(seid, pdrID, fteid.ChooseID)
							if err != nil {
								log.Info().Msgf("[ERROR] AllocateTEID err: %v", err)
								return fmt.Errorf("Can't allocate TEID: %s", causeToString(ie.CauseNoResourcesAvailable))
							}
						}
						spdrInfo.Teid = teid
						return nil
					}

					teid, err := ipam.AllocateTEID(seid, pdrID, fteid.ChooseID)
					if err != nil {
						log.Error().Msgf("AllocateTEID error: %v", err)
						return fmt.Errorf("Can't allocate TEID: %s", causeToString(ie.CauseNoResourcesAvailable))
					}

					spdrInfo.Teid = teid
					return nil
				} else {
					spdrInfo.Teid = fteid.TEID
					return nil
				}
			} else {
				spdrInfo.Teid = fteid.TEID
				return nil
			}
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
