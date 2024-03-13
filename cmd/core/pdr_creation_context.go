package core

import (
	"errors"
	"fmt"
	"net"

	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/edgecomllc/eupf/cmd/core/service"
	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
)

type PDRCreationContext struct {
	Session         *Session
	ResourceManager *service.ResourceManager
	TEIDCache       map[uint8]uint32
}

func NewPDRCreationContext(session *Session, resourceManager *service.ResourceManager) *PDRCreationContext {
	return &PDRCreationContext{
		Session:         session,
		ResourceManager: resourceManager,
		TEIDCache:       make(map[uint8]uint32),
	}
}

func (pdrContext *PDRCreationContext) extractPDR(pdr *ie.IE, spdrInfo *SPDRInfo) error {
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
		if sdfFilter.FlowDescription == "" {
			log.Warn().Msgf("SDFFilter is empty")
		} else if sdfFilterParsed, err := ParseSdfFilter(sdfFilter.FlowDescription); err == nil {
			spdrInfo.PdrInfo.SdfFilter = &sdfFilterParsed
		} else {
			log.Error().Msgf("SDFFilter err: %v", err)
			return err
		}
	}

	if teidPdiId := findIEindex(pdi, 21); teidPdiId != -1 { // IE Type F-TEID
		if fteid, err := pdi[teidPdiId].FTEID(); err == nil {
			var teid = fteid.TEID
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
						log.Error().Msgf("AllocateTEID err: %v", err)
						return fmt.Errorf("can't allocate TEID: %s", causeToString(ie.CauseNoResourcesAvailable))
					}
					teid = allocatedTeid
					spdrInfo.Allocated = true
					if fteid.HasChID() {
						pdrContext.setTEIDCache(fteid.ChooseID, teid)
					}
				}
			}
			spdrInfo.Teid = teid
			return nil
		}
		return fmt.Errorf("F-TEID IE is missing")
	} else if ueIP, err := pdr.UEIPAddress(); err == nil {
		if config.Conf.FeatureUEIP && hasCHV4(ueIP.Flags) {
			if ip, err := pdrContext.getIP(); err == nil {
				ueIP.IPv4Address = cloneIP(ip)
				spdrInfo.Allocated = true
			} else {
				log.Error().Msg(err.Error())
			}
		}
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

func (pdrContext *PDRCreationContext) deletePDR(spdrInfo SPDRInfo, mapOperations ebpf.ForwardingPlaneController) error {
	if spdrInfo.Ipv4 != nil {
		if err := mapOperations.DeletePdrDownlink(spdrInfo.Ipv4); err != nil {
			return fmt.Errorf("Can't delete IPv4 PDR: %s", err.Error())
		}
	} else if spdrInfo.Ipv6 != nil {
		if err := mapOperations.DeleteDownlinkPdrIp6(spdrInfo.Ipv6); err != nil {
			return fmt.Errorf("Can't delete IPv6 PDR: %s", err.Error())
		}
	} else {
		if _, ok := pdrContext.TEIDCache[uint8(spdrInfo.Teid)]; !ok {
			if err := mapOperations.DeletePdrUplink(spdrInfo.Teid); err != nil {
				return fmt.Errorf("Can't delete GTP PDR: %s", err.Error())
			}
			pdrContext.TEIDCache[uint8(spdrInfo.Teid)] = 0
		}
	}
	if spdrInfo.Teid != 0 {
		pdrContext.ResourceManager.FTEIDM.ReleaseTEID(pdrContext.Session.RemoteSEID)
	}
	return nil
}

func (pdrContext *PDRCreationContext) getFARID(farid uint32) uint32 {
	return pdrContext.Session.GetFar(farid).GlobalId
}

func (pdrContext *PDRCreationContext) getQERID(qerid uint32) uint32 {
	return pdrContext.Session.GetQer(qerid).GlobalId
}

func (pdrContext *PDRCreationContext) getFTEID(seID uint64, pdrID uint32) (uint32, error) {
	if pdrContext.ResourceManager == nil || pdrContext.ResourceManager.FTEIDM == nil {
		return 0, errors.New("FTEID manager is nil")
	}

	allocatedTeid, err := pdrContext.ResourceManager.FTEIDM.AllocateTEID(seID, pdrID)
	if err != nil {
		log.Error().Msgf("AllocateTEID err: %v", err)
		return 0, fmt.Errorf("Can't allocate TEID: %s", causeToString(ie.CauseNoResourcesAvailable))
	}
	return allocatedTeid, nil
}

func (pdrContext PDRCreationContext) getIP() (net.IP, error) {
	if pdrContext.ResourceManager == nil || pdrContext.ResourceManager.IPAM == nil {
		return nil, errors.New("IP address manager is nil")
	}
	allocatedIP, err := pdrContext.ResourceManager.IPAM.AllocateIP(pdrContext.Session.RemoteSEID)
	if err != nil {
		return nil, fmt.Errorf("can't allocate IP: %s", causeToString(ie.CauseNoResourcesAvailable))
	}
	return allocatedIP, nil
}

func (pdrContext *PDRCreationContext) hasTEIDCache(chooseID uint8) (uint32, bool) {
	teid, ok := pdrContext.TEIDCache[chooseID]
	return teid, ok
}

func (pdrContext *PDRCreationContext) setTEIDCache(chooseID uint8, teid uint32) {
	pdrContext.TEIDCache[chooseID] = teid
}

func hasCHV4(flags uint8) bool {
	return flags&(1<<4) != 0
}
