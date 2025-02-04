package core

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"

	"github.com/edgecomllc/eupf/cmd/ebpf"
)

func ParseSdfFilter(flowDescription string) (ebpf.SdfFilter, error) {
	re := regexp.MustCompile(`^permit (out|in) (icmp|ip|tcp|udp|\d+) from (any|[\d.]+|[\da-fA-F:]+)(?:/(\d+))?(?: (\d+|\d+-\d+))? to (assigned|any|[\d.]+|[\da-fA-F:]+)(?:/(\d+))?(?: (\d+|\d+-\d+))?$`)

	sdfInfo := ebpf.SdfFilter{}
	var err error

	match := re.FindStringSubmatch(flowDescription)
	log.Printf("Matched groups: %+q\n", match)
	if len(match) == 0 {
		return ebpf.SdfFilter{}, fmt.Errorf("SDF Filter: bad formatting. Should be compatible with regexp: %s", re.String())
	}

	if sdfInfo.Protocol, err = ParseProtocol(match[2]); err != nil {
		return ebpf.SdfFilter{}, err
	}
	if match[3] == "any" {
		if match[4] != "" {
			return ebpf.SdfFilter{}, fmt.Errorf("<any> keyword should not be used with </mask>")
		}
		sdfInfo.SrcAddress = ebpf.IpWMask{Type: 0}
	} else {
		if sdfInfo.SrcAddress, err = ParseCidrIp(match[3], match[4]); err != nil {
			return ebpf.SdfFilter{}, err
		}
	}
	sdfInfo.SrcPortRange = ebpf.PortRange{LowerBound: 0, UpperBound: 65535}
	if match[5] != "" {
		if sdfInfo.SrcPortRange, err = ParsePortRange(match[5]); err != nil {
			return ebpf.SdfFilter{}, err
		}
	}
	if match[6] == "assigned" || match[6] == "any" {
		sdfInfo.DstAddress = ebpf.IpWMask{Type: 0}
	} else {
		if sdfInfo.DstAddress, err = ParseCidrIp(match[6], match[7]); err != nil {
			return ebpf.SdfFilter{}, err
		}
	}
	sdfInfo.DstPortRange = ebpf.PortRange{LowerBound: 0, UpperBound: 65535}
	if match[8] != "" {
		if sdfInfo.DstPortRange, err = ParsePortRange(match[8]); err != nil {
			return ebpf.SdfFilter{}, err
		}
	}

	return sdfInfo, nil
}

func ParseProtocol(protocol string) (uint8, error) {
	switch protocol {
	case "58":
		protocol = "icmp6"
	case "1":
		protocol = "icmp"
	case "17":
		protocol = "udp"
	case "6":
		protocol = "tcp"
	}

	protocolMap := map[string]uint8{
		"icmp":  0,
		"ip":    1,
		"tcp":   2,
		"udp":   3,
		"icmp6": 4,
	}
	number, ok := protocolMap[protocol]
	if ok {
		return number, nil
	} else {
		return 0, fmt.Errorf("SDF filter: unsupported protocol %s", protocol)
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
		mask := net.CIDRMask(8*len(ip), 8*len(ip))
		if maskStr != "" {
			if maskUint, err := strconv.ParseUint(maskStr, 10, 64); err == nil {
				mask = net.CIDRMask(int(maskUint), 8*len(ip))
				ip = ip.Mask(mask)
			} else {
				return ebpf.IpWMask{}, fmt.Errorf("Bad IP mask formatting.")
			}
		}
		return ebpf.IpWMask{
			Type: ipType,
			Ip:   ip,
			Mask: mask,
		}, nil
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
	if portRange.LowerBound > portRange.UpperBound {
		return ebpf.PortRange{}, fmt.Errorf("Invalid port range. Left port should be less or equal to right port.")
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
