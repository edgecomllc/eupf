package rest

import (
	"github.com/edgecomllc/eupf/components/ebpf"
	"github.com/edgecomllc/eupf/pkg/domain"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
)

// ListUpfPipeline godoc
// @Summary List UPF pipeline
// @Description List UPF pipeline
// @Tags UPF
// @Produce  json
// @Success 200 {object} []ebpf.BpfMapProgArrayMember
// @Router /upf_pipeline [get]
func (h *Handler) listUpfPipeline(c *gin.Context) {
	if elements, err := ebpf.ListMapProgArrayContents(h.BpfObjects.UpfXdpObjects.UpfPipeline); err != nil {
		log.Info().Msgf("Error reading map: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.IndentedJSON(http.StatusOK, elements)
	}
}

// DisplayXdpStatistics godoc
// @Summary Display XDP statistics
// @Description Display XDP statistics
// @Tags XDP
// @Produce  json
// @Success 200 {object} XdpStats
// @Router /xdp_stats [get]
func (h *Handler) displayXdpStatistics(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, domain.XdpStats{
		Aborted:  h.ForwardPlaneStats.GetAborted(),
		Drop:     h.ForwardPlaneStats.GetDrop(),
		Pass:     h.ForwardPlaneStats.GetPass(),
		Tx:       h.ForwardPlaneStats.GetTx(),
		Redirect: h.ForwardPlaneStats.GetRedirect(),
	})
}

// DisplayPacketStats godoc
// @Summary Display packet statistics
// @Description Display packet statistics
// @Tags Packet
// @Produce  json
// @Success 200 {object} PacketStats
// @Router /packet_stats [get]
func (h *Handler) displayPacketStats(c *gin.Context) {
	packets := h.ForwardPlaneStats.GetUpfExtStatField()
	c.IndentedJSON(http.StatusOK, domain.PacketStats{
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
