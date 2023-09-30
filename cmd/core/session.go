package core

import (
	"fmt"
	"net"

	"github.com/edgecomllc/eupf/cmd/ebpf"
)

type Session struct {
	LocalSEID  uint64
	RemoteSEID uint64
	PDRs       map[uint32]SPDRInfo
	FARs       map[uint32]SFarInfo
	QERs       map[uint32]SQerInfo
}

func NewSession(localSEID uint64, remoteSEID uint64) *Session {
	return &Session{
		LocalSEID:  localSEID,
		RemoteSEID: remoteSEID,
		PDRs:       map[uint32]SPDRInfo{},
		FARs:       map[uint32]SFarInfo{},
		QERs:       map[uint32]SQerInfo{},
	}
}

type SPDRInfo struct {
	PdrInfo ebpf.PdrInfo
	Teid    uint32
	Ipv4    net.IP
	Ipv6    net.IP
}

type SFarInfo struct {
	FarInfo  ebpf.FarInfo
	GlobalId uint32
}

type SQerInfo struct {
	QerInfo  ebpf.QerInfo
	GlobalId uint32
}

func (s *Session) NewFar(id uint32, internalId uint32, farInfo ebpf.FarInfo) {
	s.FARs[id] = SFarInfo{
		FarInfo:  farInfo,
		GlobalId: internalId,
	}
}

func (s *Session) UpdateFar(id uint32, farInfo ebpf.FarInfo) {
	sFarInfo := s.FARs[id]
	sFarInfo.FarInfo = farInfo
	s.FARs[id] = sFarInfo
}

func (s *Session) GetFar(id uint32) SFarInfo {
	return s.FARs[id]
}

func (s *Session) RemoveFar(id uint32) SFarInfo {
	sFarInfo := s.FARs[id]
	delete(s.FARs, id)
	return sFarInfo
}

func (s *Session) NewQer(id uint32, internalId uint32, qerInfo ebpf.QerInfo) {
	s.QERs[id] = SQerInfo{
		QerInfo:  qerInfo,
		GlobalId: internalId,
	}
}

func (s *Session) UpdateQer(id uint32, qerInfo ebpf.QerInfo) {
	sQerInfo := s.QERs[id]
	sQerInfo.QerInfo = qerInfo
	s.QERs[id] = sQerInfo
}

func (s *Session) GetQer(id uint32) SQerInfo {
	return s.QERs[id]
}

func (s *Session) RemoveQer(id uint32) SQerInfo {
	sQerInfo := s.QERs[id]
	delete(s.QERs, id)
	return sQerInfo
}

func (s *Session) PutPDR(id uint32, info SPDRInfo) {
	s.PDRs[id] = info
}

func (s *Session) FindPdrId(info SPDRInfo) (uint32, error) {
	for pdrId := range s.PDRs {
		// Compare all three fields (TEID, IPv4, IPv6),
		// because TEID's default value is 0, even not specifically set.
		if s.PDRs[pdrId].Teid == info.Teid &&
			(s.PDRs[pdrId].Ipv4 == nil && info.Ipv4 == nil ||
				s.PDRs[pdrId].Ipv4 != nil && info.Ipv4 != nil && s.PDRs[pdrId].Ipv4.Equal(info.Ipv4)) &&
			(s.PDRs[pdrId].Ipv6 == nil && info.Ipv6 == nil ||
				s.PDRs[pdrId].Ipv6 != nil && info.Ipv6 != nil && s.PDRs[pdrId].Ipv6.Equal(info.Ipv6)) {
			return pdrId, nil
		}
	}
	return 0, fmt.Errorf("Correponding PDR not found")
}

func (s *Session) PutSdfToPdr(pdrId uint32, info ebpf.AdditionalRules) {
	pdr := s.PDRs[pdrId]
	pdr.PdrInfo.AdditionalRules = info
	s.PDRs[pdrId] = pdr
}

func (s *Session) GetPdrWithAdditionalRules(add SPDRInfo) (uint16, SPDRInfo, error) {
	if pdrId, err := s.FindPdrId(add); err == nil {
		newPdr := s.PDRs[pdrId]
		newPdr.PdrInfo.AdditionalRules = add.PdrInfo.AdditionalRules
		return uint16(pdrId), newPdr, nil
	} else {
		return 0, SPDRInfo{}, err
	}
}

func (s *Session) GetPDR(id uint16) SPDRInfo {
	return s.PDRs[uint32(id)]
}

func (s *Session) RemovePDR(id uint32) SPDRInfo {
	sPdrInfo := s.PDRs[id]
	delete(s.PDRs, id)
	return sPdrInfo
}
