package rest

import (
	"github.com/edgecomllc/eupf/components/ebpf"
	"github.com/edgecomllc/eupf/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"unsafe"
)

// ListQerMapContent godoc
// @Summary List QER map content
// @Description List QER map content
// @Tags QER
// @Produce  json
// @Success 200 {object} []ebpf.QerMapElement
// @Router /qer_map [get]
func (h *Handler) listQerMapContent(c *gin.Context) {
	if elements, err := ebpf.ListQerMapContents(h.BpfObjects.IpEntrypointObjects.QerMap); err != nil {
		logger.Infof("Error reading map: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.IndentedJSON(http.StatusOK, elements)
	}
}

// GetQerValue godoc
// @Summary List QER map content
// @Description List QER map content
// @Tags QER
// @Produce  json
// @Param id path int true "Qer ID"
// @Success 200 {object} []ebpf.QerMapElement
// @Router /qer_map/{id} [get]
func (h *Handler) getQerValue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Infof("Error converting id to int: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var value ebpf.QerInfo

	if err = h.BpfObjects.IpEntrypointObjects.QerMap.Lookup(uint32(id), unsafe.Pointer(&value)); err != nil {
		logger.Printf("Error reading map: %s", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, ebpf.QerMapElement{
		Id:           uint32(id),
		GateStatusUL: value.GateStatusUL,
		GateStatusDL: value.GateStatusDL,
		Qfi:          value.Qfi,
		MaxBitrateUL: value.MaxBitrateUL,
		MaxBitrateDL: value.MaxBitrateDL,
	})
}

func (h *Handler) setQerValue(c *gin.Context) {
	var qerElement ebpf.QerMapElement
	if err := c.BindJSON(&qerElement); err != nil {
		logger.Printf("Parsing request body error: %s", err.Error())
		return
	}

	var value = ebpf.QerInfo{
		GateStatusUL: qerElement.GateStatusUL,
		GateStatusDL: qerElement.GateStatusDL,
		Qfi:          qerElement.Qfi,
		MaxBitrateUL: qerElement.MaxBitrateUL,
		MaxBitrateDL: qerElement.MaxBitrateDL,
		StartUL:      0,
		StartDL:      0,
	}

	if err := h.BpfObjects.IpEntrypointObjects.QerMap.Put(uint32(qerElement.Id), unsafe.Pointer(&value)); err != nil {
		logger.Printf("Error writting map: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, qerElement)
}
