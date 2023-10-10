package core

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"unsafe"

	"github.com/edgecomllc/eupf/cmd/ebpf"

	"github.com/edgecomllc/eupf/cmd/config"
	eupfDocs "github.com/edgecomllc/eupf/cmd/docs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @BasePath /api/v1

type ApiServer struct {
	router *gin.Engine
}

func CreateApiServer(bpfObjects *ebpf.BpfObjects, pfcpSrv *PfcpConnection, forwardPlaneStats ebpf.UpfXdpActionStatistic) *ApiServer {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	eupfDocs.SwaggerInfo.BasePath = "/api/v1"
	v1 := router.Group("/api/v1")
	{
		v1.GET("upf_pipeline", ListUpfPipeline(bpfObjects))
		v1.GET("config", DisplayConfig())
		v1.GET("xdp_stats", DisplayXdpStatistics(forwardPlaneStats))
		v1.GET("packet_stats", DisplayPacketStats(forwardPlaneStats))
		v1.POST("config", EditConfig)

		qerMap := v1.Group("qer_map")
		{
			qerMap.GET("", ListQerMapContent(bpfObjects))
			qerMap.GET(":id", GetQerContent(bpfObjects))
		}

		associations := v1.Group("pfcp_associations")
		{
			associations.GET("", ListPfcpAssociations(pfcpSrv))
			associations.GET("full", ListPfcpAssociationsFull(pfcpSrv))
		}

		sessions := v1.Group("pfcp_sessions")
		{
			//sessions.GET("", ListPfcpSessions(pfcpSrv))
			sessions.GET("", ListPfcpSessionsFiltered(pfcpSrv))
		}
	}

	router.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return &ApiServer{router: router}
}

type EditConfigRequest struct {
	ConfigName  string `json:"config_name"`
	ConfigValue string `json:"config_value"`
}

func EditConfig(c *gin.Context) {
	var editConfigRequest EditConfigRequest
	if err := c.BindJSON(&editConfigRequest); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "Request body should have config_name and config_value fields",
		})
		return
	}
	switch editConfigRequest.ConfigName {
	case "logging_level":
		if err := SetLoggerLevel(editConfigRequest.ConfigValue); err != nil {
			c.IndentedJSON(http.StatusBadRequest,
				gin.H{
					"message": fmt.Sprintf("Logger configuring error: %s. Using '%s' level",
						err.Error(), zerolog.GlobalLevel().String()),
				})
		} else {
			c.Status(http.StatusOK)
		}
	default:
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Unsupported config_name"})
	}
}

type PacketStats struct {
	RxArp      uint64 `json:"rx_arp"`
	RxIcmp     uint64 `json:"rx_icmp"`
	RxIcmp6    uint64 `json:"rx_icmp6"`
	RxIp4      uint64 `json:"rx_ip4"`
	RxIp6      uint64 `json:"rx_ip6"`
	RxTcp      uint64 `json:"rx_tcp"`
	RxUdp      uint64 `json:"rx_udp"`
	RxOther    uint64 `json:"rx_other"`
	RxGtpEcho  uint64 `json:"rx_gtp_echo"`
	RxGtpPdu   uint64 `json:"rx_gtp_pdu"`
	RxGtpOther uint64 `json:"rx_gtp_other"`
	RxGtpUnexp uint64 `json:"rx_gtp_unexp"`
}

// DisplayPacketStats godoc
// @Summary Display packet statistics
// @Description Display packet statistics
// @Tags Packet
// @Produce  json
// @Success 200 {object} PacketStats
// @Router /packet_stats [get]
func DisplayPacketStats(forwardPlaneStats ebpf.UpfXdpActionStatistic) func(c *gin.Context) {
	return func(c *gin.Context) {
		packets := forwardPlaneStats.GetUpfExtStatField()
		c.IndentedJSON(http.StatusOK, PacketStats{
			RxArp:      packets.RxArp,
			RxIcmp:     packets.RxIcmp,
			RxIcmp6:    packets.RxIcmp6,
			RxIp4:      packets.RxIp4,
			RxIp6:      packets.RxIp6,
			RxTcp:      packets.RxTcp,
			RxUdp:      packets.RxUdp,
			RxOther:    packets.RxOther,
			RxGtpEcho:  packets.RxGtpEcho,
			RxGtpPdu:   packets.RxGtpPdu,
			RxGtpOther: packets.RxGtpOther,
			RxGtpUnexp: packets.RxGtpUnexp,
		})
	}
}

type XdpStats struct {
	Aborted  uint64 `json:"aborted"`
	Drop     uint64 `json:"drop"`
	Pass     uint64 `json:"pass"`
	Tx       uint64 `json:"tx"`
	Redirect uint64 `json:"redirect"`
}

// DisplayXdpStatistics godoc
// @Summary Display XDP statistics
// @Description Display XDP statistics
// @Tags XDP
// @Produce  json
// @Success 200 {object} XdpStats
// @Router /xdp_stats [get]
func DisplayXdpStatistics(forwardPlaneStats ebpf.UpfXdpActionStatistic) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, XdpStats{
			Aborted:  forwardPlaneStats.GetAborted(),
			Drop:     forwardPlaneStats.GetDrop(),
			Pass:     forwardPlaneStats.GetPass(),
			Tx:       forwardPlaneStats.GetTx(),
			Redirect: forwardPlaneStats.GetRedirect(),
		})
	}
}

// DisplayConfig godoc
// @Summary Display configuration
// @Description Display configuration
// @Tags Configuration
// @Produce  json
// @Success 200 {object} config.UpfConfig
// @Router /config [get]
func DisplayConfig() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, config.Conf)
	}
}

// ListPfcpAssociationsFull godoc
// @Summary List PFCP associations
// @Description List PFCP associations
// @Tags PFCP
// @Produce  json
// @Success 200 {object} map[string]core.NodeAssociation
// @Router /pfcp_associations/full [get]
func ListPfcpAssociationsFull(pfcpSrv *PfcpConnection) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, pfcpSrv.NodeAssociations)
	}
}

type NodeAssociationNoSession struct {
	ID            string
	Addr          string
	NextSessionID uint64
}
type NodeAssociationMapNoSession map[string]NodeAssociationNoSession

// ListPfcpAssociations godoc
// @Summary List PFCP associations
// @Description List PFCP associations
// @Tags PFCP
// @Produce  json
// @Success 200 {object} NodeAssociationMapNoSession
// @Router /pfcp_associations [get]
func ListPfcpAssociations(pfcpSrv *PfcpConnection) func(c *gin.Context) {
	return func(c *gin.Context) {
		nodeAssociationsNoSession := make(NodeAssociationMapNoSession)
		for k, v := range pfcpSrv.NodeAssociations {
			nodeAssociationsNoSession[k] = NodeAssociationNoSession{
				ID:            v.ID,
				Addr:          v.Addr,
				NextSessionID: v.NextSessionID,
			}
		}
		c.IndentedJSON(http.StatusOK, nodeAssociationsNoSession)
	}
}

func GetAllSessions(nodeMap *map[string]*NodeAssociation) (sessions []Session) {
	for _, nodeAssoc := range *nodeMap {
		for _, session := range nodeAssoc.Sessions {
			sessions = append(sessions, *session)
		}
	}
	return
}

func FilterSessionsByIP(nodeMap *map[string]*NodeAssociation, filterByIP net.IP) *Session {
	for _, nodeAssoc := range *nodeMap {
		for _, session := range nodeAssoc.Sessions {
			for _, PDR := range session.PDRs {
				if PDR.Ipv4.Equal(filterByIP) {
					return session
				}
			}
		}
	}
	return nil
}

func FilterSessionsByTeid(nodeMap *map[string]*NodeAssociation, filterByTeid uint32) *Session {
	for _, nodeAssoc := range *nodeMap {
		for _, session := range nodeAssoc.Sessions {
			for _, PDR := range session.PDRs {
				if PDR.Teid == filterByTeid {
					return session
				}
			}
		}
	}
	return nil
}

// ListPfcpSessionsFiltered godoc
// @Summary If no parameters are given, list all PFCP sessions. If ip or teid is given, single session will be returned. If both ip and teid are given, it is possible to return two sessions.
// @Tags PFCP
// @Produce  json
// @Param ip query string false "ip"
// @Param teid query int false "teid"
// @Success 200 {object} []core.Session
// @Router /pfcp_sessions [get]
func ListPfcpSessionsFiltered(pfcpSrv *PfcpConnection) func(c *gin.Context) {
	return func(c *gin.Context) {
		var sessions []Session
		sIp := c.Query("ip")
		sTeid := c.Query("teid")
		if sIp == "" && sTeid == "" {
			sessions = GetAllSessions(&pfcpSrv.NodeAssociations)
			c.IndentedJSON(http.StatusOK, sessions)
			return // early return if no parameters are given
		}

		if sIp != "" {
			if ip := net.ParseIP(sIp); ip != nil {
				if session := FilterSessionsByIP(&pfcpSrv.NodeAssociations, ip); session != nil {
					sessions = append(sessions, *session) // Append session by IP match
				}
			} else {
				c.IndentedJSON(http.StatusBadRequest, "Failed to parse IP")
			}
		}

		if sTeid != "" {
			if teid, err := strconv.Atoi(sTeid); err == nil {
				if session := FilterSessionsByTeid(&pfcpSrv.NodeAssociations, uint32(teid)); session != nil {
					sessions = append(sessions, *session) // Append session by TEID match
				}
			} else {
				c.IndentedJSON(http.StatusBadRequest, "Failed to parse TEID")
			}
		}
		c.IndentedJSON(http.StatusOK, sessions)
	}
}

// ListQerMapContent godoc
// @Summary List QER map content
// @Description List QER map content
// @Tags QER
// @Produce  json
// @Success 200 {object} []ebpf.QerMapElement
// @Router /qer_map [get]
func ListQerMapContent(bpfObjects *ebpf.BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {
		if elements, err := ebpf.ListQerMapContents(bpfObjects.IpEntrypointObjects.QerMap); err != nil {
			log.Info().Msgf("Error reading map: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.IndentedJSON(http.StatusOK, elements)
		}
	}
}

// GetQerContent godoc
// @Summary List QER map content
// @Description List QER map content
// @Tags QER
// @Produce  json
// @Param id path int true "Qer ID"
// @Success 200 {object} []ebpf.QerMapElement
// @Router /qer_map/{id} [get]
func GetQerContent(bpfObjects *ebpf.BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		aid, err := strconv.Atoi(id)
		if err != nil {
			log.Info().Msgf("Error converting id to int: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var value ebpf.QerInfo

		if err = bpfObjects.IpEntrypointObjects.QerMap.Lookup(uint32(aid), unsafe.Pointer(&value)); err != nil {
			log.Info().Msgf("Error reading map: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, ebpf.QerMapElement{
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
// @Success 200 {object} []ebpf.BpfMapProgArrayMember
// @Router /upf_pipeline [get]
func ListUpfPipeline(bpfObjects *ebpf.BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {
		if elements, err := ebpf.ListMapProgArrayContents(bpfObjects.UpfXdpObjects.UpfPipeline); err != nil {
			log.Info().Msgf("Error reading map: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.IndentedJSON(http.StatusOK, elements)
		}
	}
}

func (server *ApiServer) Run(addr string) error {
	return server.router.Run(addr)
}
