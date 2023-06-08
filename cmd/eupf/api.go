package main

import (
	"github.com/edgecomllc/eupf/cmd/eupf/ebpf"
	"log"
	"net"
	"net/http"
	"strconv"
	"unsafe"

	"github.com/edgecomllc/eupf/cmd/eupf/config"
	eupfDocs "github.com/edgecomllc/eupf/cmd/eupf/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//go:generate swag init --parseDependency

// @BasePath /api/v1

type ApiServer struct {
	router *gin.Engine
}

func CreateApiServer(bpfObjects *ebpf.BpfObjects, pfcpSrv *PfcpConnection, forwardPlaneStats ebpf.UpfXdpActionStatistic) *ApiServer {
	router := gin.Default()
	eupfDocs.SwaggerInfo.BasePath = "/api/v1"
	v1 := router.Group("/api/v1")
	{
		v1.GET("upf_pipeline", ListUpfPipeline(bpfObjects))
		v1.GET("config", DisplayConfig())
		v1.GET("xdp_stats", DisplayXdpStatistics(forwardPlaneStats))

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
// @Success 200 {object} NodeAssociationMap
// @Router /pfcp_associations/full [get]
func ListPfcpAssociationsFull(pfcpSrv *PfcpConnection) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, pfcpSrv.nodeAssociations)
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
		for k, v := range pfcpSrv.nodeAssociations {
			nodeAssociationsNoSession[k] = NodeAssociationNoSession{
				ID:            v.ID,
				Addr:          v.Addr,
				NextSessionID: v.NextSessionID,
			}
		}
		c.IndentedJSON(http.StatusOK, nodeAssociationsNoSession)
	}
}

func GetAllSessions(nodeMap *NodeAssociationMap) (sessions []Session) {
	for _, nodeAssoc := range *nodeMap {
		for _, session := range nodeAssoc.Sessions {
			sessions = append(sessions, session)
		}
	}
	return
}

func FilterSessionsByIP(nodeMap *NodeAssociationMap, filterByIP net.IP) *Session {
	for _, nodeAssoc := range *nodeMap {
		for _, session := range nodeAssoc.Sessions {
			for _, uplinkPDR := range session.UplinkPDRs {
				if uplinkPDR.Ipv4.Equal(filterByIP) {
					return &session
				}
			}
			for _, downlinkPDR := range session.DownlinkPDRs {
				if downlinkPDR.Ipv4.Equal(filterByIP) {
					return &session
				}
			}
		}
	}
	return nil
}

func FilterSessionsByTeid(nodeMap *NodeAssociationMap, filterByTeid uint32) *Session {
	for _, nodeAssoc := range *nodeMap {
		for _, session := range nodeAssoc.Sessions {
			for _, uplinkPDR := range session.UplinkPDRs {
				if uplinkPDR.Teid == filterByTeid {
					return &session
				}
			}
			for _, downlinkPDR := range session.DownlinkPDRs {
				if downlinkPDR.Teid == filterByTeid {
					return &session
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
// @Success 200 {object} []Session
// @Router /pfcp_sessions [get]
func ListPfcpSessionsFiltered(pfcpSrv *PfcpConnection) func(c *gin.Context) {
	return func(c *gin.Context) {
		var sessions []Session
		sIp := c.Query("ip")
		sTeid := c.Query("teid")
		if sIp == "" && sTeid == "" {
			sessions = GetAllSessions(&pfcpSrv.nodeAssociations)
			c.IndentedJSON(http.StatusOK, sessions)
			return // early return if no parameters are given
		}

		if sIp != "" {
			if ip := net.ParseIP(sIp); ip != nil {
				if session := FilterSessionsByIP(&pfcpSrv.nodeAssociations, ip); session != nil {
					sessions = append(sessions, *session) // Append session by IP match
				}
			} else {
				c.IndentedJSON(http.StatusBadRequest, "Failed to parse IP")
			}
		}

		if sTeid != "" {
			if teid, err := strconv.Atoi(sTeid); err == nil {
				if session := FilterSessionsByTeid(&pfcpSrv.nodeAssociations, uint32(teid)); session != nil {
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
// @Success 200 {object} []QerMapElement
// @Router /qer_map [get]
func ListQerMapContent(bpfObjects *ebpf.BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {
		if elements, err := ebpf.ListQerMapContents(bpfObjects.Ip_entrypointObjects.QerMap); err != nil {
			log.Printf("Error reading map: %s", err.Error())
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
// @Success 200 {object} []QerMapElement
// @Router /qer_map/{id} [get]
func GetQerContent(bpfObjects *ebpf.BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		aid, err := strconv.Atoi(id)
		if err != nil {
			log.Printf("Error converting id to int: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var value ebpf.QerInfo

		if err = bpfObjects.Ip_entrypointObjects.QerMap.Lookup(uint32(aid), unsafe.Pointer(&value)); err != nil {
			log.Printf("Error reading map: %s", err.Error())
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
// @Success 200 {object} []BpfMapProgArrayMember
// @Router /upf_pipeline [get]
func ListUpfPipeline(bpfObjects *ebpf.BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {
		if elements, err := ebpf.ListMapProgArrayContents(bpfObjects.Upf_xdpObjects.UpfPipeline); err != nil {
			log.Printf("Error reading map: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.IndentedJSON(http.StatusOK, elements)
		}
	}
}

func (server *ApiServer) Run(addr string) error {
	return server.router.Run(addr)
}
