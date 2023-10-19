package rest

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func responseError(c *gin.Context, code int, msg interface{}) {
	response := gin.H{"error": "unexpected error"}

	if msg != nil {
		response["error"] = msg
	}

	c.JSON(code, response)
}
func response200(c *gin.Context, msg interface{}) {
	response := gin.H{"status": "success"}

	if msg != nil {
		response["data"] = msg
	}

	c.JSON(http.StatusOK, response)
}

func response201(c *gin.Context, msg interface{}) {
	response := gin.H{"status": "success"}

	if msg != nil {
		response["data"] = msg
	}

	c.JSON(http.StatusCreated, response)
}
