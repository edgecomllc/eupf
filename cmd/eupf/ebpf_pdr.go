package main

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
}

func (bpfObjects *BpfObjects) PutPdrUpLink(teid uint32, pdrInfo PdrInfo) error {
	log.Printf("EBPF: Put PDR Uplink: teid=%d, pdrInfo=%+v", teid, pdrInfo)
	return bpfObjects.ip_entrypointMaps.PdrMapUplinkIp4.Put(teid, unsafe.Pointer(&pdrInfo))
}

func (bpfObjects *BpfObjects) PutPdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error {
	log.Printf("EBPF: Put PDR Downlink: ipv4=%s, pdrInfo=%+v", ipv4, pdrInfo)
	return bpfObjects.ip_entrypointMaps.PdrMapDownlinkIp4.Put(ipv4, unsafe.Pointer(&pdrInfo))
}

func (bpfObjects *BpfObjects) UpdatePdrUpLink(teid uint32, pdrInfo PdrInfo) error {
	log.Printf("EBPF: Update PDR Uplink: teid=%d, pdrInfo=%+v", teid, pdrInfo)
	return bpfObjects.ip_entrypointMaps.PdrMapUplinkIp4.Update(teid, unsafe.Pointer(&pdrInfo), ebpf.UpdateExist)
}

func (bpfObjects *BpfObjects) UpdatePdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error {
	log.Printf("EBPF: Update PDR Downlink: ipv4=%s, pdrInfo=%+v", ipv4, pdrInfo)
	return bpfObjects.ip_entrypointMaps.PdrMapDownlinkIp4.Update(ipv4, unsafe.Pointer(&pdrInfo), ebpf.UpdateExist)
}

func (bpfObjects *BpfObjects) DeletePdrUpLink(teid uint32) error {
	log.Printf("EBPF: Delete PDR Uplink: teid=%d", teid)
	return bpfObjects.ip_entrypointMaps.PdrMapUplinkIp4.Update(teid, unsafe.Pointer(&PdrInfo{}), ebpf.UpdateExist)
	//return o.ip_entrypointMaps.PdrMapUplinkIp4.Delete(teid)
}

func (bpfObjects *BpfObjects) DeletePdrDownLink(ipv4 net.IP) error {
	log.Printf("EBPF: Delete PDR Downlink: ipv4=%s", ipv4)
	return bpfObjects.ip_entrypointMaps.PdrMapDownlinkIp4.Update(ipv4, unsafe.Pointer(&PdrInfo{}), ebpf.UpdateExist)
	//return o.ip_entrypointMaps.PdrMapDownlinkIp4.Delete(ipv4)
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
		"action":                f.Action,
		"outer_header_creation": f.OuterHeaderCreation,
		"teid":                  f.Teid,
		"remote_ip":             remoteIP.String(),
		"local_ip":              localIP.String(),
	}
	return json.Marshal(data)
}

func (bpfObjects *BpfObjects) PutFar(i uint32, farInfo FarInfo) error {
	log.Printf("EBPF: Put FAR: i=%d, farInfo=%+v", i, farInfo)
	return bpfObjects.ip_entrypointMaps.FarMap.Put(i, unsafe.Pointer(&farInfo))
}

func (bpfObjects *BpfObjects) UpdateFar(i uint32, farInfo FarInfo) error {
	log.Printf("EBPF: Update FAR: i=%d, farInfo=%+v", i, farInfo)
	return bpfObjects.ip_entrypointMaps.FarMap.Update(i, unsafe.Pointer(&farInfo), ebpf.UpdateExist)
}

func (bpfObjects *BpfObjects) DeleteFar(i uint32) error {
	log.Printf("EBPF: Delete FAR: i=%d", i)
	return bpfObjects.ip_entrypointMaps.FarMap.Update(i, unsafe.Pointer(&FarInfo{}), ebpf.UpdateExist)
	//return o.ip_entrypointMaps.FarMap.Delete(i)
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

func (bpfObjects *BpfObjects) PutQer(i uint32, qerInfo QerInfo) error {
	log.Printf("EBPF: Put QER: i=%d, qerInfo=%+v", i, qerInfo)
	return bpfObjects.ip_entrypointMaps.QerMap.Put(i, unsafe.Pointer(&qerInfo))
}

func (bpfObjects *BpfObjects) UpdateQer(i uint32, qerInfo QerInfo) error {
	log.Printf("EBPF: Update QER: i=%d, qerInfo=%+v", i, qerInfo)
	return bpfObjects.ip_entrypointMaps.QerMap.Update(i, unsafe.Pointer(&qerInfo), ebpf.UpdateExist)
}

func (bpfObjects *BpfObjects) DeleteQer(i uint32) error {
	log.Printf("EBPF: Delete QER: i=%d", i)
	return bpfObjects.ip_entrypointMaps.QerMap.Update(i, unsafe.Pointer(&QerInfo{}), ebpf.UpdateExist)
}

type ForwardingPlaneController interface {
	PutPdrUpLink(teid uint32, pdrInfo PdrInfo) error
	PutPdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error
	UpdatePdrUpLink(teid uint32, pdrInfo PdrInfo) error
	UpdatePdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error
	DeletePdrUpLink(teid uint32) error
	DeletePdrDownLink(ipv4 net.IP) error
	PutFar(i uint32, farInfo FarInfo) error
	UpdateFar(i uint32, farInfo FarInfo) error
	DeleteFar(i uint32) error
	PutQer(i uint32, qerInfo QerInfo) error
	UpdateQer(i uint32, qerInfo QerInfo) error
	DeleteQer(i uint32) error
}
