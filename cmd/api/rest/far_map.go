package rest

import (
	"net/http"
	"strconv"
	"unsafe"

	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type FarMapElement struct {
	Id                    uint32 `json:"id"`
	Action                uint8  `json:"action"`
	OuterHeaderCreation   uint8  `json:"outer_header_creation"`
	Teid                  uint32 `json:"teid"`
	RemoteIP              uint32 `json:"remote_ip"`
	TransportLevelMarking uint16 `json:"transport_level_marking"`
}

// GetFarValue godoc
//
//	@Summary Get FAR map element
//	@Description Retrieve FAR map element by ID
//	@Tags FAR
//	@Produce json
//	@Param id path int true "FAR ID"
//	@Success 200 {object} FarMapElement
//	@Failure 400 {object} map[string]string
//	@Failure 404 {object} map[string]string
//	@Router /far_map/{id} [get]
func (h *ApiHandler) getFarValue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		log.Info().Msgf("Error converting id to uint32: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var value ebpf.FarInfo
	if err = h.BpfObjects.IpEntrypointObjects.FarMap.Lookup(uint32(id), unsafe.Pointer(&value)); err != nil {
		log.Printf("Error reading map: %s", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, FarMapElement{
		Id:                    uint32(id),
		Action:                value.Action,
		OuterHeaderCreation:   value.OuterHeaderCreation,
		Teid:                  value.Teid,
		RemoteIP:              value.RemoteIP,
		TransportLevelMarking: value.TransportLevelMarking,
	})
}

// SetFarValue godoc
//
//	@Summary Set FAR map element
//	@Description Create or update FAR map element
//	@Tags FAR
//	@Accept json
//	@Produce json
//	@Param id path int true "FAR ID"
//	@Param far body FarMapElement true "FAR element data"
//	@Success 201 {object} FarMapElement
//	@Failure 400 {object} map[string]string
//	@Failure 500 {object} map[string]string
//	@Router /far_map/{id} [put]
func (h *ApiHandler) setFarValue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		log.Info().Msgf("Error converting id to uint32: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var farElement FarMapElement
	if err := c.BindJSON(&farElement); err != nil {
		log.Printf("Parsing request body error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var value = ebpf.FarInfo{
		Action:                farElement.Action,
		OuterHeaderCreation:   farElement.OuterHeaderCreation,
		Teid:                  farElement.Teid,
		RemoteIP:              farElement.RemoteIP,
		TransportLevelMarking: farElement.TransportLevelMarking,
	}

	if err := h.BpfObjects.IpEntrypointObjects.FarMap.Put(uint32(id), unsafe.Pointer(&value)); err != nil {
		log.Printf("Error writting map: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, farElement)
}
