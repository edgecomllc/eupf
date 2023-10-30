package service

import (
	"net"
	"sync"
)

type IPAM struct {
	freeIPs []net.IP
	busyIPs map[uint64]net.IP
	sync.Mutex
}

func NewIPAM(ipRange string) (*IPAM, error) {
	_, ipNet, err := net.ParseCIDR(ipRange)
	if err != nil {
		return nil, err
	}

	freeIPs := make([]net.IP, 0, 255)
	busyIPs := make(map[uint64]net.IP)

	ip := ipNet.IP
	for ipNet.Contains(ip) {
		freeIPs = append(freeIPs, net.IP(ip))
		ip = nextIP(ip)
	}

	return &IPAM{
		freeIPs: freeIPs,
		busyIPs: busyIPs,
	}, nil
}

func (ipam *IPAM) AllocateIP(key uint64) net.IP {
	ipam.Lock()
	defer ipam.Unlock()

	if len(ipam.freeIPs) > 0 {
		ip := ipam.freeIPs[0]
		ipam.freeIPs = ipam.freeIPs[1:]
		ipam.busyIPs[key] = ip
		return ip
	}
	return nil
}

func (ipam *IPAM) ReleaseIP(key uint64) {
	ipam.Lock()
	defer ipam.Unlock()

	if ip, ok := ipam.busyIPs[key]; ok {
		ipam.freeIPs = append(ipam.freeIPs, ip)
		delete(ipam.busyIPs, key)
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
