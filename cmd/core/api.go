package core

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"unsafe"

	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/edgecomllc/eupf/cmd/ebpf"

	eupfDocs "github.com/edgecomllc/eupf/cmd/docs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
		v1.GET("xdp_stats", DisplayXdpStatistics(forwardPlaneStats))
		v1.GET("packet_stats", DisplayPacketStats(forwardPlaneStats))

		config := v1.Group("config")
		{
			config.GET("", DisplayConfig())
			config.POST("", EditConfig)
		}

		pdrMap := v1.Group("uplink_pdr_map")
		{
			pdrMap.GET(":id", GetUplinkPdrValue(bpfObjects))
			pdrMap.PUT(":id", SetUplinkPdrValue(bpfObjects))
		}

		qerMap := v1.Group("qer_map")
		{
			qerMap.GET("", ListQerMapContent(bpfObjects))
			qerMap.GET(":id", GetQerValue(bpfObjects))
			qerMap.PUT(":id", SetQerValue(bpfObjects))
		}

		farMap := v1.Group("far_map")
		{
			farMap.GET(":id", GetFarValue(bpfObjects))
			farMap.PUT(":id", SetFarValue(bpfObjects))
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

func EditConfig(c *gin.Context) {
	var conf config.UpfConfig
	if err := c.BindJSON(&conf); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message":       "Request body json has incorrect format. Use one or more fields from the following structure",
			"correctFormat": config.UpfConfig{},
		})
		return
	}
	if err := SetConfig(conf); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Error during editing config: %s", err.Error()),
		})
	} else {
		c.Status(http.StatusOK)
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

// GetQerValue godoc
// @Summary List QER map content
// @Description List QER map content
// @Tags QER
// @Produce  json
// @Param id path int true "Qer ID"
// @Success 200 {object} []ebpf.QerMapElement
// @Router /qer_map/{id} [get]
func GetQerValue(bpfObjects *ebpf.BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Info().Msgf("Error converting id to int: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var value ebpf.QerInfo

		if err = bpfObjects.IpEntrypointObjects.QerMap.Lookup(uint32(id), unsafe.Pointer(&value)); err != nil {
			log.Printf("Error reading map: %s", err.Error())
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, ebpf.QerMapElement{
			Id:           uint32(id),
			GateStatusUL: value.GateStatusUL,
			GateStatusDL: value.GateStatusDL,
			Qfi:          value.Qfi,
			MaxBitrateUL: value.MaxBitrateUL,
			MaxBitrateDL: value.MaxBitrateDL,
		})
	}
}

func SetQerValue(bpfObjects *ebpf.BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {

		var qerElement ebpf.QerMapElement
		if err := c.BindJSON(&qerElement); err != nil {
			log.Printf("Parsing request body error: %s", err.Error())
			return
		}

		var value = ebpf.QerInfo{
			GateStatusUL: qerElement.GateStatusUL,
			GateStatusDL: qerElement.GateStatusDL,
			Qfi:          qerElement.Qfi,
			MaxBitrateUL: qerElement.MaxBitrateUL,
			MaxBitrateDL: qerElement.MaxBitrateDL,
			StartUL:      0,
			StartDL:      0,
		}

		if err := bpfObjects.IpEntrypointObjects.QerMap.Put(uint32(qerElement.Id), unsafe.Pointer(&value)); err != nil {
			log.Printf("Error writting map: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusCreated, qerElement)
	}
}

type FarMapElement struct {
	Id                    uint32 `json:"id"`
	Action                uint8  `json:"action"`
	OuterHeaderCreation   uint8  `json:"outer_header_creation"`
	Teid                  uint32 `json:"teid"`
	RemoteIP              uint32 `json:"remote_ip"`
	LocalIP               uint32 `json:"local_ip"`
	TransportLevelMarking uint16 `json:"transport_level_marking"`
}

func GetFarValue(bpfObjects *ebpf.BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Printf("Not an integer id: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var value ebpf.FarInfo
		if err = bpfObjects.IpEntrypointObjects.FarMap.Lookup(uint32(id), unsafe.Pointer(&value)); err != nil {
			log.Printf("Error reading map: %s", err.Error())
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, FarMapElement{
			Id:                    uint32(id),
			Action:                value.Action,
			OuterHeaderCreation:   value.OuterHeaderCreation,
			Teid:                  value.Teid,
			RemoteIP:              value.RemoteIP,
			LocalIP:               value.LocalIP,
			TransportLevelMarking: value.TransportLevelMarking,
		})
	}
}

func SetFarValue(bpfObjects *ebpf.BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {

		var farElement FarMapElement
		if err := c.BindJSON(&farElement); err != nil {
			log.Printf("Parsing request body error: %s", err.Error())
			return
		}

		var value = ebpf.FarInfo{
			Action:                farElement.Action,
			OuterHeaderCreation:   farElement.OuterHeaderCreation,
			Teid:                  farElement.Teid,
			RemoteIP:              farElement.RemoteIP,
			LocalIP:               farElement.LocalIP,
			TransportLevelMarking: farElement.TransportLevelMarking,
		}

		if err := bpfObjects.IpEntrypointObjects.FarMap.Put(uint32(farElement.Id), unsafe.Pointer(&value)); err != nil {
			log.Printf("Error writting map: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusCreated, farElement)
	}
}

type PdrElement struct {
	Id                 uint32 `json:"id"`
	OuterHeaderRemoval uint8  `json:"outer_header_removal"`
	FarId              uint32 `json:"far_id"`
	QerId              uint32 `json:"qer_id"`
}

func GetUplinkPdrValue(bpfObjects *ebpf.BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Printf("Not an integer id: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var value ebpf.PdrInfo
		if err = bpfObjects.IpEntrypointObjects.PdrMapUplinkIp4.Lookup(uint32(id), unsafe.Pointer(&value)); err != nil {
			log.Printf("Error reading map: %s", err.Error())
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, PdrElement{
			Id:                 uint32(id),
			OuterHeaderRemoval: value.OuterHeaderRemoval,
			FarId:              value.FarId,
			QerId:              value.QerId,
		})
	}
}

func SetUplinkPdrValue(bpfObjects *ebpf.BpfObjects) func(c *gin.Context) {
	return func(c *gin.Context) {

		var pdrElement PdrElement
		if err := c.BindJSON(&pdrElement); err != nil {
			log.Printf("Parsing request body error: %s", err.Error())
			return
		}

		var value = ebpf.PdrInfo{
			OuterHeaderRemoval: pdrElement.OuterHeaderRemoval,
			FarId:              pdrElement.FarId,
			QerId:              pdrElement.QerId,
		}

		if err := bpfObjects.IpEntrypointObjects.PdrMapUplinkIp4.Put(uint32(pdrElement.Id), unsafe.Pointer(&value)); err != nil {
			log.Printf("Error writting map: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusCreated, pdrElement)
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
