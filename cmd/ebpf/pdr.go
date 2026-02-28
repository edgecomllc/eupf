package ebpf

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"unsafe"

	"github.com/cilium/ebpf"

	"github.com/rs/zerolog/log"
)

// The BPF_ARRAY map type has no delete operation. The only way to delete an element is to replace it with a new one.

type PdrInfo struct {
	OuterHeaderRemoval uint8
	FarId              uint32
	QerId              uint32
	Urr1Id             uint32
	Urr2Id             uint32
	SdfFilter          []SdfFilter
	TraceFlag          bool
	NotifyFlag         bool
}

type SdfFilter struct {
	Protocol     uint8 // 0: icmp, 1: ip, 2: tcp, 3: udp, 4: icmp6
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

func PreprocessPdrWithSdf(lookup func(interface{}, interface{}) error, key interface{}, pdrInfo PdrInfo) (IpEntrypointPdrInfo, error) {
	var defaultPdr IpEntrypointPdrInfo
	if err := lookup(key, &defaultPdr); err != nil {
		return CombinePdrWithSdf(nil, pdrInfo), nil
	}

	return CombinePdrWithSdf(&defaultPdr, pdrInfo), nil
}

func (bpfObjects *BpfObjects) PutPdrUplink(teid uint32, pdrInfo PdrInfo) error {
	log.Debug().Msgf("EBPF: Put PDR Uplink: teid=%d, pdrInfo=%+v", teid, pdrInfo)
	var pdrToStore IpEntrypointPdrInfo
	var err error
	if pdrInfo.SdfFilter != nil {
		if pdrToStore, err = PreprocessPdrWithSdf(bpfObjects.PdrMapUplinkIp4.Lookup, teid, pdrInfo); err != nil {
			return err
		}
	} else {
		pdrToStore = ToIpEntrypointPdrInfo(pdrInfo)
	}
	return bpfObjects.PdrMapUplinkIp4.Put(teid, unsafe.Pointer(&pdrToStore))
}

func (bpfObjects *BpfObjects) PutPdrDownlink(ip net.IP, pdrInfo PdrInfo) error {
	log.Debug().Msgf("EBPF: Put PDR Downlink: ip=%s, pdrInfo=%+v", ip, pdrInfo)
	var pdrToStore IpEntrypointPdrInfo
	var err error
	lookupFunc := bpfObjects.PdrMapDownlinkIp4.Lookup
	putFunc := bpfObjects.PdrMapDownlinkIp4.Put
	if len(ip) != net.IPv4len {
		lookupFunc = bpfObjects.PdrMapDownlinkIp6.Lookup
		putFunc = bpfObjects.PdrMapDownlinkIp6.Put
	}
	if pdrInfo.SdfFilter != nil {

		if pdrToStore, err = PreprocessPdrWithSdf(lookupFunc, ip, pdrInfo); err != nil {
			return err
		}
	} else {
		pdrToStore = ToIpEntrypointPdrInfo(pdrInfo)
	}

	return putFunc(ip, unsafe.Pointer(&pdrToStore))
}

func (bpfObjects *BpfObjects) UpdatePdrUplink(teid uint32, pdrInfo PdrInfo) error {
	log.Debug().Msgf("EBPF: Update PDR Uplink: teid=%d, pdrInfo=%+v", teid, pdrInfo)
	var pdrToStore IpEntrypointPdrInfo
	var err error
	if pdrInfo.SdfFilter != nil {
		if pdrToStore, err = PreprocessPdrWithSdf(bpfObjects.PdrMapUplinkIp4.Lookup, teid, pdrInfo); err != nil {
			return err
		}
	} else {
		pdrToStore = ToIpEntrypointPdrInfo(pdrInfo)
	}
	return bpfObjects.PdrMapUplinkIp4.Update(teid, unsafe.Pointer(&pdrToStore), ebpf.UpdateExist)
}

func (bpfObjects *BpfObjects) UpdatePdrDownlink(ipv4 net.IP, pdrInfo PdrInfo) error {
	log.Debug().Msgf("EBPF: Update PDR Downlink: ipv4=%s, pdrInfo=%+v", ipv4, pdrInfo)
	var pdrToStore IpEntrypointPdrInfo
	var err error
	if pdrInfo.SdfFilter != nil {
		if pdrToStore, err = PreprocessPdrWithSdf(bpfObjects.PdrMapDownlinkIp4.Lookup, ipv4, pdrInfo); err != nil {
			return err
		}
	} else {
		pdrToStore = ToIpEntrypointPdrInfo(pdrInfo)
	}
	return bpfObjects.PdrMapDownlinkIp4.Update(ipv4, unsafe.Pointer(&pdrToStore), ebpf.UpdateExist)
}

func (bpfObjects *BpfObjects) DeletePdrUplink(teid uint32) error {
	log.Debug().Msgf("EBPF: Delete PDR Uplink: teid=%d", teid)
	return bpfObjects.PdrMapUplinkIp4.Delete(teid)
}

func (bpfObjects *BpfObjects) DeletePdrDownlink(ipv4 net.IP) error {
	log.Debug().Msgf("EBPF: Delete PDR Downlink: ipv4=%s", ipv4)
	return bpfObjects.PdrMapDownlinkIp4.Delete(ipv4)
}

func (bpfObjects *BpfObjects) PutDownlinkPdrIp6(ipv6 net.IP, pdrInfo PdrInfo) error {
	log.Debug().Msgf("EBPF: Put PDR Ipv6 Downlink: ipv6=%s, pdrInfo=%+v", ipv6, pdrInfo)
	var pdrToStore IpEntrypointPdrInfo
	var err error
	if pdrInfo.SdfFilter != nil {
		if pdrToStore, err = PreprocessPdrWithSdf(bpfObjects.PdrMapDownlinkIp6.Lookup, ipv6, pdrInfo); err != nil {
			return err
		}
	} else {
		pdrToStore = ToIpEntrypointPdrInfo(pdrInfo)
	}
	return bpfObjects.PdrMapDownlinkIp6.Put(ipv6, unsafe.Pointer(&pdrToStore))
}

func (bpfObjects *BpfObjects) UpdateDownlinkPdrIp6(ipv6 net.IP, pdrInfo PdrInfo) error {
	log.Debug().Msgf("EBPF: Update PDR Ipv6 Downlink: ipv6=%s, pdrInfo=%+v", ipv6, pdrInfo)
	var pdrToStore IpEntrypointPdrInfo
	var err error
	if pdrInfo.SdfFilter != nil {
		if pdrToStore, err = PreprocessPdrWithSdf(bpfObjects.PdrMapDownlinkIp6.Lookup, ipv6, pdrInfo); err != nil {
			return err
		}
	} else {
		pdrToStore = ToIpEntrypointPdrInfo(pdrInfo)
	}
	return bpfObjects.PdrMapDownlinkIp6.Update(ipv6, unsafe.Pointer(&pdrToStore), ebpf.UpdateExist)
}

func (bpfObjects *BpfObjects) DeleteDownlinkPdrIp6(ipv6 net.IP) error {
	log.Debug().Msgf("EBPF: Delete PDR Ipv6 Downlink: ipv6=%s", ipv6)
	return bpfObjects.PdrMapDownlinkIp6.Delete(ipv6)
}

type FarInfo struct {
	Action                uint8
	OuterHeaderCreation   uint8
	Teid                  uint32
	RemoteIP              uint32
	Trigger               uint8 // trigger for applying the FAR // todo maxim: change to ringbuffer
	TransportLevelMarking uint16
}

func (f FarInfo) MarshalJSON() ([]byte, error) {
	remoteIP := make(net.IP, 4)
	binary.LittleEndian.PutUint32(remoteIP, f.RemoteIP)
	data := map[string]interface{}{
		"action":                  f.Action,
		"outer_header_creation":   f.OuterHeaderCreation,
		"teid":                    f.Teid,
		"remote_ip":               remoteIP.String(),
		"transport_level_marking": f.TransportLevelMarking,
	}
	return json.Marshal(data)
}

func (bpfObjects *BpfObjects) NewFar(farInfo FarInfo) (uint32, error) {
	internalId, err := bpfObjects.GetNextFAR()
	if err != nil {
		return 0, err
	}

	log.Debug().Msgf("EBPF: Put FAR: internalId=%d, farInfo=%+v", internalId, farInfo)

	farToStore := IpEntrypointFarInfo{
		Action:                farInfo.Action,
		OuterHeaderCreation:   farInfo.OuterHeaderCreation,
		Teid:                  farInfo.Teid,
		Remoteip:              farInfo.RemoteIP,
		TransportLevelMarking: farInfo.TransportLevelMarking,
	}
	return internalId, bpfObjects.FarMap.Put(internalId, unsafe.Pointer(&farToStore))
}

func (bpfObjects *BpfObjects) GetFar(internalId uint32) (FarInfo, error) {
	log.Debug().Msgf("EBPF: Get FAR: internalId=%d", internalId)

	farToStore := IpEntrypointFarInfo{}
	if err := bpfObjects.FarMap.Lookup(internalId, unsafe.Pointer(&farToStore)); err != nil {
		return FarInfo{}, err
	}

	farInfo := FarInfo{
		Action:                farToStore.Action,
		OuterHeaderCreation:   farToStore.OuterHeaderCreation,
		Teid:                  farToStore.Teid,
		RemoteIP:              farToStore.Remoteip,
		Trigger:               farToStore.Trigger,
		TransportLevelMarking: farToStore.TransportLevelMarking,
	}

	log.Debug().Msgf("EBPF: Get FAR: internalId=%d, farInfo=%+v", internalId, farInfo)
	return farInfo, nil
}

func (bpfObjects *BpfObjects) UpdateFar(internalId uint32, farInfo FarInfo) error {
	log.Debug().Msgf("EBPF: Update FAR: internalId=%d, farInfo=%+v", internalId, farInfo)

	farToStore := IpEntrypointFarInfo{
		Action:                farInfo.Action,
		OuterHeaderCreation:   farInfo.OuterHeaderCreation,
		Teid:                  farInfo.Teid,
		Remoteip:              farInfo.RemoteIP,
		TransportLevelMarking: farInfo.TransportLevelMarking,
	}
	return bpfObjects.FarMap.Update(internalId, unsafe.Pointer(&farToStore), ebpf.UpdateExist)
}

func (bpfObjects *BpfObjects) DeleteFar(intenalId uint32) error {
	log.Debug().Msgf("EBPF: Delete FAR: intenalId=%d", intenalId)
	bpfObjects.ReleaseFAR(intenalId)
	return bpfObjects.FarMap.Update(intenalId, unsafe.Pointer(&FarInfo{}), ebpf.UpdateExist)
}

type QerInfo struct {
	GateStatusUL uint8
	GateStatusDL uint8
	Qfi          uint8
	Dscp         uint8
	MaxBitrateUL uint64
	MaxBitrateDL uint64
}

func (bpfObjects *BpfObjects) NewQer(qerInfo QerInfo) (uint32, error) {
	internalId, err := bpfObjects.GetNextQER()
	if err != nil {
		return 0, err
	}
	log.Debug().Msgf("EBPF: Put QER: internalId=%d, qerInfo=%+v", internalId, qerInfo)

	qerToStore := IpEntrypointQerInfo{
		UlGateStatus:     qerInfo.GateStatusUL,
		DlGateStatus:     qerInfo.GateStatusDL,
		Qfi:              qerInfo.Qfi,
		Dscp:             qerInfo.Dscp,
		UlMaximumBitrate: qerInfo.MaxBitrateUL,
		DlMaximumBitrate: qerInfo.MaxBitrateDL,
		UlStart:          0,
		DlStart:          0,
	}
	return internalId, bpfObjects.QerMap.Put(internalId, unsafe.Pointer(&qerToStore))
}

func (bpfObjects *BpfObjects) UpdateQer(internalId uint32, qerInfo QerInfo) error {
	log.Debug().Msgf("EBPF: Update QER: internalId=%d, qerInfo=%+v", internalId, qerInfo)

	qerToStore := IpEntrypointQerInfo{
		UlGateStatus:     qerInfo.GateStatusUL,
		DlGateStatus:     qerInfo.GateStatusDL,
		Qfi:              qerInfo.Qfi,
		Dscp:             qerInfo.Dscp,
		UlMaximumBitrate: qerInfo.MaxBitrateUL,
		DlMaximumBitrate: qerInfo.MaxBitrateDL,
		UlStart:          0,
		DlStart:          0,
	}
	return bpfObjects.QerMap.Update(internalId, unsafe.Pointer(&qerToStore), ebpf.UpdateExist)
}

func (bpfObjects *BpfObjects) DeleteQer(internalId uint32) error {
	log.Debug().Msgf("EBPF: Delete QER: internalId=%d", internalId)
	bpfObjects.ReleaseQER(internalId)
	return bpfObjects.QerMap.Update(internalId, unsafe.Pointer(&QerInfo{}), ebpf.UpdateExist)
}

// TODO: add required fields and implement methods
type UrrInfo struct {
	UplinkVolume    uint64
	DownlinkVolume  uint64
	VolumeThreshold uint64
	TimeThreshold   float64
}

func (bpfObjects *BpfObjects) NewUrr(urrInfo UrrInfo) (uint32, error) {
	internalId, err := bpfObjects.GetNextURR()
	if err != nil {
		return 0, err
	}
	log.Debug().Msgf("EBPF: Put URR: internalId=%d, urrInfo=%+v", internalId, urrInfo)

	urrToStore := IpEntrypointUrrInfo{
		Ul: urrInfo.UplinkVolume,
		Dl: urrInfo.DownlinkVolume,
	}

	return internalId, bpfObjects.UrrMap.Put(internalId, unsafe.Pointer(&urrToStore))
}

func (bpfObjects *BpfObjects) UpdateUrr(internalId uint32, urrInfo UrrInfo) error {
	log.Debug().Msgf("EBPF: Update URR: internalId=%d, urrInfo=%+v", internalId, urrInfo)

	urrToStore := IpEntrypointUrrInfo{
		Ul: urrInfo.UplinkVolume,
		Dl: urrInfo.DownlinkVolume,
	}
	return bpfObjects.UrrMap.Update(internalId, unsafe.Pointer(&urrToStore), ebpf.UpdateExist)
}

func (bpfObjects *BpfObjects) GetUrr(internalId uint32) (UrrInfo, error) {
	log.Debug().Msgf("EBPF: Get URR: internalId=%d", internalId)

	urrToStore := IpEntrypointUrrInfo{}
	if err := bpfObjects.UrrMap.Lookup(internalId, unsafe.Pointer(&urrToStore)); err != nil {
		return UrrInfo{}, err
	}

	urrInfo := UrrInfo{
		UplinkVolume:   urrToStore.Ul,
		DownlinkVolume: urrToStore.Dl,
	}
	return urrInfo, nil
}

func (bpfObjects *BpfObjects) DeleteUrr(internalId uint32) (UrrInfo, error) {
	log.Debug().Msgf("EBPF: Delete URR: internalId=%d", internalId)

	urrToStore := IpEntrypointUrrInfo{}
	if err := bpfObjects.UrrMap.Lookup(internalId, unsafe.Pointer(&urrToStore)); err != nil {
		return UrrInfo{}, err
	}
	bpfObjects.ReleaseURR(internalId)
	if err := bpfObjects.UrrMap.Update(internalId, unsafe.Pointer(&IpEntrypointUrrInfo{}), ebpf.UpdateExist); err != nil {
		return UrrInfo{}, err
	}

	urrInfo := UrrInfo{
		UplinkVolume:   urrToStore.Ul,
		DownlinkVolume: urrToStore.Dl,
	}
	return urrInfo, nil
}

type ForwardingPlaneController interface {
	PutPdrUplink(teid uint32, pdrInfo PdrInfo) error
	PutPdrDownlink(ipv4 net.IP, pdrInfo PdrInfo) error
	PutDownlinkPdrIp6(ipv6 net.IP, pdrInfo PdrInfo) error
	UpdatePdrUplink(teid uint32, pdrInfo PdrInfo) error
	UpdatePdrDownlink(ipv4 net.IP, pdrInfo PdrInfo) error
	DeletePdrUplink(teid uint32) error
	DeletePdrDownlink(ipv4 net.IP) error
	UpdateDownlinkPdrIp6(ipv6 net.IP, pdrInfo PdrInfo) error
	DeleteDownlinkPdrIp6(ipv6 net.IP) error
	NewFar(farInfo FarInfo) (uint32, error)
	GetFar(internalId uint32) (FarInfo, error)
	UpdateFar(internalId uint32, farInfo FarInfo) error
	DeleteFar(internalId uint32) error
	NewQer(qerInfo QerInfo) (uint32, error)
	UpdateQer(internalId uint32, qerInfo QerInfo) error
	DeleteQer(internalId uint32) error
	NewUrr(urrInfo UrrInfo) (uint32, error)
	UpdateUrr(internalId uint32, urrInfo UrrInfo) error
	GetUrr(internalId uint32) (UrrInfo, error)
	DeleteUrr(internalId uint32) (UrrInfo, error)
}

func CombinePdrWithSdf(defaultPdr *IpEntrypointPdrInfo, sdfPdr PdrInfo) IpEntrypointPdrInfo {
	var pdrToStore IpEntrypointPdrInfo
	// Default mapping options.
	if defaultPdr != nil {
		pdrToStore.DefaultPdr.OuterHeaderRemoval = defaultPdr.DefaultPdr.OuterHeaderRemoval
		pdrToStore.DefaultPdr.FarId = defaultPdr.DefaultPdr.FarId
		pdrToStore.DefaultPdr.QerId = defaultPdr.DefaultPdr.QerId
		pdrToStore.DefaultPdr.UrrId = defaultPdr.DefaultPdr.UrrId
		pdrToStore.TraceFlag = defaultPdr.TraceFlag
		pdrToStore.SdfMode = 2

	} else {
		if sdfPdr.TraceFlag {
			pdrToStore.TraceFlag = 1
		} else {
			pdrToStore.TraceFlag = 0
		}
		pdrToStore.SdfMode = 1
	}

	for sdfIdx := 0; sdfIdx < len(sdfPdr.SdfFilter) && sdfIdx < len(pdrToStore.DedicatedPdrs); sdfIdx++ {

		sdfFilterSource := &sdfPdr.SdfFilter[sdfIdx]
		sdfFilterTarget := &pdrToStore.DedicatedPdrs[sdfIdx]

		sdfFilterTarget.Pdr.OuterHeaderRemoval = sdfPdr.OuterHeaderRemoval
		sdfFilterTarget.Pdr.FarId = sdfPdr.FarId
		sdfFilterTarget.Pdr.QerId = sdfPdr.QerId
		sdfFilterTarget.Pdr.UrrId[0] = sdfPdr.Urr1Id
		sdfFilterTarget.Pdr.UrrId[1] = sdfPdr.Urr2Id

		sdfFilterTarget.SdfFilter.Protocol = sdfFilterSource.Protocol
		sdfFilterTarget.SdfFilter.SrcAddr.Type = sdfFilterSource.SrcAddress.Type
		sdfFilterTarget.SdfFilter.SrcAddr.Ip = Copy16Ip(sdfFilterSource.SrcAddress.Ip)
		sdfFilterTarget.SdfFilter.SrcAddr.Mask = Copy16Ip(sdfFilterSource.SrcAddress.Mask)
		sdfFilterTarget.SdfFilter.SrcPort.LowerBound = sdfFilterSource.SrcPortRange.LowerBound
		sdfFilterTarget.SdfFilter.SrcPort.UpperBound = sdfFilterSource.SrcPortRange.UpperBound
		sdfFilterTarget.SdfFilter.DstAddr.Type = sdfFilterSource.DstAddress.Type
		sdfFilterTarget.SdfFilter.DstAddr.Ip = Copy16Ip(sdfFilterSource.DstAddress.Ip)
		sdfFilterTarget.SdfFilter.DstAddr.Mask = Copy16Ip(sdfFilterSource.DstAddress.Mask)
		sdfFilterTarget.SdfFilter.DstPort.LowerBound = sdfFilterSource.DstPortRange.LowerBound
		sdfFilterTarget.SdfFilter.DstPort.UpperBound = sdfFilterSource.DstPortRange.UpperBound

		if sdfPdr.NotifyFlag {
			sdfFilterTarget.Notify = 1
		} else {
			sdfFilterTarget.Notify = 0
		}
	}
	return pdrToStore
}

func ToIpEntrypointPdrInfo(defaultPdr PdrInfo) IpEntrypointPdrInfo {
	var pdrToStore IpEntrypointPdrInfo
	pdrToStore.DefaultPdr.OuterHeaderRemoval = defaultPdr.OuterHeaderRemoval
	pdrToStore.DefaultPdr.FarId = defaultPdr.FarId
	pdrToStore.DefaultPdr.QerId = defaultPdr.QerId
	pdrToStore.DefaultPdr.UrrId[0] = defaultPdr.Urr1Id
	pdrToStore.DefaultPdr.UrrId[1] = defaultPdr.Urr2Id
	if defaultPdr.TraceFlag {
		pdrToStore.TraceFlag = 1
	} else {
		pdrToStore.TraceFlag = 0
	}
	return pdrToStore
}

//func FromIpEntrypointPdrInfo(storedPdr IpEntrypointPdrInfo) PdrInfo {
//	var pdrInfo PdrInfo
//	pdrInfo.OuterHeaderRemoval = storedPdr.OuterHeaderRemoval
//	pdrInfo.FarId = storedPdr.FarId
//	pdrInfo.QerId = storedPdr.QerId
//	return pdrInfo
//}

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

func (sdfFilter *SdfFilter) String() string {
	return fmt.Sprintf("%+v", *sdfFilter)
}
