package rest

import (
	"github.com/edgecomllc/eupf/components/core"
	"github.com/edgecomllc/eupf/components/ebpf"
	"github.com/edgecomllc/eupf/config"
	eupfDocs "github.com/edgecomllc/eupf/docs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	BpfObjects        *ebpf.BpfObjects
	PfcpSrv           *core.PfcpConnection
	ForwardPlaneStats *ebpf.UpfXdpActionStatistic
	Cfg               *config.Config
}

func NewHandler(bpfObjects *ebpf.BpfObjects, pfcpSrv *core.PfcpConnection, forwardPlaneStats *ebpf.UpfXdpActionStatistic, cfg *config.Config) *Handler {
	return &Handler{
		BpfObjects:        bpfObjects,
		PfcpSrv:           pfcpSrv,
		ForwardPlaneStats: forwardPlaneStats,
		Cfg:               cfg,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	eupfDocs.SwaggerInfo.BasePath = "/api/v1"
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(context *gin.Context) {
			response200(context, nil)
		})

		h.initDefaultRoutes(v1)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

func (h *Handler) InitMetricsRoute(cfg *config.Config) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger(), gin.ErrorLogger())
	gin.SetMode(gin.DebugMode)

	api := router.Group("/api/v1")

	{
		api.GET("/metrics", func() gin.HandlerFunc {
			return func(c *gin.Context) {
				promhttp.Handler().ServeHTTP(c.Writer, c.Request)
			}
		}())

	}

	return router
}

func (h *Handler) initDefaultRoutes(group *gin.RouterGroup) {

	group.GET("upf_pipeline", h.listUpfPipeline)
	group.GET("xdp_stats", h.displayXdpStatistics)
	group.GET("packet_stats", h.displayPacketStats)

	config := group.Group("config")
	{
		config.GET("", h.displayConfig)
		config.POST("", h.editConfig)
	}

	pdrMap := group.Group("uplink_pdr_map")
	{
		pdrMap.GET(":id", h.getUplinkPdrValue)
		pdrMap.PUT(":id", h.setUplinkPdrValue)
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
		//sessions.GET("", ListPfcpSessions(pfcpSrv))
		sessions.GET("", h.listPfcpSessionsFiltered)
	}
}
