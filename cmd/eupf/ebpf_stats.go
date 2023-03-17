package main

import "unsafe"

//	enum xdp_action {
//		XDP_ABORTED = 0,
//		XDP_DROP,
//		XDP_PASS,
//		XDP_TX,
//		XDP_REDIRECT,
//	};

type GetStat func() (uint32, error)

// Create "EbpfGetStat" callback
func CreateEbpfGetStats(bpfObjects *BpfObjects) (GetStat, GetStat, GetStat, GetStat, GetStat) {
	Aborted := func() (uint32, error) {
		var result uint32
		err := bpfObjects.ip_entrypointMaps.UpfXdpStatistic.Lookup(0, unsafe.Pointer(&result))
		return result, err
	}
	Drop := func() (uint32, error) {
		var result uint32
		err := bpfObjects.ip_entrypointMaps.UpfXdpStatistic.Lookup(1, unsafe.Pointer(&result))
		return result, err
	}
	Pass := func() (uint32, error) {
		var result uint32
		err := bpfObjects.ip_entrypointMaps.UpfXdpStatistic.Lookup(2, unsafe.Pointer(&result))
		return result, err
	}
	Tx := func() (uint32, error) {
		var result uint32
		err := bpfObjects.ip_entrypointMaps.UpfXdpStatistic.Lookup(3, unsafe.Pointer(&result))
		return result, err
	}
	Redirect := func() (uint32, error) {
		var result uint32
		err := bpfObjects.ip_entrypointMaps.UpfXdpStatistic.Lookup(4, unsafe.Pointer(&result))
		return result, err
	}
	return Aborted, Drop, Pass, Tx, Redirect
}
