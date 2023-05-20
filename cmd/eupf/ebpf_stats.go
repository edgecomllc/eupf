package main

import (
	"log"
	"unsafe"
)

type UpfXdpActionStatistic struct {
	bpfObjects *BpfObjects
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

type UpfStatistic struct {
	Counters UpfCounters
	XdpStats [5]uint32
}

// Getters for the upf_xdp_statistic (xdp_action)

func (stat *UpfXdpActionStatistic) getUpfXdpStatisticField(field uint32) uint32 {

	var statistic UpfStatistic
	err := stat.bpfObjects.ip_entrypointMaps.UpfExtStat.Lookup(uint32(0), unsafe.Pointer(&statistic))
	if err != nil {
		log.Println(err)
		return 0
	}

	return statistic.XdpStats[field]
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

	var statistic UpfStatistic
	var counter UpfCounters
	err := stat.bpfObjects.ip_entrypointMaps.UpfExtStat.Lookup(uint32(0), unsafe.Pointer(&statistic))
	if err != nil {
		log.Println(err)
		return counter
	}

	return statistic.Counters
}
