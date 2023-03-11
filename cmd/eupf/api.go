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

	return &ApiServer{router: router}
}

func (server *ApiServer) Run(addr string) {
	server.router.Run(addr)
}
