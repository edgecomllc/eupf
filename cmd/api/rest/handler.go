package rest

import (
	"net/http"

	"github.com/cilium/ebpf/link"
	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/edgecomllc/eupf/cmd/core"
	"github.com/edgecomllc/eupf/cmd/ebpf"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@BasePath	/api/v1

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

type ApiHandler struct {
	BpfObjects        *ebpf.BpfObjects
	pfcpSrv           map[string]*core.PfcpConnection
	ForwardPlaneStats *ebpf.UpfXdpActionStatistic
	Cfg               *config.UpfConfig
	Links             *[]link.Link
	GtpPathManager    *core.GtpPathManager
}

func NewApiHandler(
	bpfObjects *ebpf.BpfObjects,
	pfcpSrv map[string]*core.PfcpConnection,
	forwardPlaneStats *ebpf.UpfXdpActionStatistic,
	cfg *config.UpfConfig,
	links *[]link.Link,
	gtpPathManager *core.GtpPathManager,
) *ApiHandler {
	return &ApiHandler{
		BpfObjects:        bpfObjects,
		pfcpSrv:           pfcpSrv,
		ForwardPlaneStats: forwardPlaneStats,
		Cfg:               cfg,
		Links:             links,
		GtpPathManager:    gtpPathManager,
	}
}

func (h *ApiHandler) GetPFCPSrv() map[string]*core.PfcpConnection {
	return h.pfcpSrv
}

func (h *ApiHandler) InitRoutes() *gin.Engine {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.IndentedJSON(http.StatusOK, "OK")
		})

		h.initDefaultRoutes(v1)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

func (h *ApiHandler) initDefaultRoutes(group *gin.RouterGroup) {

	group.GET("xdp_stats", h.displayXdpStatistics)
	group.GET("packet_stats", h.displayPacketStats)
	group.GET("route_stats", h.displayRouteStats)

	config := group.Group("config")
	{
		config.GET("", h.displayConfig)
		config.POST("/logging_level", h.editLoggingLevelConfig)
		config.POST("/logging_caller", h.editLoggingCallerConfig)
		config.POST("/dataplane_ebpf", h.editDataPlaneConfig)
		config.POST("/dataplane_addresses", h.editDataPlaneAddressesConfig)
		config.POST("/pfcp_n4", h.editPFCPN4Config)
		config.POST("/pfcp_sxa", h.editPFCPSxaConfig)
		config.POST("/pfcp_sxb", h.editPFCPSxbConfig)
		config.POST("/pfcp_timers", h.editPFCPTimersConfig)
		config.POST("/gtp_path", h.editGTPPathConfig)
	}

	pdrMap := group.Group("uplink_pdr_map")
	{
		pdrMap.GET(":id", h.getUplinkPdrValue)
		pdrMap.PUT(":id", h.setUplinkPdrValue)
	}

	pdrDownlinkMap := group.Group("downlink_pdr_map")
	{
		pdrDownlinkMap.GET(":id", h.getDownlinkPdrValue)
		pdrDownlinkMap.PUT(":id", h.setDownlinkPdrValue)
	}

	qerMap := group.Group("qer_map")
	{
		qerMap.GET("", h.listQerMapContent)
		qerMap.GET(":id", h.getQerValue)
		qerMap.PUT(":id", h.setQerValue)
	}

	farMap := group.Group("far_map")
	{
		farMap.GET(":id", h.getFarValue)
		farMap.PUT(":id", h.setFarValue)
	}

	associations := group.Group("pfcp_associations")
	{
		associations.GET("", h.listPfcpAssociations)
		associations.GET("full", h.listPfcpAssociationsFull)
	}

	sessions := group.Group("pfcp_sessions")
	{
		sessions.GET("", h.listPfcpSessionsFiltered)
		sessions.DELETE("", h.deletePfcpSessions)
	}

	subscriberTracing := group.Group("subscriber_trace")
	{
		subscriberTracing.GET("", h.listTraces)
		subscriberTracing.POST("", h.startTrace)
		subscriberTracing.DELETE("", h.stopTrace)
	}
}

func (h *ApiHandler) InitMetricsRoute() *gin.Engine {
	pfcpSrv := h.GetPFCPSrv()
	conns := make([]*core.PfcpConnection, 0, len(pfcpSrv))

	for _, conn := range pfcpSrv {
		conns = append(conns, conn)
	}

	core.RegisterMetrics(*h.ForwardPlaneStats, conns)

	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	router.GET("/metrics", func() gin.HandlerFunc {
		return func(c *gin.Context) {
			core.GatherMetrics(*h.ForwardPlaneStats)
			promhttp.Handler().ServeHTTP(c.Writer, c.Request)
		}
	}())

	return router
}
