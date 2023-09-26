package core

import (
	"fmt"
	"net"
	"strconv"
	"strings"

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

func (s *Session) GetPDR(id uint16) SPDRInfo {
	return s.PDRs[uint32(id)]
}

func (s *Session) RemovePDR(id uint32) SPDRInfo {
	sPdrInfo := s.PDRs[id]
	delete(s.PDRs, id)
	return sPdrInfo
}

func ParseSdfFilter(flowDescription string) (ebpf.SdfFilter, error) {
	splitted := strings.Split(flowDescription, " ")
	if splitted[0] == "deny" {
		return ebpf.SdfFilter{}, fmt.Errorf("SDF Filter: <deny> not supported.")
	}
	if splitted[1] == "in" {
		return ebpf.SdfFilter{}, fmt.Errorf("SDF Filter: <in> not supported.")
	}
	var err error
	sdfInfo := ebpf.SdfFilter{}
	if sdfInfo.Protocol, err = ParseProtocol(splitted[2]); err != nil {
		return ebpf.SdfFilter{}, err
	}
	if splitted[4] == "any" {
		sdfInfo.SrcAddress = ebpf.IpWMask{Type: 0}
	} else {
		if sdfInfo.SrcAddress, err = ParseCidrIp(splitted[4]); err != nil {
			return ebpf.SdfFilter{}, err
		}
	}
	var offset int
	sdfInfo.SrcPortRange = ebpf.PortRange{LowerBound: 0, UpperBound: 65535}
	if splitted[5] != "to" {
		if sdfInfo.SrcPortRange, err = ParsePortRange(splitted[5]); err != nil {
			return ebpf.SdfFilter{}, err
		}
		offset += 1
	}
	if splitted[6+offset] == "assigned" {
		return ebpf.SdfFilter{}, fmt.Errorf("SDF Filter: <assigned> not supported.")
	} else {
		if sdfInfo.DstAddress, err = ParseCidrIp(splitted[6+offset]); err != nil {
			return ebpf.SdfFilter{}, err
		}
	}
	sdfInfo.DstPortRange = ebpf.PortRange{LowerBound: 0, UpperBound: 65535}
	if len(splitted) > 7+offset {
		if splitted[7+offset] != "option" {
			if sdfInfo.DstPortRange, err = ParsePortRange(splitted[7+offset]); err != nil {
				return ebpf.SdfFilter{}, err
			}
			// offset += 1
		} else {
			// splitted[8 + offset] - option field
			return ebpf.SdfFilter{}, fmt.Errorf("SDF Filter: <option> not supported.")
		}
	}
	return sdfInfo, nil
}

func ParseProtocol(protocol string) (uint8, error) {
	protocolMap := map[string]uint8{
		"icmp": 0,
		"ip":   1,
		"tcp":  2,
		"udp":  3,
	}
	number, ok := protocolMap[protocol]
	if ok {
		return number, nil
	} else {
		return 0, fmt.Errorf("Unsupported protocol.")
	}
}

func ParseCidrIp(str string) (ebpf.IpWMask, error) {
	var ipType uint8
	if i := strings.Index(str, "/"); i < 0 {
		if ip := net.ParseIP(str); ip != nil {
			if ip.To4() != nil {
				ipType = 1
				ip = ip.To4()
			} else {
				ipType = 2
			}
			return ebpf.IpWMask{Type: ipType, Ip: ip, Mask: nil}, nil
		} else {
			return ebpf.IpWMask{}, fmt.Errorf("Bad IP formatting.")
		}
	} else {
		if _, ipNet, err := net.ParseCIDR(str); err == nil {
			if ipNet.IP.To4() != nil {
				ipType = 1
				ipNet.IP = ipNet.IP.To4()
			} else {
				ipType = 2
			}
			return ebpf.IpWMask{Type: ipType, Ip: ipNet.IP, Mask: ipNet.Mask}, nil
		} else {
			return ebpf.IpWMask{}, err
		}
	}
}

func ParsePortRange(str string) (ebpf.PortRange, error) {
	splittedPortRange := strings.Split(str, "-")
	portRange := ebpf.PortRange{}
	var err error
	if len(splittedPortRange) == 2 {
		if portRange.LowerBound, err = ParsePort(splittedPortRange[0]); err != nil {
			return ebpf.PortRange{}, err
		}
		if portRange.UpperBound, err = ParsePort(splittedPortRange[1]); err != nil {
			return ebpf.PortRange{}, err
		}
	} else if len(splittedPortRange) == 1 {
		if portRange.LowerBound, err = ParsePort(splittedPortRange[0]); err != nil {
			return ebpf.PortRange{}, err
		}
		portRange.UpperBound = portRange.LowerBound
	} else {
		return ebpf.PortRange{}, fmt.Errorf("Bad port / port range formatting.")
	}
	return portRange, nil
}

func ParsePort(str string) (uint16, error) {
	if port64, err := strconv.ParseUint(str, 10, 64); err == nil {
		if port64 > 65535 {
			return 0, fmt.Errorf("Invalid port. Port must be inside bounds [0, 65535].")
		}
		return uint16(port64), nil
	} else {
		return 0, err
	}
}
