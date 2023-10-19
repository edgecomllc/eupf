package rest

import (
	"github.com/edgecomllc/eupf/components/ebpf"
	"github.com/edgecomllc/eupf/pkg/domain"
	"github.com/edgecomllc/eupf/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"unsafe"
)

func (h *Handler) getUplinkPdrValue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Printf("Not an integer id: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var value ebpf.PdrInfo
	if err = h.BpfObjects.IpEntrypointObjects.PdrMapUplinkIp4.Lookup(uint32(id), unsafe.Pointer(&value)); err != nil {
		logger.Printf("Error reading map: %s", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, domain.PdrElement{
		Id:                 uint32(id),
		OuterHeaderRemoval: value.OuterHeaderRemoval,
		FarId:              value.FarId,
		QerId:              value.QerId,
	})
}

func (h *Handler) setUplinkPdrValue(c *gin.Context) {
	var pdrElement domain.PdrElement
	if err := c.BindJSON(&pdrElement); err != nil {
		logger.Printf("Parsing request body error: %s", err.Error())
		return
	}

	var value = ebpf.PdrInfo{
		OuterHeaderRemoval: pdrElement.OuterHeaderRemoval,
		FarId:              pdrElement.FarId,
		QerId:              pdrElement.QerId,
	}

	if err := h.BpfObjects.IpEntrypointObjects.PdrMapUplinkIp4.Put(uint32(pdrElement.Id), unsafe.Pointer(&value)); err != nil {
		logger.Printf("Error writting map: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, pdrElement)
}
