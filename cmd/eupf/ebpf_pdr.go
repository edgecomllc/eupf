package main

import (
	"net"
	// // #include "xdp/ip_entrypoint.c"
	// "C"

	"unsafe"

	"github.com/cilium/ebpf"
)

type PdrInfo struct {
	OuterHeaderRemoval uint8
	FarId              uint16
}

// func (pdrInfo *PdrInfo) ConvertToC() C.struct_pdr_info {
// 	pdrInfoC := C.struct_pdr_info{
// 		outer_header_removal: C.uint8_t(pdrInfo.OuterHeaderRemoval),
// 		far_id:               C.uint16_t(pdrInfo.FarId),
// 	}
// 	return pdrInfoC
// }

func (o *BpfObjects) PutPdrUpLink(teid uint32, pdrInfo PdrInfo) error {
	return o.ip_entrypointMaps.PdrMapUplinkIp4.Put(teid, unsafe.Pointer(&pdrInfo))
}

func (o *BpfObjects) PutPdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error {
	return o.ip_entrypointMaps.PdrMapDownlinkIp4.Put(ipv4, pdrInfo)
}

func (o *BpfObjects) UpdatePdrUpLink(teid uint32, pdrInfo PdrInfo) error {
	return o.ip_entrypointMaps.PdrMapUplinkIp4.Update(teid, pdrInfo, ebpf.UpdateExist)
}

func (o *BpfObjects) UpdatePdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error {
	return o.ip_entrypointMaps.PdrMapDownlinkIp4.Update(ipv4, pdrInfo, ebpf.UpdateExist)
}

func (o *BpfObjects) DeletePdrUpLink(teid uint32) error {
	return o.ip_entrypointMaps.PdrMapUplinkIp4.Delete(teid)
}

func (o *BpfObjects) DeletePdrDownLink(ipv4 net.IP) error {
	return o.ip_entrypointMaps.PdrMapDownlinkIp4.Delete(ipv4)
}

type FarInfo struct {
	Action              uint8
	OuterHeaderCreation uint8
	Teid                uint32
	Srcip               uint32
}

// func (farInfo *FarInfo) ConvertToC() C.struct_far_info {
// 	farInfoC := C.struct_far_info{
// 		action:                C.uint8_t(farInfo.Action),
// 		outer_header_creation: C.uint8_t(farInfo.OuterHeaderCreation),
// 		teid:                  C.uint32_t(farInfo.Teid),
// 		srcip:                 C.uint32_t(farInfo.Srcip),
// 	}
// 	return farInfoC
// }

func (o *BpfObjects) PutFar(i uint32, farInfo FarInfo) error {
	return o.ip_entrypointMaps.FarMap.Put(i, unsafe.Pointer(&farInfo))
}

func (o *BpfObjects) UpdateFar(i uint32, farInfo FarInfo) error {
	return o.ip_entrypointMaps.FarMap.Update(i, farInfo, ebpf.UpdateExist)
}

func (o *BpfObjects) DeleteFar(i uint32) error {
	return o.ip_entrypointMaps.FarMap.Delete(i)
}
