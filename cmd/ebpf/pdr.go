package ebpf

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
	"unsafe"

	"github.com/cilium/ebpf"
)

// The BPF_ARRAY map type has no delete operation. The only way to delete an element is to replace it with a new one.

type PdrInfo struct {
	OuterHeaderRemoval uint8
	FarId              uint32
	QerId              uint32
	SdfFilter          *SdfFilter
}

type SdfFilter struct {
	Protocol     uint8 // 0: icmp, 1: ip, 2: tcp, 3: udp
	SrcAddress   IpWMask
	SrcPortRange PortRange
	DstAddress   IpWMask
	DstPortRange PortRange
}

type IpWMask struct {
	Type uint8 // 0: any, 1: ip4, 2: ip6
	Ip   net.IP
	Mask net.IPMask
}

type PortRange struct {
	LowerBound uint16
	UpperBound uint16
}

func HandlePdrWithSdf(lookup func(interface{}, interface{}) error, key interface{}, pdrInfo PdrInfo) (*IpEntrypointPdrInfo, error) {
	var pdrToStore IpEntrypointPdrInfo
	if pdrInfo.SdfFilter != nil {
		var defaultPdr IpEntrypointPdrInfo
		if err := lookup(key, &defaultPdr); err == nil {
			pdrToStore = CombinePdrWithSdf(defaultPdr, pdrInfo)
		} else {
			return nil, err
		}
	} else {
		pdrToStore = ToIpEntrypointPdrInfo(pdrInfo)
	}
	return &pdrToStore, nil
}

func (bpfObjects *BpfObjects) PutPdrUpLink(teid uint32, pdrInfo PdrInfo) error {
	log.Printf("EBPF: Put PDR Uplink: teid=%d, pdrInfo=%+v", teid, pdrInfo)
	if pdrToStore, err := HandlePdrWithSdf(bpfObjects.PdrMapUplinkIp4.Lookup, teid, pdrInfo); err == nil {
		return bpfObjects.PdrMapUplinkIp4.Put(teid, unsafe.Pointer(pdrToStore))
	} else {
		return err
	}
}

func (bpfObjects *BpfObjects) PutPdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error {
	log.Printf("EBPF: Put PDR Downlink: ipv4=%s, pdrInfo=%+v", ipv4, pdrInfo)
	if pdrToStore, err := HandlePdrWithSdf(bpfObjects.PdrMapDownlinkIp4.Lookup, ipv4, pdrInfo); err == nil {
		return bpfObjects.PdrMapDownlinkIp4.Put(ipv4, unsafe.Pointer(pdrToStore))
	} else {
		return err
	}
}

func (bpfObjects *BpfObjects) UpdatePdrUpLink(teid uint32, pdrInfo PdrInfo) error {
	log.Printf("EBPF: Update PDR Uplink: teid=%d, pdrInfo=%+v", teid, pdrInfo)
	if pdrToStore, err := HandlePdrWithSdf(bpfObjects.PdrMapUplinkIp4.Lookup, teid, pdrInfo); err == nil {
		return bpfObjects.PdrMapUplinkIp4.Update(teid, unsafe.Pointer(pdrToStore), ebpf.UpdateExist)
	} else {
		return err
	}
}

func (bpfObjects *BpfObjects) UpdatePdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error {
	log.Printf("EBPF: Update PDR Downlink: ipv4=%s, pdrInfo=%+v", ipv4, pdrInfo)
	if pdrToStore, err := HandlePdrWithSdf(bpfObjects.PdrMapDownlinkIp4.Lookup, ipv4, pdrInfo); err == nil {
		return bpfObjects.PdrMapDownlinkIp4.Update(ipv4, unsafe.Pointer(pdrToStore), ebpf.UpdateExist)
	} else {
		return err
	}
}

func (bpfObjects *BpfObjects) DeletePdrUpLink(teid uint32) error {
	log.Printf("EBPF: Delete PDR Uplink: teid=%d", teid)
	return bpfObjects.PdrMapUplinkIp4.Update(teid, unsafe.Pointer(&PdrInfo{}), ebpf.UpdateExist)
	//return o.PdrMapUplinkIp4.Delete(teid)
}

func (bpfObjects *BpfObjects) DeletePdrDownLink(ipv4 net.IP) error {
	log.Printf("EBPF: Delete PDR Downlink: ipv4=%s", ipv4)
	return bpfObjects.PdrMapDownlinkIp4.Update(ipv4, unsafe.Pointer(&PdrInfo{}), ebpf.UpdateExist)
	//return o.PdrMapDownlinkIp4.Delete(ipv4)
}

func (bpfObjects *BpfObjects) PutDownlinkPdrIp6(ipv6 net.IP, pdrInfo PdrInfo) error {
	log.Printf("EBPF: Put PDR Ipv6 Downlink: ipv6=%s, pdrInfo=%+v", ipv6, pdrInfo)
	if pdrToStore, err := HandlePdrWithSdf(bpfObjects.PdrMapDownlinkIp6.Lookup, ipv6, pdrInfo); err == nil {
		return bpfObjects.PdrMapDownlinkIp6.Put(ipv6, unsafe.Pointer(&pdrToStore))
	} else {
		return err
	}
}

func (bpfObjects *BpfObjects) UpdateDownlinkPdrIp6(ipv6 net.IP, pdrInfo PdrInfo) error {
	log.Printf("EBPF: Update PDR Ipv6 Downlink: ipv6=%s, pdrInfo=%+v", ipv6, pdrInfo)
	if pdrToStore, err := HandlePdrWithSdf(bpfObjects.PdrMapDownlinkIp6.Lookup, ipv6, pdrInfo); err == nil {
		return bpfObjects.PdrMapDownlinkIp6.Update(ipv6, unsafe.Pointer(&pdrToStore), ebpf.UpdateExist)
	} else {
		return err
	}
}

func (bpfObjects *BpfObjects) DeleteDownlinkPdrIp6(ipv6 net.IP) error {
	log.Printf("EBPF: Delete PDR Ipv6 Downlink: ipv6=%s", ipv6)
	return bpfObjects.PdrMapDownlinkIp6.Delete(ipv6)
}

type FarInfo struct {
	Action                uint8
	OuterHeaderCreation   uint8
	Teid                  uint32
	RemoteIP              uint32
	LocalIP               uint32
	TransportLevelMarking uint16
}

func (f FarInfo) MarshalJSON() ([]byte, error) {
	remoteIP := make(net.IP, 4)
	localIP := make(net.IP, 4)
	binary.LittleEndian.PutUint32(remoteIP, f.RemoteIP)
	binary.LittleEndian.PutUint32(localIP, f.LocalIP)
	data := map[string]interface{}{
		"action":                  f.Action,
		"outer_header_creation":   f.OuterHeaderCreation,
		"teid":                    f.Teid,
		"remote_ip":               remoteIP.String(),
		"local_ip":                localIP.String(),
		"transport_level_marking": f.TransportLevelMarking,
	}
	return json.Marshal(data)
}

func (bpfObjects *BpfObjects) NewFar(farInfo FarInfo) (uint32, error) {
	internalId, err := bpfObjects.FarIdTracker.GetNext()
	if err != nil {
		return 0, err
	}
	log.Printf("EBPF: Put FAR: internalId=%d, qerInfo=%+v", internalId, farInfo)
	return internalId, bpfObjects.FarMap.Put(internalId, unsafe.Pointer(&farInfo))
}

func (bpfObjects *BpfObjects) UpdateFar(internalId uint32, farInfo FarInfo) error {
	log.Printf("EBPF: Update FAR: internalId=%d, farInfo=%+v", internalId, farInfo)
	return bpfObjects.FarMap.Update(internalId, unsafe.Pointer(&farInfo), ebpf.UpdateExist)
}

func (bpfObjects *BpfObjects) DeleteFar(intenalId uint32) error {
	log.Printf("EBPF: Delete FAR: intenalId=%d", intenalId)
	bpfObjects.FarIdTracker.Release(intenalId)
	return bpfObjects.FarMap.Update(intenalId, unsafe.Pointer(&FarInfo{}), ebpf.UpdateExist)
}

type QerInfo struct {
	GateStatusUL uint8
	GateStatusDL uint8
	Qfi          uint8
	MaxBitrateUL uint32
	MaxBitrateDL uint32
	StartUL      uint64
	StartDL      uint64
}

func (bpfObjects *BpfObjects) NewQer(qerInfo QerInfo) (uint32, error) {
	internalId, err := bpfObjects.QerIdTracker.GetNext()
	if err != nil {
		return 0, err
	}
	log.Printf("EBPF: Put QER: internalId=%d, qerInfo=%+v", internalId, qerInfo)
	return internalId, bpfObjects.QerMap.Put(internalId, unsafe.Pointer(&qerInfo))
}

func (bpfObjects *BpfObjects) UpdateQer(internalId uint32, qerInfo QerInfo) error {
	log.Printf("EBPF: Update QER: internalId=%d, qerInfo=%+v", internalId, qerInfo)
	return bpfObjects.QerMap.Update(internalId, unsafe.Pointer(&qerInfo), ebpf.UpdateExist)
}

func (bpfObjects *BpfObjects) DeleteQer(internalId uint32) error {
	log.Printf("EBPF: Delete QER: internalId=%d", internalId)
	bpfObjects.QerIdTracker.Release(internalId)
	return bpfObjects.QerMap.Update(internalId, unsafe.Pointer(&QerInfo{}), ebpf.UpdateExist)
}

type ForwardingPlaneController interface {
	PutPdrUpLink(teid uint32, pdrInfo PdrInfo) error
	PutPdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error
	UpdatePdrUpLink(teid uint32, pdrInfo PdrInfo) error
	UpdatePdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error
	DeletePdrUpLink(teid uint32) error
	DeletePdrDownLink(ipv4 net.IP) error
	PutDownlinkPdrIp6(ipv6 net.IP, pdrInfo PdrInfo) error
	UpdateDownlinkPdrIp6(ipv6 net.IP, pdrInfo PdrInfo) error
	DeleteDownlinkPdrIp6(ipv6 net.IP) error
	NewFar(farInfo FarInfo) (uint32, error)
	UpdateFar(internalId uint32, farInfo FarInfo) error
	DeleteFar(internalId uint32) error
	NewQer(qerInfo QerInfo) (uint32, error)
	UpdateQer(internalId uint32, qerInfo QerInfo) error
	DeleteQer(internalId uint32) error
}

func CombinePdrWithSdf(defaultPdr IpEntrypointPdrInfo, sdfPdr PdrInfo) IpEntrypointPdrInfo {
	var pdrToStore IpEntrypointPdrInfo
	// Default mapping options.
	pdrToStore.OuterHeaderRemoval = defaultPdr.OuterHeaderRemoval
	pdrToStore.FarId = defaultPdr.FarId
	pdrToStore.QerId = defaultPdr.QerId
	// SDF mapping options.
	pdrToStore.SdfRules.SdfFilter.Protocol = sdfPdr.SdfFilter.Protocol
	pdrToStore.SdfRules.SdfFilter.SrcAddr.Type = sdfPdr.SdfFilter.SrcAddress.Type
	pdrToStore.SdfRules.SdfFilter.SrcAddr.Ip = Copy16Ip(sdfPdr.SdfFilter.SrcAddress.Ip)
	pdrToStore.SdfRules.SdfFilter.SrcAddr.Mask = Copy16Ip(sdfPdr.SdfFilter.SrcAddress.Mask)
	pdrToStore.SdfRules.SdfFilter.SrcPort.LowerBound = sdfPdr.SdfFilter.SrcPortRange.LowerBound
	pdrToStore.SdfRules.SdfFilter.SrcPort.UpperBound = sdfPdr.SdfFilter.SrcPortRange.UpperBound
	pdrToStore.SdfRules.SdfFilter.DstAddr.Type = sdfPdr.SdfFilter.DstAddress.Type
	pdrToStore.SdfRules.SdfFilter.DstAddr.Ip = Copy16Ip(sdfPdr.SdfFilter.DstAddress.Ip)
	pdrToStore.SdfRules.SdfFilter.DstAddr.Mask = Copy16Ip(sdfPdr.SdfFilter.DstAddress.Mask)
	pdrToStore.SdfRules.SdfFilter.DstPort.LowerBound = sdfPdr.SdfFilter.DstPortRange.LowerBound
	pdrToStore.SdfRules.SdfFilter.DstPort.UpperBound = sdfPdr.SdfFilter.DstPortRange.UpperBound
	pdrToStore.SdfRules.FarId = sdfPdr.FarId
	pdrToStore.SdfRules.QerId = sdfPdr.QerId
	return pdrToStore
}

func ToIpEntrypointPdrInfo(defaultPdr PdrInfo) IpEntrypointPdrInfo {
	var pdrToStore IpEntrypointPdrInfo
	pdrToStore.OuterHeaderRemoval = defaultPdr.OuterHeaderRemoval
	pdrToStore.FarId = defaultPdr.FarId
	pdrToStore.QerId = defaultPdr.QerId
	return pdrToStore
}

func Copy16Ip[T ~[]byte](arr T) [16]byte {
	const Ipv4len = 4
	const Ipv6len = 16
	var c [Ipv6len]byte
	var arrLen int
	if len(arr) == Ipv4len {
		arrLen = Ipv4len
	} else if len(arr) == Ipv6len {
		arrLen = Ipv6len
	} else if len(arr) == 0 || arr == nil {
		return c
	}
	for i := 0; i < arrLen; i++ {
		c[i] = (arr)[arrLen-1-i]
	}
	return c
}
