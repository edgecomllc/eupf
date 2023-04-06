package main

import (
	"log"
	"unsafe"
)

type UpfXdpActionStatistic struct {
	bpfObjects *BpfObjects
}

// Getters for the upf_xdp_statistic (xdp_action)

func (stat *UpfXdpActionStatistic) getUpfXdpStatisticField(field uint32) uint32 {
	var result uint32
	err := stat.bpfObjects.ip_entrypointMaps.UpfXdpStatistic.Lookup(field, unsafe.Pointer(&result))
	if err != nil {
		log.Println(err)
	}
	return result
}

func (stat *UpfXdpActionStatistic) GetAborted() uint32 {
	return stat.getUpfXdpStatisticField(uint32(0))
}

func (stat *UpfXdpActionStatistic) GetDrop() uint32 {
	return stat.getUpfXdpStatisticField(uint32(1))
}

func (stat *UpfXdpActionStatistic) GetPass() uint32 {
	return stat.getUpfXdpStatisticField(uint32(2))
}

func (stat *UpfXdpActionStatistic) GetTx() uint32 {
	return stat.getUpfXdpStatisticField(uint32(3))
}

func (stat *UpfXdpActionStatistic) GetRedirect() uint32 {
	return stat.getUpfXdpStatisticField(uint32(4))
}

// Getters for the upf_ext_stat (upf_counters)
// #TODO: Do not retrieve the whole struct each time.
func (stat *UpfXdpActionStatistic) getUpfExtStatField() UpfCounters {
	var result UpfCounters
	err := stat.bpfObjects.ip_entrypointMaps.UpfExtStat.Lookup(uint32(0), unsafe.Pointer(&result))
	if err != nil {
		log.Println(err)
	}
	return result
}

type UpfCounters struct {
	RxArp      uint64
	RxIcmp     uint64
	RxIcmp6    uint64
	RxIp4      uint64
	RxIp6      uint64
	RxTcp      uint64
	RxUdp      uint64
	RxOther    uint64
	RxGtpEcho  uint64
	RxGtpPdu   uint64
	RxGtpOther uint64
	RxGtpUnexp uint64
}
