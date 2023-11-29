package core

import (
	"errors"
	"fmt"
	"github.com/edgecomllc/eupf/cmd/core/service"
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

func (pcc *PDRCreationContext) extractPDR(pdr *ie.IE, spdrInfo *SPDRInfo) error {
	if outerHeaderRemoval, err := pdr.OuterHeaderRemovalDescription(); err == nil {
		spdrInfo.PdrInfo.OuterHeaderRemoval = outerHeaderRemoval
	}
	if farid, err := pdr.FARID(); err == nil {
		spdrInfo.PdrInfo.FarId = pcc.getFARID(farid)
	}
	if qerid, err := pdr.QERID(); err == nil {
		spdrInfo.PdrInfo.QerId = pcc.getQERID(qerid)
	}

	pdi, err := pdr.PDI()
	if err != nil {
		return fmt.Errorf("PDI IE is missing")
	}

	if sdfFilter, err := pdr.SDFFilter(); err == nil {
		if sdfFilterParsed, err := ParseSdfFilter(sdfFilter.FlowDescription); err == nil {
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
					if teidFromCache, ok := pcc.hasTEIDCache(fteid.ChooseID); ok {
						allocate = false
						teid = teidFromCache
						spdrInfo.Allocated = true
					}
				}
				if allocate {
					allocatedTeid, err := pcc.getFTEID(pcc.Session.RemoteSEID, spdrInfo.PdrID)
					if err != nil {
						log.Error().Msgf("AllocateTEID err: %v", err)
						return fmt.Errorf("can't allocate TEID: %s", causeToString(ie.CauseNoResourcesAvailable))
					}
					teid = allocatedTeid
					spdrInfo.Allocated = true
					if fteid.HasChID() {
						pcc.setTEIDCache(fteid.ChooseID, teid)
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

func (pcc *PDRCreationContext) getFARID(farid uint32) uint32 {
	return pcc.Session.GetFar(farid).GlobalId
}

func (pcc *PDRCreationContext) getQERID(qerid uint32) uint32 {
	return pcc.Session.GetQer(qerid).GlobalId
}

func (pcc *PDRCreationContext) getFTEID(seID uint64, pdrID uint32) (uint32, error) {
	if pcc.ResourceManager == nil {
		return 0, errors.New("resource manager is nil")
	} else {
		if pcc.ResourceManager.FTEIDM == nil {
			return 0, errors.New("FTEID manager is nil")
		}
	}
	allocatedTeid, err := pcc.ResourceManager.FTEIDM.AllocateTEID(seID, pdrID)
	if err != nil {
		log.Error().Msgf("AllocateTEID err: %v", err)
		return 0, fmt.Errorf("Can't allocate TEID: %s", causeToString(ie.CauseNoResourcesAvailable))
	}
	return allocatedTeid, nil
}

func (pcc *PDRCreationContext) hasTEIDCache(chooseID uint8) (uint32, bool) {
	teid, ok := pcc.TEIDCache[chooseID]
	return teid, ok
}

func (pcc *PDRCreationContext) setTEIDCache(chooseID uint8, teid uint32) {
	pcc.TEIDCache[chooseID] = teid
}
