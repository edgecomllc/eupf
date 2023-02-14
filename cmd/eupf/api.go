package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ApiServer struct {
	r *gin.Engine
}

func CreateApiServer(bpfObjects *BpfObjects) *ApiServer {
	r := gin.Default()
	r.GET("/upf_pipeline", func(c *gin.Context) {
		elements, err := ListMapProgArrayContents(bpfObjects.upf_xdpObjects.UpfPipeline)
		if err != nil {
			log.Printf("Error reading map: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, elements)
	})
	return &ApiServer{r: r}
}

func (a *ApiServer) Run(addr string) {
	a.r.Run(addr)
}
