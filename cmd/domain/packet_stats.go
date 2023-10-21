package domain

type PacketStats struct {
	RxArp      uint64 `json:"rx_arp"`
	RxIcmp     uint64 `json:"rx_icmp"`
	RxIcmp6    uint64 `json:"rx_icmp6"`
	RxIp4      uint64 `json:"rx_ip4"`
	RxIp6      uint64 `json:"rx_ip6"`
	RxTcp      uint64 `json:"rx_tcp"`
	RxUdp      uint64 `json:"rx_udp"`
	RxOther    uint64 `json:"rx_other"`
	RxGtpEcho  uint64 `json:"rx_gtp_echo"`
	RxGtpPdu   uint64 `json:"rx_gtp_pdu"`
	RxGtpOther uint64 `json:"rx_gtp_other"`
	RxGtpUnexp uint64 `json:"rx_gtp_unexp"`
}
