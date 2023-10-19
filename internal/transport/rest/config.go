package rest

import (
	"fmt"
	"github.com/edgecomllc/eupf/config"
	"github.com/edgecomllc/eupf/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

// DisplayConfig godoc
// @Summary Display configuration
// @Description Display configuration
// @Tags Configuration
// @Produce  json
// @Success 200 {object} config.UpfConfig
// @Router /config [get]
func (h *Handler) displayConfig(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, *h.Cfg)
}

func (h *Handler) editConfig(c *gin.Context) {
	var conf config.Config
	if err := c.BindJSON(&conf); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message":       "Request body json has incorrect format. Use one or more fields from the following structure",
			"correctFormat": config.Config{},
		})
		return
	}

	*h.Cfg = conf

	if err := logger.SetLoggerLevel(conf.LoggingLevel); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Error during editing config: %s", err.Error()),
		})
	} else {
		c.Status(http.StatusOK)
	}
}
