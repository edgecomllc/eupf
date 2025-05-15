package rest

import (
	"fmt"
	"net"
	"net/http"
	"reflect"

	"github.com/edgecomllc/eupf/cmd/core"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// DisplayConfig godoc
//
//	@Summary		Display configuration
//	@Description	Display configuration
//	@Tags			Configuration
//	@Produce		json
//	@Success		200	{object}	config.UpfConfig
//	@Router			/config [get]
func (h *ApiHandler) displayConfig(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, *h.Cfg)
}

// EditLoggingLevelConfig godoc
//
//	@Summary Update logging level
//	@Description Update the log level (debug, info, warn, error, etc.)
//	@Tags Configuration
//	@Accept json
//	@Produce json
//	@Param config body LoggingLevelConfig true "Logging level config"
//	@Success 200 {object} map[string]string
//	@Failure 400 {object} map[string]string
//	@Router /config/logging_level [post]
func (h *ApiHandler) editLoggingLevelConfig(c *gin.Context) {
	var config LoggingLevelConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := core.UpdateLogLevel(config.LoggingLevel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if err := h.updateConfigFile(config); err != nil {
		log.Error().Msgf("Error updating config: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// EditLoggingCallerConfig godoc
//
//	@Summary Update logging caller info
//	@Description Enable or disable function caller information in logs
//	@Tags Configuration
//	@Accept json
//	@Produce json
//	@Param config body LoggingCallerConfig true "Logging caller configuration"
//	@Success 200 {object} map[string]string
//	@Failure 400 {object} map[string]string
//	@Router /config/logging_caller [post]
func (h *ApiHandler) editLoggingCallerConfig(c *gin.Context) {
	var config LoggingCallerConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	core.UpdateLogCaller(config.LoggingCaller)

	if err := h.updateConfigFile(config); err != nil {
		log.Error().Msgf("Error updating config: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// EditDataPlaneConfig godoc
//
//	@Summary Update dataplane eBPF interfaces
//	@Description Bind network interfaces for eBPF dataplane processing
//	@Tags Configuration
//	@Accept json
//	@Produce json
//	@Param config body DataPlaneEbpfConfig true "Dataplane eBPF configuration"
//	@Success 200 {object} map[string]string
//	@Failure 400 {object} map[string]string
//	@Failure 500 {object} map[string]string
//	@Router /config/dataplane_ebpf [post]
func (h *ApiHandler) editDataPlaneConfig(c *gin.Context) {
	var config DataPlaneEbpfConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := core.BindDataPlaneInterfaces(h.BpfObjects, h.Links, config.InterfaceName, config.XDPAttachMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	if err := h.updateConfigFile(config); err != nil {
		log.Error().Msgf("Error updating config: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// EditDataPlaneAddressesConfig godoc
//
//	@Summary Update dataplane N3 and N9 addresses
//	@Description Update IP addresses for N3 and N9 interfaces
//	@Tags Configuration
//	@Accept json
//	@Produce json
//	@Param config body DataPlaneAddressesConfig true "Dataplane addresses configuration"
//	@Success 200 {object} map[string]string
//	@Failure 400 {object} map[string]string
//	@Failure 500 {object} map[string]string
//	@Router /config/dataplane_addresses [post]
func (h *ApiHandler) editDataPlaneAddressesConfig(c *gin.Context) {
	var config DataPlaneAddressesConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	n3Addr := net.ParseIP(config.N3Address)
	if n3Addr == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "n3_address is not a valid IP"})

		return
	}

	n9Addr := net.ParseIP(config.N9Address)
	if n9Addr == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "n9_address is not a valid IP"})

		return
	}

	err := core.UpdateDataPlaneAddresses(h.GetPFCPSrv(), h.GtpPathManager, n3Addr, n9Addr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	if err := h.updateConfigFile(config); err != nil {
		log.Error().Msgf("Error updating config: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// EditPFCPN4Config godoc
//
//	@Summary Update N4 PFCP connection settings
//	@Description Update local and remote node information for PFCP N4 connection
//	@Tags Configuration
//	@Accept json
//	@Produce json
//	@Param config body PFCPN4Config true "PFCP N4 connection configuration"
//	@Success 200 {object} map[string]string
//	@Failure 400 {object} map[string]string
//	@Failure 500 {object} map[string]string
//	@Router /config/pfcp_n4 [post]
func (h *ApiHandler) editPFCPN4Config(c *gin.Context) {
	var config PFCPN4Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pfcpSrv := h.GetPFCPSrv()

	if _, exists := pfcpSrv[core.N4PFCPKeyName]; exists {
		err := core.UpdatePFCPConnections(
			pfcpSrv,
			core.N4PFCPKeyName,
			config.PFCPAddress,
			config.PFCPNodeID,
			config.PFCPRemoteNode,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PFCP connection for N4 not found"})

		return
	}

	if err := h.updateConfigFile(config); err != nil {
		log.Error().Msgf("Error updating config: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// EditPFCPSxaConfig godoc
//
//	@Summary Update Sxa PFCP connection settings
//	@Description Update local and remote node information for PFCP Sxa connection
//	@Tags Configuration
//	@Accept json
//	@Produce json
//	@Param config body PFCPSxaConfig true "PFCP Sxa connection configuration"
//	@Success 200 {object} map[string]string
//	@Failure 400 {object} map[string]string
//	@Failure 500 {object} map[string]string
//	@Router /config/pfcp_sxa [post]
func (h *ApiHandler) editPFCPSxaConfig(c *gin.Context) {
	var config PFCPSxaConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pfcpSrv := h.GetPFCPSrv()

	if _, exists := pfcpSrv[core.SxaPFCPKeyName]; exists {
		err := core.UpdatePFCPConnections(
			pfcpSrv,
			core.SxaPFCPKeyName,
			config.SXAAddress,
			config.SXANodeID,
			config.SXARemoteNode,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PFCP connection for Sxa not found"})

		return
	}

	if err := h.updateConfigFile(config); err != nil {
		log.Error().Msgf("Error updating config: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// EditPFCPSxbConfig godoc
//
//	@Summary Update Sxb PFCP connection settings
//	@Description Update local and remote node information for PFCP Sxb connection
//	@Tags Configuration
//	@Accept json
//	@Produce json
//	@Param config body PFCPSxbConfig true "PFCP Sxb connection configuration"
//	@Success 200 {object} map[string]string
//	@Failure 400 {object} map[string]string
//	@Failure 500 {object} map[string]string
//	@Router /config/pfcp_sxb [post]
func (h *ApiHandler) editPFCPSxbConfig(c *gin.Context) {
	var config PFCPSxbConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pfcpSrv := h.GetPFCPSrv()

	if _, exists := pfcpSrv[core.SxbPFCPKeyName]; exists {
		err := core.UpdatePFCPConnections(
			pfcpSrv,
			core.SxbPFCPKeyName,
			config.SXBAddress,
			config.SXBNodeID,
			config.SXBRemoteNode,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PFCP connection for Sxb not found"})

		return
	}

	if err := h.updateConfigFile(config); err != nil {
		log.Error().Msgf("Error updating config: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// EditPFCPTimersConfig godoc
//
//	@Summary Update PFCP timers settings
//	@Description Update heartbeat and association setup timeout values
//	@Tags Configuration
//	@Accept json
//	@Produce json
//	@Param config body PFCPTimersConfig true "PFCP timers configuration"
//	@Success 200 {object} map[string]string
//	@Failure 400 {object} map[string]string
//	@Failure 500 {object} map[string]string
//	@Router /config/pfcp_timers [post]
func (h *ApiHandler) editPFCPTimersConfig(c *gin.Context) {
	var config PFCPTimersConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	pfcpSrv := h.GetPFCPSrv()

	core.UpdateAssociationSetupTimeout(pfcpSrv, config.AssociationSetupTimeout)
	core.UpdateHeartbeatTimeout(pfcpSrv, config.HeartbeatTimeout)

	if err := h.updateConfigFile(config); err != nil {
		log.Error().Msgf("Error updating config: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// EditGTPPathConfig godoc
//
//	@Summary Update GTP path configuration
//	@Description Update GTP peers and echo interval
//	@Tags Configuration
//	@Accept json
//	@Produce json
//	@Param config body GTPPathConfig true "GTP path configuration"
//	@Success 200 {object} map[string]string
//	@Failure 400 {object} map[string]string
//	@Failure 500 {object} map[string]string
//	@Router /config/gtp_path [post]
func (h *ApiHandler) editGTPPathConfig(c *gin.Context) {
	var config GTPPathConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	core.UpdateGTPManager(h.GtpPathManager, config.GtpPeer, config.GtpEchoInterval)

	if err := h.updateConfigFile(config); err != nil {
		log.Error().Msgf("Error updating config: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func (h *ApiHandler) updateConfigFile(data interface{}) error {
	cfgs := make(map[string]interface{})

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		log.Error().Msgf("input must be a struct")

		return fmt.Errorf("input must be a struct")
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		cfgs[jsonTag] = fieldValue.Interface()
	}

	err := h.Cfg.UpdateFile(cfgs)
	if err != nil {
		log.Error().Err(err).Msgf("failed to update config: %s", err.Error())

		return err
	}

	return nil
}
