package main

import (
	eupfDocs "github.com/edgecomllc/eupf/cmd/eupf/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	"strconv"
	"unsafe"
)

//go:generate swag init --parseDependency

// @BasePath /api/v1

type ApiServer struct {
	router *gin.Engine
}

func CreateApiServer(bpfObjects *BpfObjects, pfcpSrv *PfcpConnection, forwardPlaneStats UpfXdpActionStatistic) *ApiServer {
	router := gin.Default()
	eupfDocs.SwaggerInfo.BasePath = "/api/v1"
	v1 := router.Group("/api/v1")
	{
		v1.GET("/upf_pipeline", ListUpfPipeline(bpfObjects))
		qerMap := v1.Group("/qer_map")
		{
			qerMap.GET("", ListQerMapContent(bpfObjects))
			qerMap.GET(":id", GetQerContent(bpfObjects))
		}
		v1.GET("/pfcp_associations", ListPfcpAssociations(pfcpSrv))
		v1.GET("/config", DisplayConfig())
		v1.GET("/xdp_stats", DisplayXdpStatistics(forwardPlaneStats))
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return &ApiServer{router: router}
}

type XdpStats struct {
	Aborted  uint32 `json:"aborted"`
	Drop     uint32 `json:"drop"`
	Pass     uint32 `json:"pass"`
	Tx       uint32 `json:"tx"`
	Redirect uint32 `json:"redirect"`
}

// DisplayXdpStatistics godoc
// @Summary Display XDP statistics
// @Description Display XDP statistics
// @Tags XDP
// @Produce  json
// @Success 200 {object} XdpStats
// @Router /xdp_stats [get]
func DisplayXdpStatistics(forwardPlaneStats UpfXdpActionStatistic) func(c *gin.Context) {
	return func(c *gin.Context) {
		xdpStats := XdpStats{
			Aborted:  forwardPlaneStats.GetAborted(),
			Drop:     forwardPlaneStats.GetDrop(),
			Pass:     forwardPlaneStats.GetPass(),
			Tx:       forwardPlaneStats.GetTx(),
			Redirect: forwardPlaneStats.GetRedirect(),
		}
		c.IndentedJSON(http.StatusOK, xdpStats)
	}
}

// DisplayConfig godoc
// @Summary Display configuration
// @Description Display configuration
// @Tags Configuration
// @Produce  json
// @Success 200 {object} UpfConfig
// @Router /config [get]
func DisplayConfig() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, config)
	}
}

// ListPfcpAssociations godoc
// @Summary List PFCP associations
// @Description List PFCP associations
// @Tags PFCP
// @Produce  json
// @Success 200 {object} NodeAssociationMap
// @Router /pfcp_associations [get]
func ListPfcpAssociations(pfcpSrv *PfcpConnection) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, pfcpSrv.nodeAssociations)
	}
}

// ListQerMapContent godoc
// @Summary List QER map content
// @Description List QER map content
// @Tags QER
// @Produce  json
// @Success 200 {object} []QerMapElement
// @Router /qer_map [get]
func ListQerMapContent(bpfObjects *BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {
		elements, err := ListQerMapContents(bpfObjects.ip_entrypointObjects.QerMap)
		if err != nil {
			log.Printf("Error reading map: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, elements)
	}
}

// GetQerContent godoc
// @Summary List QER map content
// @Description List QER map content
// @Tags QER
// @Produce  json
// @Param id path int true "Qer ID"
// @Success 200 {object} []QerMapElement
// @Router /qer_map/{id} [get]
func GetQerContent(bpfObjects *BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		aid, err := strconv.Atoi(id)
		if err != nil {
			log.Printf("Error converting id to int: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var value QerInfo
		err = bpfObjects.ip_entrypointObjects.QerMap.Lookup(uint32(aid), unsafe.Pointer(&value))
		if err != nil {
			log.Printf("Error reading map: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, QerMapElement{
			Id:           uint32(aid),
			GateStatusUL: value.GateStatusUL,
			GateStatusDL: value.GateStatusDL,
			Qfi:          value.Qfi,
			MaxBitrateUL: value.MaxBitrateUL,
			MaxBitrateDL: value.MaxBitrateDL,
		})
	}
}

// ListUpfPipeline godoc
// @Summary List UPF pipeline
// @Description List UPF pipeline
// @Tags UPF
// @Produce  json
// @Success 200 {object} []BpfMapProgArrayMember
// @Router /upf_pipeline [get]
func ListUpfPipeline(bpfObjects *BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {
		elements, err := ListMapProgArrayContents(bpfObjects.upf_xdpObjects.UpfPipeline)
		if err != nil {
			log.Printf("Error reading map: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, elements)
	}
}

func (server *ApiServer) Run(addr string) error {
	err := server.router.Run(addr)
	return err
}
