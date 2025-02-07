package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/edgecomllc/eupf/cmd/core/tracing"
	"github.com/gin-gonic/gin"
)

func (h *ApiHandler) verifyTraceArguments(imsi, msisdn string) (bool, error) {
	if imsi != "" {
		if _, err := strconv.ParseUint(imsi, 10, 64); err != nil {
			return false, fmt.Errorf("incorrect IMSI: %s. Should contain digits only", imsi)
		}
	}

	if msisdn != "" {
		if _, err := strconv.ParseUint(msisdn, 10, 64); err != nil {
			return false, fmt.Errorf("incorrect MSISDN: %s. Should contain digits only", msisdn)
		}
	}

	return true, nil
}

func (h *ApiHandler) listTraces(c *gin.Context) {
	imsi := c.Query("imsi")
	msisdn := c.Query("msisdn")

	if _, err := h.verifyTraceArguments(imsi, msisdn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var subsList []tracing.TraceRecord
	for _, conn := range h.PfcpSrv {
		if imsi != "" {
			if s, err := conn.GetTracingRecordByImsi(imsi); err == nil {
				subsList = append(subsList, s)
			}
		} else if msisdn != "" {
			if s, err := conn.GetTracingRecordByMsisdn(msisdn); err == nil {
				subsList = append(subsList, s)
			}
		} else {
			traces, _ := conn.GetAllTracingRecords()
			subsList = append(subsList, traces...)
		}
	}

	if len(subsList) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	c.IndentedJSON(http.StatusOK, subsList)
}

func (h *ApiHandler) startTrace(c *gin.Context) {
	imsi := c.Query("imsi")
	msisdn := c.Query("msisdn")

	if imsi == "" && msisdn == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing parameter. IMSI or MSISDN required"})
		return
	}

	if _, err := h.verifyTraceArguments(imsi, msisdn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, conn := range h.PfcpSrv {
		if imsi != "" {
			_ = conn.EnableTracingByImsi(imsi)
		}

		if msisdn != "" {
			_ = conn.EnableTracingByMsisdn(msisdn)
		}
	}

	c.Status(http.StatusCreated)
}

func (h *ApiHandler) stopTrace(c *gin.Context) {
	imsi := c.Query("imsi")
	msisdn := c.Query("msisdn")

	if _, err := h.verifyTraceArguments(imsi, msisdn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var subsList []tracing.TraceRecord
	for _, conn := range h.PfcpSrv {
		if imsi != "" {
			if s, err := conn.DisableTracingByImsi(imsi); err == nil {
				subsList = append(subsList, s)
			}
		} else if msisdn != "" {
			if s, err := conn.DisableTracingByMsisdn(msisdn); err == nil {
				subsList = append(subsList, s)
			}
		} else {
			traces, _ := conn.DisableTracing()
			subsList = append(subsList, traces...)
		}
	}

	if len(subsList) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	c.IndentedJSON(http.StatusOK, subsList)
}
