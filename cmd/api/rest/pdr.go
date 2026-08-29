package rest

import (
	"net"
	"net/http"
	"strconv"
	"unsafe"

	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type PdrUplinkElement struct {
	Id                 uint32 `json:"teid"`
	FarId              uint32 `json:"far_id"`
	QerId              uint32 `json:"qer_id"`
	OuterHeaderRemoval uint8  `json:"outer_header_removal"`
	Trace              bool   `json:"trace"`
}

type PdrDownlinkElement struct {
	Id                 string `json:"ip"`
	FarId              uint32 `json:"far_id"`
	QerId              uint32 `json:"qer_id"`
	OuterHeaderRemoval uint8  `json:"outer_header_removal"`
	Trace              bool   `json:"trace"`
}

// GetUplinkPdrValue godoc
//
//	@Summary Get uplink PDR map element
//	@Description Retrieve uplink PDR map element by ID
//	@Tags PDR
//	@Produce json
//	@Param id path int true "PDR ID"
//	@Success 200 {object} PdrUplinkElement
//	@Failure 400 {object} map[string]string
//	@Failure 404 {object} map[string]string
//	@Router /uplink_pdr_map/{id} [get]
func (h *ApiHandler) getUplinkPdrValue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		log.Warn().Err(err).Msgf("can't parse object id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't parse object id"})
		return
	}

	var value ebpf.PdrInfo
	if err = h.BpfObjects.IpEntrypointObjects.PdrMapTeidIp4.Lookup(uint32(id), unsafe.Pointer(&value)); err != nil {
		log.Warn().Err(err).Msgf("can't get pdr from map")
		c.JSON(http.StatusNotFound, gin.H{"error": "no pdr found"})
		return
	}

	c.IndentedJSON(http.StatusOK, PdrUplinkElement{
		Id:                 uint32(id),
		OuterHeaderRemoval: value.OuterHeaderRemoval,
		FarId:              value.FarId,
		QerId:              value.QerId,
		Trace:              value.TraceFlag,
	})
}

// SetUplinkPdrValue godoc
//
//	@Summary Set uplink PDR map element
//	@Description Create or update uplink PDR map element
//	@Tags PDR
//	@Accept json
//	@Produce json
//	@Param id path int true "PDR ID"
//	@Param pdr body PdrUplinkElement true "PDR element data"
//	@Success 201 {object} PdrUplinkElement
//	@Failure 400 {object} map[string]string
//	@Failure 500 {object} map[string]string
//	@Router /uplink_pdr_map/{id} [put]
func (h *ApiHandler) setUplinkPdrValue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		log.Warn().Err(err).Msgf("can't parse object id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't parse object id"})
		return
	}

	var pdrElement PdrUplinkElement
	pdrElement.Id = uint32(id)
	if err := c.BindJSON(&pdrElement); err != nil {
		log.Warn().Err(err).Msgf("can't parse request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't parse request body"})
		return
	}

	value := ebpf.PdrInfo{
		OuterHeaderRemoval: pdrElement.OuterHeaderRemoval,
		FarId:              pdrElement.FarId,
		QerId:              pdrElement.QerId,
		TraceFlag:          pdrElement.Trace,
	}

	if err := h.BpfObjects.PutPdrUplink(uint32(id), value); err != nil {
		log.Warn().Err(err).Msgf("can't set pdr")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "can't set pdr"})
		return
	}

	c.IndentedJSON(http.StatusCreated, pdrElement)
}

// GetDownlinkPdrValue godoc
//
//	@Summary Get downlink PDR map element
//	@Description Retrieve downlink PDR map element by IP address
//	@Tags PDR
//	@Produce json
//	@Param ip path string true "IP Address (IPv4 or IPv6)"
//	@Success 200 {object} PdrDownlinkElement
//	@Failure 400 {object} map[string]string
//	@Failure 404 {object} map[string]string
//	@Router /downlink_pdr_map/{id} [get]
func (h *ApiHandler) getDownlinkPdrValue(c *gin.Context) {
	ip := net.ParseIP(c.Param("id"))
	if ip == nil {
		log.Warn().Msgf("can't parse object id: not an IPv4/IPv6")
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't parse object id: not an IPv4/IPv6"})
		return
	}

	var value ebpf.PdrInfo
	switch len(ip) {
	case 4:
		if err := h.BpfObjects.IpEntrypointObjects.PdrMapDownlinkIp4.Lookup(ip, unsafe.Pointer(&value)); err != nil {
			log.Warn().Err(err).Msgf("can't get pdr from ipv4 map")
			c.JSON(http.StatusNotFound, gin.H{"error": "no pdr found"})
			return
		}
	case 16:
		if err := h.BpfObjects.IpEntrypointObjects.PdrMapDownlinkIp6.Lookup(ip, unsafe.Pointer(&value)); err != nil {
			log.Warn().Err(err).Msgf("can't get pdr from ipv6 map")
			c.JSON(http.StatusNotFound, gin.H{"error": "no pdr found"})
			return
		}
	default:
		log.Warn().Msgf("wrong ip address")
		c.JSON(http.StatusNotFound, gin.H{"error": "wrong ip address"})
		return
	}

	c.IndentedJSON(http.StatusOK, PdrDownlinkElement{
		Id:                 ip.String(),
		OuterHeaderRemoval: value.OuterHeaderRemoval,
		FarId:              value.FarId,
		QerId:              value.QerId,
		Trace:              value.TraceFlag,
	})
}

// SetDownlinkPdrValue godoc
//
//	@Summary Set downlink PDR map element
//	@Description Create or update downlink PDR map element by IP address
//	@Tags PDR
//	@Accept json
//	@Produce json
//	@Param ip path string true "IP Address (IPv4 or IPv6)"
//	@Param pdr body PdrDownlinkElement true "PDR element data"
//	@Success 201 {object} PdrDownlinkElement
//	@Failure 400 {object} map[string]string
//	@Failure 500 {object} map[string]string
//	@Router /downlink_pdr_map/{id} [put]
func (h *ApiHandler) setDownlinkPdrValue(c *gin.Context) {
	ip := net.ParseIP(c.Param("id"))
	if ip == nil {
		log.Warn().Msgf("can't parse object id: not an IPv4/IPv6")
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't parse object id: not an IPv4/IPv6"})
		return
	}

	var pdrElement PdrDownlinkElement
	pdrElement.Id = ip.String()
	if err := c.BindJSON(&pdrElement); err != nil {
		log.Warn().Err(err).Msgf("can't parse request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't parse request body"})
		return
	}

	var value = ebpf.PdrInfo{
		OuterHeaderRemoval: pdrElement.OuterHeaderRemoval,
		FarId:              pdrElement.FarId,
		QerId:              pdrElement.QerId,
		TraceFlag:          pdrElement.Trace,
	}

	if err := h.BpfObjects.PutPdrDownlink(ip, value); err != nil {
		log.Warn().Err(err).Msgf("can't set pdr")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "can't set pdr"})
		return
	}

	c.IndentedJSON(http.StatusCreated, pdrElement)
}
