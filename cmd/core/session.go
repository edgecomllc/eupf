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

func (s *Session) FindDefaultPdrId(teid uint32, ipv4 net.IP, ipv6 net.IP) (uint16, error) {
	for pdrId := range s.PDRs {
		if s.PDRs[pdrId].PdrInfo.SdfFilter != nil {
			continue
		}
		// Compare all three fields (TEID, IPv4, IPv6),
		// because TEID's default value is 0, even not specifically set.
		if s.PDRs[pdrId].Teid == teid &&
			(s.PDRs[pdrId].Ipv4 == nil && ipv4 == nil ||
				s.PDRs[pdrId].Ipv4 != nil && ipv4 != nil && s.PDRs[pdrId].Ipv4.Equal(ipv4)) &&
			(s.PDRs[pdrId].Ipv6 == nil && ipv6 == nil ||
				s.PDRs[pdrId].Ipv6 != nil && ipv6 != nil && s.PDRs[pdrId].Ipv6.Equal(ipv6)) {
			return uint16(pdrId), nil
		}
	}
	return 0, fmt.Errorf("Default PDR not found by 3-tuple: teid: %d, ipv4: %v, ipv6: %v", teid, ipv4, ipv6)
}

func (s *Session) GetPDR(id uint16) SPDRInfo {
	return s.PDRs[uint32(id)]
}

func (s *Session) RemovePDR(id uint32) SPDRInfo {
	sPdrInfo := s.PDRs[id]
	delete(s.PDRs, id)
	return sPdrInfo
}
