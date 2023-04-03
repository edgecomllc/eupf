package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ApiServer struct {
	router *gin.Engine
}

func CreateApiServer(bpfObjects *BpfObjects, pfcp_srv *PfcpConnection, forwardPlaneStats UpfXdpActionStatistic) *ApiServer {
	router := gin.Default()
	router.GET("/upf_pipeline", func(c *gin.Context) {
		elements, err := ListMapProgArrayContents(bpfObjects.upf_xdpObjects.UpfPipeline)
		if err != nil {
			log.Printf("Error reading map: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, elements)
	})
	router.GET("/context_map", func(c *gin.Context) {
		elements, err := ListContextMapContents(bpfObjects.ip_entrypointObjects.ContextMapIp4)
		if err != nil {
			log.Printf("Error reading map: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, elements)
	})
	router.GET("/qer_map", func(c *gin.Context) {
		elements, err := ListQerMapContents(bpfObjects.ip_entrypointObjects.QerMap)
		if err != nil {
			log.Printf("Error reading map: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, elements)
	})
	router.GET("/pfcp_associations", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, pfcp_srv.nodeAssociations)
	})
	router.GET("/config", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, config)
	})
	router.GET("/xdp_stats", func(c *gin.Context) {
		type XdpStats struct {
			Aborted  uint32 `json:"aborted"`
			Drop     uint32 `json:"drop"`
			Pass     uint32 `json:"pass"`
			Tx       uint32 `json:"tx"`
			Redirect uint32 `json:"redirect"`
		}

		xdpStats := XdpStats{
			Aborted:  forwardPlaneStats.GetAborted(),
			Drop:     forwardPlaneStats.GetDrop(),
			Pass:     forwardPlaneStats.GetPass(),
			Tx:       forwardPlaneStats.GetTx(),
			Redirect: forwardPlaneStats.GetRedirect(),
		}
		c.IndentedJSON(http.StatusOK, xdpStats)
	})
	return &ApiServer{router: router}
}

func (server *ApiServer) Run(addr string) {
	server.router.Run(addr)
}
