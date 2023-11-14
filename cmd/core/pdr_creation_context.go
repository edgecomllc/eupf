package core

import (
	"fmt"
	"github.com/edgecomllc/eupf/cmd/core/service"
	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
	"net"
)

type PDRCreationContext struct {
	Session         *Session
	ResourceManager *service.ResourceManager
	TEIDCache       map[uint8]uint32
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
		log.Info().Msgf("[ERROR] AllocateTEID err: %v", err)
		return 0, fmt.Errorf("Can't allocate TEID: %s", causeToString(ie.CauseNoResourcesAvailable))
	}
	return allocatedTeid, nil
}

func (pcc *PDRCreationContext) getUEIP(seID uint64) (net.IP, error) {
	ip, err := pcc.ResourceManager.IPAM.AllocateIP(seID)
	if err != nil {
		log.Info().Msgf("[ERROR] AllocateIP err: %v", err)
		return nil, fmt.Errorf("Can't allocate IP: %s", causeToString(ie.CauseNoResourcesAvailable))
	}
	return ip, nil
}
