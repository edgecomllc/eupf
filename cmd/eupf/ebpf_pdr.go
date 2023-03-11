package main

import (
	"net"

	"github.com/cilium/ebpf"
)

type PdrInfo struct {
	OuterHeaderRemoval uint8
	FarId              uint32
}

func (o *BpfObjects) PutPdrUpLink(teid uint32, pdrInfo PdrInfo) error {
	return o.ip_entrypointMaps.PdrMapUplinkIp4.Put(teid, pdrInfo)
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

type FarInfo struct {
	Action              uint8
	OuterHeaderCreation uint8
	Teid                uint32
	Srcip               uint32
}

func (o *BpfObjects) PutFar(i uint32, farInfo FarInfo) error {
	return o.ip_entrypointMaps.FarMap.Put(i, farInfo)
}

func (o *BpfObjects) UpdateFar(i uint32, farInfo FarInfo) error {
	return o.ip_entrypointMaps.FarMap.Update(i, farInfo, ebpf.UpdateExist)
}
