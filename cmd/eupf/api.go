package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type ApiServer struct {
	router *gin.Engine
}

func CreateApiServer(bpfObjects *BpfObjects, pfcp_srv *PfcpConnection) *ApiServer {
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
			Aborted:  uint32(getMetricValue(UpfXdpAborted)),
			Drop:     uint32(getMetricValue(UpfXdpDrop)),
			Pass:     uint32(getMetricValue(UpfXdpPass)),
			Tx:       uint32(getMetricValue(UpfXdpTx)),
			Redirect: uint32(getMetricValue(UpfXdpRedirect)),
		}
		c.IndentedJSON(http.StatusOK, xdpStats)
	})
	return &ApiServer{router: router}
}

func (server *ApiServer) Run(addr string) {
	server.router.Run(addr)
}

func getMetricValue(col prometheus.Collector) float64 {
	c := make(chan prometheus.Metric, 1) // 1 for metric with no vector
	col.Collect(c)                       // collect current metric value into the channel
	m := dto.Metric{}
	_ = (<-c).Write(&m) // read metric value from the channel
	return *m.Counter.Value
}
