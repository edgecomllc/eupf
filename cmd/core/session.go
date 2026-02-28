package core

import (
	"errors"
	"net"

	"github.com/edgecomllc/eupf/cmd/ebpf"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Session struct {
	LocalSEID   uint64
	RemoteSEID  uint64
	IMSI        string
	MSISDN      string
	PDRs        map[uint32]SPDRInfo
	FARs        map[uint32]SFarInfo
	QERs        map[uint32]SQerInfo
	URRs        map[uint32]SUrrInfo
	URRSequence uint32
	Traced      bool
}

func NewSession(localSEID, remoteSEID uint64, IMSI, MSISDN string, traced bool) *Session {
	return &Session{
		LocalSEID:  localSEID,
		RemoteSEID: remoteSEID,
		IMSI:       IMSI,
		MSISDN:     MSISDN,
		PDRs:       map[uint32]SPDRInfo{},
		FARs:       map[uint32]SFarInfo{},
		QERs:       map[uint32]SQerInfo{},
		URRs:       map[uint32]SUrrInfo{},
		Traced:     traced,
	}
}

type SPDRInfo struct {
	PdrID           uint32
	PdrInfo         ebpf.PdrInfo
	Teid            uint32
	Ipv4            net.IP
	Ipv6            net.IP
	NetworkInstance string
	SourceInterface uint8
	Allocated       bool
	PCCInfo         *PCCInfo
}

type PCCInfo struct {
	PCCName      string
	Notify       bool
	RawSDFFilter string
	SDFFilter    ebpf.SdfFilter
	FAR          ebpf.FarInfo
	QER          ebpf.QerInfo
}

type SFarInfo struct {
	FarInfo  ebpf.FarInfo
	GlobalId uint32
}

type SQerInfo struct {
	QerInfo  ebpf.QerInfo
	GlobalId uint32
}

type SUrrInfo struct {
	UrrInfo         ebpf.UrrInfo
	GlobalId        uint32
	ReportSeqNumber uint32
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

func (s *Session) RemoveFar(id uint32) (SFarInfo, error) {
	sFarInfo, ok := s.FARs[id]
	if !ok {
		return SFarInfo{}, ErrSessionNotFound
	}
	delete(s.FARs, id)
	return sFarInfo, nil
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

func (s *Session) RemoveQer(id uint32) (SQerInfo, error) {
	sQerInfo, ok := s.QERs[id]
	if !ok {
		return SQerInfo{}, ErrSessionNotFound
	}
	delete(s.QERs, id)
	return sQerInfo, nil
}

func (s *Session) NewUrr(id uint32, internalId uint32, urrInfo ebpf.UrrInfo) {
	s.URRs[id] = SUrrInfo{
		UrrInfo:  urrInfo,
		GlobalId: internalId,
	}
}

func (s *Session) UpdateUrr(id uint32, urrInfo ebpf.UrrInfo) {
	sUrrInfo := s.URRs[id]
	sUrrInfo.UrrInfo = urrInfo
	s.URRs[id] = sUrrInfo
}

func (s *Session) GetUrr(id uint32) SUrrInfo {
	return s.URRs[id]
}

func (s *Session) RemoveUrr(id uint32) (SUrrInfo, error) {
	sUrrInfo, ok := s.URRs[id]
	if !ok {
		return SUrrInfo{}, ErrSessionNotFound
	}
	delete(s.URRs, id)
	return sUrrInfo, nil
}

func (s *Session) PutPDR(id uint32, info SPDRInfo) {
	s.PDRs[id] = info
}

func (s *Session) GetPDR(id uint16) SPDRInfo {
	return s.PDRs[uint32(id)]
}

func (s *Session) RemovePDR(id uint32) (SPDRInfo, error) {
	sPdrInfo, ok := s.PDRs[id]
	if !ok {
		return SPDRInfo{}, ErrSessionNotFound
	}
	delete(s.PDRs, id)
	return sPdrInfo, nil
}

func (s *Session) GetSessionImsi() string {
	return s.IMSI
}

func (s *Session) GetSessionMsisdn() string {
	return s.MSISDN
}

func (s *Session) IsSessionTraced() bool {
	return s.Traced
}
