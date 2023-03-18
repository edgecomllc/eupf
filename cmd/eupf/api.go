package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ApiServer struct {
	router *gin.Engine
}

func CreateApiServer(bpfObjects *BpfObjects, pfcp_srv *PfcpConnection) *ApiServer {
	ForwardPlaneStats := UpfXdpActionStatistic{
		bpfObjects: bpfObjects,
	}
	router := gin.Default()
	router.GET("/upf_pipeline", func(c *gin.Context) {
		elements, err := ListMapProgArrayContents(bpfObjects.upf_xdpObjects.UpfPipeline)
		if err != nil {
			log.Printf("Error reading map: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, elements)
	}).GET("/context_map", func(c *gin.Context) {
		elements, err := ListContextMapContents(bpfObjects.ip_entrypointObjects.ContextMapIp4)
		if err != nil {
			log.Printf("Error reading map: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, elements)
	}).GET("/pfcp_associations", func(c *gin.Context) {
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
		// Fill from existing prometheus metrics
		xdpStats := XdpStats{
			Aborted:  ForwardPlaneStats.GetAborted(),
			Drop:     ForwardPlaneStats.GetDrop(),
			Pass:     ForwardPlaneStats.GetPass(),
			Tx:       ForwardPlaneStats.GetTx(),
			Redirect: ForwardPlaneStats.GetRedirect(),
		}
		c.IndentedJSON(http.StatusOK, xdpStats)
	})
	return &ApiServer{router: router}
}

func (server *ApiServer) Run(addr string) {
	server.router.Run(addr)
}
