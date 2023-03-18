package main

import (
	"net"
	"unsafe"

	"github.com/cilium/ebpf"
)

// The BPF_ARRAY map type has no delete operation. The only way to delete an element is to replace it with a new one.

type PdrInfo struct {
	OuterHeaderRemoval uint8
	FarId              uint16
}

func (o *BpfObjects) PutPdrUpLink(teid uint32, pdrInfo PdrInfo) error {
	return o.ip_entrypointMaps.PdrMapUplinkIp4.Put(teid, unsafe.Pointer(&pdrInfo))
}

func (o *BpfObjects) PutPdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error {
	return o.ip_entrypointMaps.PdrMapDownlinkIp4.Put(ipv4, unsafe.Pointer(&pdrInfo))
}

func (o *BpfObjects) UpdatePdrUpLink(teid uint32, pdrInfo PdrInfo) error {
	return o.ip_entrypointMaps.PdrMapUplinkIp4.Update(teid, unsafe.Pointer(&pdrInfo), ebpf.UpdateExist)
}

func (o *BpfObjects) UpdatePdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error {
	return o.ip_entrypointMaps.PdrMapDownlinkIp4.Update(ipv4, unsafe.Pointer(&pdrInfo), ebpf.UpdateExist)
}

func (o *BpfObjects) DeletePdrUpLink(teid uint32) error {
	return o.ip_entrypointMaps.PdrMapUplinkIp4.Update(teid, unsafe.Pointer(&PdrInfo{}), ebpf.UpdateExist)
	//return o.ip_entrypointMaps.PdrMapUplinkIp4.Delete(teid)
}

func (o *BpfObjects) DeletePdrDownLink(ipv4 net.IP) error {
	return o.ip_entrypointMaps.PdrMapDownlinkIp4.Update(ipv4, unsafe.Pointer(&PdrInfo{}), ebpf.UpdateExist)
	//return o.ip_entrypointMaps.PdrMapDownlinkIp4.Delete(ipv4)
}

type FarInfo struct {
	Action              uint8
	OuterHeaderCreation uint8
	Teid                uint32
	Srcip               uint32
}

func (o *BpfObjects) PutFar(i uint32, farInfo FarInfo) error {
	return o.ip_entrypointMaps.FarMap.Put(i, unsafe.Pointer(&farInfo))
}

func (o *BpfObjects) UpdateFar(i uint32, farInfo FarInfo) error {
	return o.ip_entrypointMaps.FarMap.Update(i, unsafe.Pointer(&farInfo), ebpf.UpdateExist)
}

func (o *BpfObjects) DeleteFar(i uint32) error {
	return o.ip_entrypointMaps.FarMap.Update(i, unsafe.Pointer(&FarInfo{}), ebpf.UpdateExist)
	//return o.ip_entrypointMaps.FarMap.Delete(i)
}

type BpfMapOperations interface {
	PutPdrUpLink(teid uint32, pdrInfo PdrInfo) error
	PutPdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error
	UpdatePdrUpLink(teid uint32, pdrInfo PdrInfo) error
	UpdatePdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error
	DeletePdrUpLink(teid uint32) error
	DeletePdrDownLink(ipv4 net.IP) error
	PutFar(i uint32, farInfo FarInfo) error
	UpdateFar(i uint32, farInfo FarInfo) error
	DeleteFar(i uint32) error
}
