package core

import (
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

func (pcc *PDRCreationContext) getFARID(farid uint32) uint32 {
	return pcc.Session.GetFar(farid).GlobalId
}

func (pcc *PDRCreationContext) getQERID(qerid uint32) uint32 {
	return pcc.Session.GetQer(qerid).GlobalId
}

func (pcc *PDRCreationContext) getFTEID(seID uint64, pdrID uint32) (uint32, error) {
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
