package service

import (
	"errors"
	"net"
	"sync"
)

type ResourceManager struct {
	IPAM   *IPAM
	FTEIDM *FTEIDM
	UEIP   bool
	FTUP   bool
}

type FTEIDM struct {
	freeTEIDs []uint32
	busyTEIDs map[uint64]map[uint16]uint32 // map[seID]map[pdrID]teid
	sync.RWMutex
}

type IPAM struct {
	freeIPs []net.IP
	busyIPs map[uint64]net.IP
	sync.RWMutex
}

func NewResourceManager(ueip, ftup bool, ipRange string, teidRange uint32) (*ResourceManager, error) {

	var ipam IPAM
	var fteidm FTEIDM

	if ueip {
		_, ipNet, err := net.ParseCIDR(ipRange)
		if err != nil {
			return nil, err
		}

		freeIPs := make([]net.IP, 0, 10000)
		busyIPs := make(map[uint64]net.IP)

		ip := ipNet.IP

		for ipNet.Contains(ip) {
			freeIPs = append(freeIPs, net.IP(ip))
			ip = nextIP(ip)
		}

		ipam = IPAM{
			freeIPs: freeIPs,
			busyIPs: busyIPs,
		}
	}

	if ftup {
		freeTEIDs := make([]uint32, 0, 10000)
		busyTEIDs := make(map[uint64]map[uint16]uint32)

		var teid uint32

		for teid = 1; teid <= teidRange; teid++ {
			freeTEIDs = append(freeTEIDs, teid)
		}

		fteidm = FTEIDM{
			freeTEIDs: freeTEIDs,
			busyTEIDs: busyTEIDs,
		}
	}

	return &ResourceManager{
		IPAM:   &ipam,
		FTEIDM: &fteidm,
		UEIP:   ueip,
		FTUP:   ftup,
	}, nil
}

func (ipam *IPAM) AllocateIP(key uint64) (net.IP, error) {
	ipam.Lock()
	defer ipam.Unlock()

	if len(ipam.freeIPs) > 0 {
		ip := ipam.freeIPs[0]
		ipam.freeIPs = ipam.freeIPs[1:]
		ipam.busyIPs[key] = ip
		return ip, nil
	} else {
		return nil, errors.New("no free ip available")
	}
}

func (ipam *FTEIDM) AllocateTEID(seID uint64, pdrID uint16) (uint32, error) {
	ipam.Lock()
	defer ipam.Unlock()

	if len(ipam.freeTEIDs) > 0 {
		teid := ipam.freeTEIDs[0]
		ipam.freeTEIDs = ipam.freeTEIDs[1:]
		if _, ok := ipam.busyTEIDs[seID]; !ok {
			pdr := make(map[uint16]uint32)
			pdr[pdrID] = teid
			ipam.busyTEIDs[seID] = pdr
		} else {
			ipam.busyTEIDs[seID][pdrID] = teid
		}
		return teid, nil
	} else {
		return 0, errors.New("no free TEID available")
	}
}

func (rm *ResourceManager) ReleaseResources(seID uint64) {

	rm.IPAM.Lock()
	if ip, ok := rm.IPAM.busyIPs[seID]; ok {
		rm.IPAM.freeIPs = append(rm.IPAM.freeIPs, ip)
		delete(rm.IPAM.busyIPs, seID)
	}
	rm.IPAM.Unlock()

	rm.FTEIDM.Lock()
	if teid, ok := rm.FTEIDM.busyTEIDs[seID]; ok {
		for _, t := range teid {
			rm.FTEIDM.freeTEIDs = append(rm.FTEIDM.freeTEIDs, t)
		}
		delete(rm.FTEIDM.busyTEIDs, seID)
	}
	rm.FTEIDM.Unlock()
}

func nextIP(ip net.IP) net.IP {
	nextIP := make(net.IP, len(ip))
	copy(nextIP, ip)
	for i := len(nextIP) - 1; i >= 0; i-- {
		nextIP[i]++
		if nextIP[i] > 0 {
			break
		}
	}
	return nextIP
}
