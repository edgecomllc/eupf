package ebpf

import (
	"net"
	"unsafe"

	"github.com/cilium/ebpf"
)

type UeRebalanceLpmKey struct {
	PrefixLen uint32
	Data      [4]byte
}

func PutUeRebalancePrefix(m *ebpf.Map, cidr string) error {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	ones, _ := ipNet.Mask.Size()
	key := UeRebalanceLpmKey{
		PrefixLen: uint32(ones),
	}
	copy(key.Data[:], ipNet.IP.To4())
	var value uint8 = 1
	return m.Update(unsafe.Pointer(&key), unsafe.Pointer(&value), ebpf.UpdateAny)
}
