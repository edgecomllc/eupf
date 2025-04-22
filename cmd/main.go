package main

import (
	"fmt"
	"github.com/cilium/ebpf/link"
	"github.com/edgecomllc/eupf/cmd/api/rest"
	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/edgecomllc/eupf/cmd/core"
	"github.com/edgecomllc/eupf/cmd/core/service"
	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/edgecomllc/eupf/cmd/server"
	"github.com/edgecomllc/eupf/cmd/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:generate swag init --parseDependency --parseInternal --parseDepth 1 -g api/rest/handler.go

func main() {
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	config.Init()

	// Warning: inefficient log writing.
	// As zerolog docs says: "Pretty logging on the console is made possible using the provided (but inefficient) zerolog.ConsoleWriter."
	core.InitLogger()
	core.SetLoggerCaller(config.Conf.LoggingCaller)
	if err := core.SetLoggerLevel(config.Conf.LoggingLevel); err != nil {
		log.Error().Msgf("Logger configuring error: %s. Using '%s' level", err.Error(), zerolog.GlobalLevel().String())
	}

	if err := ebpf.IncreaseResourceLimits(); err != nil {
		log.Fatal().Msgf("Can't increase resource limits: %s", err.Error())
	}

	bpfObjects := ebpf.NewBpfObjects()
	if err := bpfObjects.Load(); err != nil {
		log.Fatal().Msgf("Loading bpf objects failed: %s", err.Error())
	}

	defer bpfObjects.Close()
	if config.Conf.EbpfMapResize {
		if err := bpfObjects.ResizeAllMaps(config.Conf.QerMapSize, config.Conf.FarMapSize, config.Conf.PdrMapSize); err != nil {
			log.Fatal().Msgf("Failed to set ebpf map sizes: %s", err)
		}
	}

	var err error
	traceFileName := fmt.Sprintf("dumps/%s-trace.pcap", time.Now().Format(time.RFC3339))
	dumper, err := utils.NewPacketDumper(traceFileName)
	if err != nil {
		log.Error().Msgf("Can't start dumper: %s", err.Error())
	} else {
		defer dumper.Close(false)
		go dumper.ReadTraceMap(bpfObjects.TraceMap)
		go dumper.Write()
	}

	sdfNotifier, _ := ebpf.NewSdfNotifyListener(bpfObjects.SdfNotifyMap)
	defer sdfNotifier.Close()
	go sdfNotifier.ReadEventLoop()

	links := make([]link.Link, 0, len(config.Conf.InterfaceName))
	for _, ifaceName := range config.Conf.InterfaceName {
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			log.Fatal().Msgf("Lookup network iface %q: %s", ifaceName, err.Error())
		}

		// Attach the program.
		l, err := link.AttachXDP(link.XDPOptions{
			Program:   bpfObjects.UpfIpEntrypointFunc,
			Interface: iface.Index,
			Flags:     core.StringToXDPAttachMode(config.Conf.XDPAttachMode),
		})
		if err != nil {
			log.Fatal().Msgf("Could not attach XDP program: %s", err.Error())
		}

		links = append(links, l)

		log.Info().Msgf("Attached XDP program to iface %q (index %d)", iface.Name, iface.Index)
	}

	defer func() {
		for _, l := range links {
			err := l.Close()
			if err != nil {
				log.Error().Msgf("error closing link: %s", err.Error())
			}
		}
	}()

	log.Info().Msgf("Initialize resources: UEIP pool (CIDR: \"%s\"), TEID pool (size: %d)", config.Conf.UEIPPool, config.Conf.FTEIDPool)
	resourceManager, err := service.NewResourceManager(config.Conf.UEIPPool, config.Conf.FTEIDPool)
	if err != nil {
		log.Error().Msgf("failed to create ResourceManager - err: %v", err)
	}

	// Create PFCP connection
	pfcpConn, err := core.NewPfcpConnection(
		config.Conf.PfcpAddress,
		config.Conf.PfcpNodeId,
		config.Conf.N3Address,
		config.Conf.N9Address,
		bpfObjects,
		resourceManager,
		dumper,
		nil,
	)

	if err != nil {
		log.Fatal().Msgf("Could not create PFCP connection: %s", err.Error())
	}
	remoteNodes := []core.AssociationConnector{}
	for _, remoteNode := range config.Conf.PfcpRemoteNode {
		connector, err := core.NewDefaultAssociationConnector(remoteNode)
		if err != nil {
			log.Warn().Msgf("failed to create default association connector: %v", err)
			continue
		}
		remoteNodes = append(remoteNodes, connector)
	}
	pfcpConn.SetRemoteNodes(remoteNodes)
	go pfcpConn.Run()
	defer pfcpConn.Close()

	// Create Sxa connection
	sxaConn, err := core.NewPfcpConnection(
		config.Conf.SxaLocalAddress,
		config.Conf.SxaLocalNodeId,
		config.Conf.S1UAddress,
		config.Conf.S5S8Address,
		bpfObjects,
		nil,
		dumper,
		nil,
	)

	if err != nil {
		log.Fatal().Msgf("Could not create Sxa connection: %s", err.Error())
	}
	sxaRemoteNodes := []core.AssociationConnector{}
	for _, remoteNode := range config.Conf.SxaRemoteNode {
		connector, err := core.NewSxaAssociationConnector(remoteNode, config.Conf.S1UAddress, config.Conf.S5S8Address)
		if err != nil {
			log.Warn().Msgf("failed to create sxa association connector: %v", err)
			continue
		}
		sxaRemoteNodes = append(remoteNodes, connector)
	}
	sxaConn.SetRemoteNodes(sxaRemoteNodes)
	go sxaConn.Run()
	defer sxaConn.Close()

	// Create Sxb connection
	sxbConn, err := core.NewPfcpConnection(
		config.Conf.SxbLocalAddress,
		config.Conf.SxbLocalNodeId,
		config.Conf.PAAddress,
		config.Conf.PAAddress,
		bpfObjects,
		nil,
		dumper,
		sdfNotifier.GetNotificationChannel(),
	)

	if err != nil {
		log.Fatal().Msgf("Could not create Sxb connection: %s", err.Error())
	}
	sxbRemoteNodes := []core.AssociationConnector{}
	for _, remoteNode := range config.Conf.SxbRemoteNode {
		connector, err := core.NewSxbAssociationConnector(remoteNode, config.Conf.PAAddress)
		if err != nil {
			log.Warn().Msgf("failed to create sxb association connector: %v", err)
			continue
		}
		sxbRemoteNodes = append(remoteNodes, connector)
	}
	sxbConn.SetRemoteNodes(sxbRemoteNodes)
	go sxbConn.Run()
	defer sxbConn.Close()

	gtpPathManager := core.NewGtpPathManager(config.Conf.N3Address+core.GTPPortStr, time.Duration(config.Conf.GtpEchoInterval)*time.Second)
	for _, peer := range config.Conf.GtpPeer {
		gtpPathManager.AddGtpPath(peer)
	}
	gtpPathManager.Run()
	defer gtpPathManager.Stop()

	ForwardPlaneStats := ebpf.UpfXdpActionStatistic{
		BpfObjects: bpfObjects,
	}

	h := rest.NewApiHandler(
		bpfObjects,
		map[string]*core.PfcpConnection{
			core.N4PFCPKeyName:  pfcpConn,
			core.SxaPFCPKeyName: sxaConn,
			core.SxbPFCPKeyName: sxbConn,
		},
		&ForwardPlaneStats,
		&config.Conf,
		&links,
		gtpPathManager,
	)

	engine := h.InitRoutes()
	metricsEngine := h.InitMetricsRoute()

	apiSrv := server.New(config.Conf.ApiAddress, engine)
	metricsSrv := server.New(config.Conf.MetricsAddress, metricsEngine)

	// Start api servers
	go func() {
		if err := apiSrv.Run(); err != nil {
			log.Fatal().Msgf("Could not start api server: %s", err.Error())
		}
	}()

	// Start metrics servers
	go func() {
		if err := metricsSrv.Run(); err != nil {
			log.Fatal().Msgf("Could not start metrics server: %s", err.Error())
		}
	}()

	// Print the contents of the BPF hash map (source IP address -> packet count).
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// s, err := FormatMapContents(bpfObjects.UpfXdpObjects.UpfPipeline)
			// if err != nil {
			// 	log.Printf("Error reading map: %s", err)
			// 	continue
			// }
			// log.Printf("Pipeline map contents:\n%s", s)
		case <-stopper:
			log.Info().Msgf("Received signal, exiting program..")
			return
		}
	}
}
