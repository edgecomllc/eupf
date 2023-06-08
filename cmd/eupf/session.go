package main

import (
	"github.com/edgecomllc/eupf/cmd/eupf/ebpf"
	"log"
	"net"
)

type Session struct {
	LocalSEID    uint64
	RemoteSEID   uint64
	UplinkPDRs   map[uint32]SPDRInfo
	DownlinkPDRs map[uint32]SPDRInfo
	FARs         map[uint32]SFarInfo
	QERs         map[uint32]SQerInfo
}

type SPDRInfo struct {
	PdrInfo ebpf.PdrInfo
	Teid    uint32
	Ipv4    net.IP
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
	log.Printf("NewFar: id=%d, internalId=%d, farInfo=%v\n", id, internalId, farInfo)
	s.FARs[id] = SFarInfo{
		FarInfo:  farInfo,
		GlobalId: internalId,
	}
}

func (s *Session) UpdateFar(id uint32, farInfo ebpf.FarInfo) {
	log.Printf("UpdateFar: id=%d, farInfo=%v\n", id, farInfo)
	sFarInfo := s.FARs[id]
	sFarInfo.FarInfo = farInfo
	s.FARs[id] = sFarInfo
}

func (s *Session) GetFar(id uint32) SFarInfo {
	log.Printf("GetFar: id=%d\n", id)
	return s.FARs[id]
}

func (s *Session) RemoveFar(id uint32) SFarInfo {
	log.Printf("RemoveFar: id=%d\n", id)
	sFarInfo := s.FARs[id]
	delete(s.FARs, id)
	return sFarInfo
}

func (s *Session) NewQer(id uint32, internalId uint32, qerInfo ebpf.QerInfo) {
	log.Printf("NewQer: id=%d, internalId=%d, qerInfo=%v\n", id, internalId, qerInfo)
	s.QERs[id] = SQerInfo{
		QerInfo:  qerInfo,
		GlobalId: internalId,
	}
}

func (s *Session) UpdateQer(id uint32, qerInfo ebpf.QerInfo) {
	log.Printf("UpdateQer: id=%d, qerInfo=%v\n", id, qerInfo)
	sQerInfo := s.QERs[id]
	sQerInfo.QerInfo = qerInfo
	s.QERs[id] = sQerInfo
}

func (s *Session) GetQer(id uint32) SQerInfo {
	log.Printf("GetQer: id=%d\n", id)
	return s.QERs[id]
}

func (s *Session) RemoveQer(id uint32) SQerInfo {
	log.Printf("RemoveQer: id=%d\n", id)
	sQerInfo := s.QERs[id]
	delete(s.QERs, id)
	return sQerInfo
}

func (s *Session) PutUplinkPDR(id uint32, info SPDRInfo) {
	log.Printf("PutUplinkPDR: id=%d, info=%v\n", id, info)
	s.UplinkPDRs[id] = info
}

func (s *Session) GetUplinkPDR(id uint16) SPDRInfo {
	log.Printf("GetUplinkPDR: id=%d\n", id)
	return s.UplinkPDRs[uint32(id)]
}

func (s *Session) RemoveUplinkPDR(id uint32) SPDRInfo {
	log.Printf("RemoveUplinkPDR: id=%d\n", id)
	sPdrInfo := s.UplinkPDRs[id]
	delete(s.UplinkPDRs, id)
	return sPdrInfo
}

func (s *Session) PutDownlinkPDR(id uint32, info SPDRInfo) {
	log.Printf("PutDownlinkPDR: id=%d, info=%v\n", id, info)
	s.DownlinkPDRs[id] = info
}

func (s *Session) GetDownlinkPDR(id uint16) SPDRInfo {
	log.Printf("GetDownlinkPDR: id=%d\n", id)
	return s.DownlinkPDRs[uint32(id)]
}

func (s *Session) RemoveDownlinkPDR(id uint32) SPDRInfo {
	log.Printf("RemoveDownlinkPDR: id=%d\n", id)
	sPdrInfo := s.DownlinkPDRs[id]
	delete(s.DownlinkPDRs, id)
	return sPdrInfo
}
