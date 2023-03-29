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

func (stat *UpfXdpActionStatistic) getUpfExtStatField(field uint32) uint32 {
	var result uint32
	err := stat.bpfObjects.ip_entrypointMaps.UpfExtStat.Lookup(field, unsafe.Pointer(&result))
	if err != nil {
		log.Println(err)
	}
	return result
}

func (stat *UpfXdpActionStatistic) GetRxTotal() uint32 {
	return stat.getUpfExtStatField(uint32(0))
}

func (stat *UpfXdpActionStatistic) GetRxArp() uint32 {
	return stat.getUpfExtStatField(uint32(1))
}

func (stat *UpfXdpActionStatistic) GetRxIcmp() uint32 {
	return stat.getUpfExtStatField(uint32(2))
}

func (stat *UpfXdpActionStatistic) GetRxIcmp6() uint32 {
	return stat.getUpfExtStatField(uint32(3))
}

func (stat *UpfXdpActionStatistic) GetRxIp4() uint32 {
	return stat.getUpfExtStatField(uint32(4))
}

func (stat *UpfXdpActionStatistic) GetRxIp6() uint32 {
	return stat.getUpfExtStatField(uint32(5))
}

func (stat *UpfXdpActionStatistic) GetRxTcp() uint32 {
	return stat.getUpfExtStatField(uint32(6))
}

func (stat *UpfXdpActionStatistic) GetRxUdp() uint32 {
	return stat.getUpfExtStatField(uint32(7))
}

func (stat *UpfXdpActionStatistic) GetRxOther() uint32 {
	return stat.getUpfExtStatField(uint32(8))
}

func (stat *UpfXdpActionStatistic) GetRxGtpEcho() uint32 {
	return stat.getUpfExtStatField(uint32(9))
}

func (stat *UpfXdpActionStatistic) GetRxGtpPdu() uint32 {
	return stat.getUpfExtStatField(uint32(10))
}

func (stat *UpfXdpActionStatistic) GetRxGtpOther() uint32 {
	return stat.getUpfExtStatField(uint32(11))
}

func (stat *UpfXdpActionStatistic) GetRxGtpUnexp() uint32 {
	return stat.getUpfExtStatField(uint32(12))
}

func (stat *UpfXdpActionStatistic) GetRxGtpUnsup() uint32 {
	return stat.getUpfExtStatField(uint32(13))
}
