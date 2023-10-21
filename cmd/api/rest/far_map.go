package rest

import (
	"github.com/edgecomllc/eupf/cmd/domain"
	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"unsafe"
)

func (h *ApiHandler) getFarValue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Not an integer id: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var value ebpf.FarInfo
	if err = h.BpfObjects.IpEntrypointObjects.FarMap.Lookup(uint32(id), unsafe.Pointer(&value)); err != nil {
		log.Printf("Error reading map: %s", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, domain.FarMapElement{
		Id:                    uint32(id),
		Action:                value.Action,
		OuterHeaderCreation:   value.OuterHeaderCreation,
		Teid:                  value.Teid,
		RemoteIP:              value.RemoteIP,
		LocalIP:               value.LocalIP,
		TransportLevelMarking: value.TransportLevelMarking,
	})
}

func (h *ApiHandler) setFarValue(c *gin.Context) {
	var farElement domain.FarMapElement
	if err := c.BindJSON(&farElement); err != nil {
		log.Printf("Parsing request body error: %s", err.Error())
		return
	}

	var value = ebpf.FarInfo{
		Action:                farElement.Action,
		OuterHeaderCreation:   farElement.OuterHeaderCreation,
		Teid:                  farElement.Teid,
		RemoteIP:              farElement.RemoteIP,
		LocalIP:               farElement.LocalIP,
		TransportLevelMarking: farElement.TransportLevelMarking,
	}

	if err := h.BpfObjects.IpEntrypointObjects.FarMap.Put(uint32(farElement.Id), unsafe.Pointer(&value)); err != nil {
		log.Printf("Error writting map: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, farElement)
}
