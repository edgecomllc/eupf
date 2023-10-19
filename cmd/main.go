package main

import (
	"context"
	"github.com/cilium/ebpf/link"
	"github.com/edgecomllc/eupf/components/core"
	ebpf2 "github.com/edgecomllc/eupf/components/ebpf"
	"github.com/edgecomllc/eupf/config"
	"github.com/edgecomllc/eupf/internal/server"
	"github.com/edgecomllc/eupf/internal/transport/rest"
	"github.com/edgecomllc/eupf/pkg/logger"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:generate swag init --parseDependency

func init() {
	logger.SetLogger(logger.NewZeroLogger("debug"))
}

func main() {
	var cfg *config.Config
	var err error

	if cfg, err = config.New(); err != nil {
		logger.Fatalf("config err %+v", err)
	}

	var (
		_, cancel = context.WithCancel(context.Background())
		quit      = make(chan os.Signal, 1)
	)

	bpfObjects, err := ebpf2.New(cfg)
	if err != nil {
		logger.Fatalf("bpf init err: %+v", err)
	}
	defer bpfObjects.Close()
	bpfObjects.BuildPipeline()

	for _, ifaceName := range cfg.InterfaceName {
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			log.Fatalf("Lookup network iface %q: %s", ifaceName, err.Error())
		}

		// Attach the program.
		l, err := link.AttachXDP(link.XDPOptions{
			Program:   bpfObjects.UpfIpEntrypointFunc,
			Interface: iface.Index,
			Flags:     StringToXDPAttachMode(cfg.XDPAttachMode),
		})
		if err != nil {
			log.Fatalf("Could not attach XDP program: %s", err.Error())
		}
		defer l.Close()

		logger.Infof("Attached XDP program to iface %q (index %d)", iface.Name, iface.Index)
	}

	pfcpConn, err := core.New(cfg, bpfObjects)
	if err != nil {
		log.Fatalf("Could not create PFCP connection: %s", err.Error())
	}
	go pfcpConn.Run()
	defer pfcpConn.Close()

	forwardPlaneStats := ebpf2.NewUpfXdpActionStatistic(bpfObjects)

	h := rest.NewHandler(bpfObjects, pfcpConn, forwardPlaneStats, cfg)

	engine := h.InitRoutes()

	// Start api server
	srv := server.New(cfg.ApiAddress, engine)
	go func() {
		if err = srv.Run(); err != nil {
			logger.Fatalf("Could not start api server: %s", err.Error())
		}
	}()

	//core.RegisterMetrics(forwardPlaneStats, pfcpConn)
	//go func() {
	//	if err := core.StartMetrics(cfg.MetricsAddress); err != nil {
	//		log.Fatalf("Could not start metrics server: %s", err.Error())
	//	}
	//}()

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit

	ctxSrvStop, cancelSrvStop := context.WithTimeout(context.Background(), 5*time.Second)

	if err = srv.Stop(ctxSrvStop); err != nil {
		logger.Fatale(err)
	}

	cancelSrvStop()
	cancel()

	logger.Infof("admin shutdown signal %v", sig)
}

func StringToXDPAttachMode(Mode string) link.XDPAttachFlags {
	switch Mode {
	case "generic":
		return link.XDPGenericMode
	case "native":
		return link.XDPDriverMode
	case "offload":
		return link.XDPOffloadMode
	default:
		return link.XDPGenericMode
	}
}
