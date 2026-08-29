//go:build !linux

package ebpf

import (
	"errors"

	"github.com/cilium/ebpf"
)

type IpEntrypointObjects struct {
	UpfIpEntrypointFunc  *ebpf.Program
	FarMap               *ebpf.Map
	QerMap               *ebpf.Map
	UrrMap               *ebpf.Map
	PdrMapDownlinkIp4    *ebpf.Map
	PdrMapDownlinkIp6    *ebpf.Map
	PdrMapTeidIp4        *ebpf.Map
	UpfExtStat           *ebpf.Map
	UpfRouteStat         *ebpf.Map
}

func (o *IpEntrypointObjects) Close() error { return nil }

type IpEntrypointPdrInfo struct {
	DefaultPdr struct {
		OuterHeaderRemoval uint8
		FarId              uint32
		QerId              uint32
		UrrId              [2]uint32
	}
	TraceFlag     uint8
	NrFlag        uint8
	SdfMode       uint8
	DedicatedPdrs [8]struct {
		Pdr struct {
			OuterHeaderRemoval uint8
			FarId              uint32
			QerId              uint32
			UrrId              [2]uint32
		}
		SdfFilter struct {
			Protocol      uint8
			SrcAddr       ipEntrypointIpWMask
			SrcPort       ipEntrypointPortRange
			DstAddr       ipEntrypointIpWMask
			DstPort       ipEntrypointPortRange
			Bidirectional uint8
			SpfFlag       uint8
		}
		Notify uint8
	}
}

type ipEntrypointIpWMask struct {
	Type uint8
	Ip   [16]byte
	Mask [16]byte
}

type ipEntrypointPortRange struct {
	LowerBound uint16
	UpperBound uint16
}

type IpEntrypointFarInfo struct {
	Action                uint8
	OuterHeaderCreation   uint8
	Teid                  uint32
	Remoteip              uint32
	Trigger               uint8
	TransportLevelMarking uint16
}

type IpEntrypointQerInfo struct {
	UlGateStatus     uint8
	DlGateStatus     uint8
	Qfi              uint8
	Dscp             uint8
	UlMaximumBitrate uint64
	DlMaximumBitrate uint64
	UlStart          uint64
	DlStart          uint64
}

type IpEntrypointUrrInfo struct {
	Ul uint64
	Dl uint64
}

type IpEntrypointUpfStatistic struct {
	XdpActions [16]uint64
	UpfCounters UpfCounters
}

type IpEntrypointRouteStat struct {
	FibLookupIp4Cache     uint64
	FibLookupIp4Ok        uint64
	FibLookupIp4ErrorDrop uint64
	FibLookupIp4ErrorPass uint64
	FibLookupIp6Cache     uint64
	FibLookupIp6Ok        uint64
	FibLookupIp6ErrorDrop uint64
	FibLookupIp6ErrorPass uint64
}

func LoadIpEntrypoint() (*ebpf.CollectionSpec, error) {
	return nil, errors.New("ebpf generation requires Linux")
}
