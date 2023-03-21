package main

import (
	"log"
	"unsafe"
)

type UpfXdpActionStatistic struct {
	bpfObjects *BpfObjects
}

func (stat *UpfXdpActionStatistic) GetAborted() uint32 {
	var result uint32
	err := stat.bpfObjects.ip_entrypointMaps.UpfXdpStatistic.Lookup(uint32(0), unsafe.Pointer(&result))
	if err != nil {
		log.Println(err)
	}
	return result
}
func (stat *UpfXdpActionStatistic) GetDrop() uint32 {
	var result uint32
	err := stat.bpfObjects.ip_entrypointMaps.UpfXdpStatistic.Lookup(uint32(1), unsafe.Pointer(&result))
	if err != nil {
		log.Println(err)
	}
	return result
}
func (stat *UpfXdpActionStatistic) GetPass() uint32 {
	var result uint32
	err := stat.bpfObjects.ip_entrypointMaps.UpfXdpStatistic.Lookup(uint32(2), unsafe.Pointer(&result))
	if err != nil {
		log.Println(err)
	}
	return result
}
func (stat *UpfXdpActionStatistic) GetTx() uint32 {
	var result uint32
	err := stat.bpfObjects.ip_entrypointMaps.UpfXdpStatistic.Lookup(uint32(3), unsafe.Pointer(&result))
	if err != nil {
		log.Println(err)
	}
	return result
}

func (stat *UpfXdpActionStatistic) GetRedirect() uint32 {
	var result uint32
	err := stat.bpfObjects.ip_entrypointMaps.UpfXdpStatistic.Lookup(uint32(4), unsafe.Pointer(&result))
	if err != nil {
		log.Println(err)
	}
	return result
}
