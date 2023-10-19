package domain

import (
	"github.com/edgecomllc/eupf/components/ebpf"
	"net"
)

type Session struct {
	LocalSEID  uint64
	RemoteSEID uint64
	PDRs       map[uint32]SPDRInfo
	FARs       map[uint32]SFarInfo
	QERs       map[uint32]SQerInfo
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
