package service

import (
	"errors"
	"net"
	"sync"
)

type ResourceManager struct {
	IPAM   *IPAM
	FTEIDM *FTEIDM
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

func NewResourceManager(ueip, ftup bool, ipRange string) (*ResourceManager, error) {

	var ipam IPAM
	var fteidm FTEIDM

	if ueip {
		_, ipNet, err := net.ParseCIDR(ipRange)
		if err != nil {
			return nil, err
		}

		freeIPs := make([]net.IP, 0, 65536)
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
		freeTEIDs := make([]uint32, 0, 65536)
		busyTEIDs := make(map[uint64]map[uint16]uint32)

		var teid uint32

		for teid = 1; teid <= 65536; teid++ {
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

func (ipam *IPAM) ReleaseIP(key uint64) {
	ipam.Lock()
	defer ipam.Unlock()

	if ip, ok := ipam.busyIPs[key]; ok {
		ipam.freeIPs = append(ipam.freeIPs, ip)
		delete(ipam.busyIPs, key)
	}
}

func (ipam *FTEIDM) ReleaseTEID(seID uint64) {
	ipam.Lock()
	defer ipam.Unlock()

	if teid, ok := ipam.busyTEIDs[seID]; ok {
		for _, t := range teid {
			ipam.freeTEIDs = append(ipam.freeTEIDs, t)
		}
		delete(ipam.busyTEIDs, seID)
	}
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
