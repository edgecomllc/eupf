package service

import (
	"errors"
	"net"
	"strconv"
	"sync"
)

type IPAM struct {
	freeIPs   []net.IP
	busyIPs   map[uint64]net.IP
	freeTEIDs []uint32
	busyTEIDs map[string]uint32
	sync.RWMutex
}

func NewIPAM(ipRange string) (*IPAM, error) {
	_, ipNet, err := net.ParseCIDR(ipRange)
	if err != nil {
		return nil, err
	}

	freeIPs := make([]net.IP, 0, 255)
	busyIPs := make(map[uint64]net.IP)

	freeTEIDs := make([]uint32, 0, 255)
	busyTEIDs := make(map[string]uint32)

	var teid uint32
	ip := ipNet.IP

	for ipNet.Contains(ip) {
		freeIPs = append(freeIPs, net.IP(ip))
		ip = nextIP(ip)
		teid++
		freeTEIDs = append(freeTEIDs, teid)

	}

	return &IPAM{
		freeIPs:   freeIPs,
		busyIPs:   busyIPs,
		freeTEIDs: freeTEIDs,
		busyTEIDs: busyTEIDs,
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

func (ipam *IPAM) AllocateTEID(seID uint64, pdrID uint16, chooseID uint8) (uint32, error) {
	ipam.Lock()
	defer ipam.Unlock()

	if len(ipam.freeTEIDs) > 0 {
		key := strconv.Itoa(int(seID)) + ":" + strconv.Itoa(int(pdrID))
		if chooseID != 0 {
			key += ":" + strconv.Itoa(int(chooseID))
		}
		teid := ipam.freeTEIDs[0]
		ipam.freeTEIDs = ipam.freeTEIDs[1:]
		ipam.busyTEIDs[key] = teid
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

func (ipam *IPAM) ReleaseTEID(seID uint64, pdrID uint16, chooseID uint8) {
	ipam.Lock()
	defer ipam.Unlock()

	key := strconv.Itoa(int(seID)) + ":" + strconv.Itoa(int(pdrID))
	if chooseID != 0 {
		key += ":" + strconv.Itoa(int(chooseID))
	}
	if teid, ok := ipam.busyTEIDs[key]; ok {
		ipam.freeTEIDs = append(ipam.freeTEIDs, teid)
		delete(ipam.busyTEIDs, key)
	}
}

func (ipam *IPAM) GetTEID(seID uint64, pdrID uint16, chooseID uint8) (uint32, bool) {
	ipam.RLock()
	defer ipam.RUnlock()

	key := strconv.Itoa(int(seID)) + ":" + strconv.Itoa(int(pdrID))
	if chooseID != 0 {
		key += ":" + strconv.Itoa(int(chooseID))
	}
	teid, ok := ipam.busyTEIDs[key]
	return teid, ok
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
