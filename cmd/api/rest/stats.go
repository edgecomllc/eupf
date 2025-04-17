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

type RouteStats struct {
	FibLookupIp4Cache     uint64 `json:"fib_lookup_ip4_cache"`
	FibLookupIp4Ok        uint64 `json:"fib_lookup_ip4_ok"`
	FibLookupIp4ErrorDrop uint64 `json:"fib_lookup_ip4_error_drop"`
	FibLookupIp4ErrorPass uint64 `json:"fib_lookup_ip4_error_pass"`

	FibLookupIp6Cache     uint64 `json:"fib_lookup_ip6_cache"`
	FibLookupIp6Ok        uint64 `json:"fib_lookup_ip6_ok"`
	FibLookupIp6ErrorDrop uint64 `json:"fib_lookup_ip6_error_drop"`
	FibLookupIp6ErrorPass uint64 `json:"fib_lookup_ip6_error_pass"`
}

// DisplayXdpStatistics godoc
//
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
//
//	@Summary		Display packet statistics
//	@Description	Display packet statistics
//	@Tags			Packet
//	@Produce		json
//	@Success		200	{object}	PacketStats
//	@Router			/packet_stats [get]
func (h *ApiHandler) displayPacketStats(c *gin.Context) {
	packets := h.ForwardPlaneStats.GetUpfExtStat()
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

// DisplayRouteStats godoc
//
//	@Summary		Display route statistics
//	@Description	Display route statistics
//	@Tags			Route
//	@Produce		json
//	@Success		200	{object}	RouteStats
//	@Router			/route_stats [get]
func (h *ApiHandler) displayRouteStats(c *gin.Context) {
	route := h.ForwardPlaneStats.GetUpfRouteStat()
	c.IndentedJSON(http.StatusOK, RouteStats{
		FibLookupIp4Cache:     route.FibLookupIp4Cache,
		FibLookupIp4Ok:        route.FibLookupIp4Ok,
		FibLookupIp4ErrorDrop: route.FibLookupIp4ErrorDrop,
		FibLookupIp4ErrorPass: route.FibLookupIp4ErrorPass,

		FibLookupIp6Cache:     route.FibLookupIp6Cache,
		FibLookupIp6Ok:        route.FibLookupIp6Ok,
		FibLookupIp6ErrorDrop: route.FibLookupIp6ErrorDrop,
		FibLookupIp6ErrorPass: route.FibLookupIp6ErrorPass,
	})
}
