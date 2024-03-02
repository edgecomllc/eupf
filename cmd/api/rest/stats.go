package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type XdpStats struct {
	Aborted  uint64 `json:"aborted"`
	Drop     uint64 `json:"drop"`
	Pass     uint64 `json:"pass"`
	Tx       uint64 `json:"tx"`
	Redirect uint64 `json:"redirect"`
}

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

// DisplayXdpStatistics godoc
//	@Summary		Display XDP statistics
//	@Description	Display XDP statistics
//	@Tags			XDP
//	@Produce		json
//	@Success		200	{object}	XdpStats
//	@Router			/xdp_stats [get]
func (h *ApiHandler) displayXdpStatistics(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, XdpStats{
		Aborted:  h.ForwardPlaneStats.GetAborted(),
		Drop:     h.ForwardPlaneStats.GetDrop(),
		Pass:     h.ForwardPlaneStats.GetPass(),
		Tx:       h.ForwardPlaneStats.GetTx(),
		Redirect: h.ForwardPlaneStats.GetRedirect(),
	})
}

// DisplayPacketStats godoc
//	@Summary		Display packet statistics
//	@Description	Display packet statistics
//	@Tags			Packet
//	@Produce		json
//	@Success		200	{object}	PacketStats
//	@Router			/packet_stats [get]
func (h *ApiHandler) displayPacketStats(c *gin.Context) {
	packets := h.ForwardPlaneStats.GetUpfExtStatField()
	c.IndentedJSON(http.StatusOK, PacketStats{
		RxArp:      packets.RxArp,
		RxIcmp:     packets.RxIcmp,
		RxIcmp6:    packets.RxIcmp6,
		RxIp4:      packets.RxIp4,
		RxIp6:      packets.RxIp6,
		RxTcp:      packets.RxTcp,
		RxUdp:      packets.RxUdp,
		RxOther:    packets.RxOther,
		RxGtpEcho:  packets.RxGtpEcho,
		RxGtpPdu:   packets.RxGtpPdu,
		RxGtpOther: packets.RxGtpOther,
		RxGtpUnexp: packets.RxGtpUnexp,
	})
}
