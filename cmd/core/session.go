package core

import (
	"fmt"
	"net"
	"regexp"
	"strconv"

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
	re := regexp.MustCompile(`^permit out (icmp|ip|tcp|udp) from (any|[\d.]+|[\da-fA-F:]+)(?:/(\d+))?(?: (\d+|\d+-\d+))? to ([\d.]+|[\da-fA-F:]+)(?:/(\d+))?(?: (\d+|\d+-\d+))?$`)
	match := re.FindStringSubmatch(flowDescription)
	if len(match) == 0 {
		return ebpf.SdfFilter{}, fmt.Errorf("SDF Filter: bad formatting. Check for compatibility with regexp.")
	}
	var err error
	sdfInfo := ebpf.SdfFilter{}
	if sdfInfo.Protocol, err = ParseProtocol(match[1]); err != nil {
		return ebpf.SdfFilter{}, err
	}
	if match[2] == "any" {
		sdfInfo.SrcAddress = ebpf.IpWMask{Type: 0}
	} else {
		if sdfInfo.SrcAddress, err = ParseCidrIp(match[2], match[3]); err != nil {
			return ebpf.SdfFilter{}, err
		}
	}
	sdfInfo.SrcPortRange = ebpf.PortRange{LowerBound: 0, UpperBound: 65535}
	if match[4] != "" {
		if sdfInfo.SrcPortRange, err = ParsePortRange(match[4]); err != nil {
			return ebpf.SdfFilter{}, err
		}
	}
	if sdfInfo.DstAddress, err = ParseCidrIp(match[5], match[6]); err != nil {
		return ebpf.SdfFilter{}, err
	}
	sdfInfo.DstPortRange = ebpf.PortRange{LowerBound: 0, UpperBound: 65535}
	if match[7] != "" {
		if sdfInfo.DstPortRange, err = ParsePortRange(match[7]); err != nil {
			return ebpf.SdfFilter{}, err
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

func ParseCidrIp(ipStr, maskStr string) (ebpf.IpWMask, error) {
	var ipType uint8
	if ip := net.ParseIP(ipStr); ip != nil {
		if ip.To4() != nil {
			ipType = 1
			ip = ip.To4()
		} else {
			ipType = 2
		}
		var mask net.IPMask
		if maskStr != "" {
			if maskUint, err := strconv.ParseUint(maskStr, 10, 64); err == nil {
				mask = net.CIDRMask(int(maskUint), 8*len(ip))
				ip = ip.Mask(mask)
			} else {
				return ebpf.IpWMask{}, fmt.Errorf("Bad IP mask formatting.")
			}
		}
		return ebpf.IpWMask{Type: ipType, Ip: ip, Mask: mask}, nil
	} else {
		return ebpf.IpWMask{}, fmt.Errorf("Bad IP formatting.")
	}
}

func ParsePortRange(str string) (ebpf.PortRange, error) {
	re := regexp.MustCompile(`^(\d+)(?:-(\d+))?$`)
	match := re.FindStringSubmatch(str)
	portRange := ebpf.PortRange{}
	var err error
	if portRange.LowerBound, err = ParsePort(match[1]); err != nil {
		return ebpf.PortRange{}, err
	}
	if match[2] != "" {
		if portRange.UpperBound, err = ParsePort(match[2]); err != nil {
			return ebpf.PortRange{}, err
		}
	} else {
		portRange.UpperBound = portRange.LowerBound
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
