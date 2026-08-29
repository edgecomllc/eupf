package main

import (
	"context"
	"encoding/binary"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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
)

//go:generate swag init --parseDependency --parseInternal --parseDepth 1 -g api/rest/handler.go

func waitForAllAssocReleases(ctx context.Context, conns ...*core.PfcpConnection) {
	const (
		pfcpAssocReleaseTimeout      = time.Second * 10
		pfcpAssocReleaseCheckTimeout = time.Millisecond * 100
	)

	ctx, cancel := context.WithTimeout(ctx, pfcpAssocReleaseTimeout)
	defer cancel()

	tick := time.NewTicker(pfcpAssocReleaseCheckTimeout)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Warn().Msg("Association release timeout")
			return
		case <-tick.C:
			totalAssocs := 0
			for _, conn := range conns {
				totalAssocs += len(conn.NodeAssociations)
			}

			if totalAssocs == 0 {
				log.Info().Msg("All associations released")
				return
			}
		}
	}
}

func main() {
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)
	resetAssociationChan := make(chan struct{})

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
	bpfObjects.SetPdrMapSize(config.Conf.PdrMapSize)
	bpfObjects.SetFarMapSize(config.Conf.FarMapSize)
	bpfObjects.SetQerMapSize(config.Conf.QerMapSize)
	bpfObjects.SetUrrMapSize(config.Conf.UrrMapSize)
	if err := bpfObjects.Load(); err != nil {
		log.Fatal().Msgf("Loading bpf objects failed: %s", err.Error())
	}

	defer bpfObjects.Close()

	var err error
	dumper, err := utils.NewPacketDumper(
		"dumps/",
		config.Conf.TraceMaxDumpFiles,
		config.Conf.TraceMaxDumpSize,
		config.Conf.TraceMaxDumpPackets,
	)
	if err != nil {
		log.Error().Msgf("Can't start dumper: %s", err.Error())
	} else {
		defer dumper.Close(false)
		go dumper.ReadTraceMap(bpfObjects.TraceMap)
		go dumper.Write()
	}

	entrypointConfig := ebpf.IpEntrypointDataplaneConfig{
		N3Ipv4Address: binary.LittleEndian.Uint32(net.ParseIP(config.Conf.N3Address).To4()),
		N9Ipv4Address: binary.LittleEndian.Uint32(net.ParseIP(config.Conf.N9Address).To4()),
		TraceIn:       0,
		TraceOut:      0,
		TraceBlocked:  0,
		Ip6RaSupport:  0,
	}

	if config.Conf.TraceIn {
		entrypointConfig.TraceIn = 1
	}

	if config.Conf.TraceOut {
		entrypointConfig.TraceOut = 1
	}

	if config.Conf.TraceBlocked {
		entrypointConfig.TraceBlocked = 1
	}

	if config.Conf.IP6RaSupport {
		ip6Prefix, ip6Net, _ := net.ParseCIDR(config.Conf.Ip6RaPrefix)
		ip6PrefixLength, _ := ip6Net.Mask.Size()

		entrypointConfig.Ip6RaSupport = 1
		copy(entrypointConfig.Ip6RaPrefix[:], ip6Prefix.To16())
		entrypointConfig.Ip6RaPrefixLength = uint16(ip6PrefixLength)
	}

	if err := bpfObjects.GlobalConfig.Set(entrypointConfig); err != nil {
		log.Fatal().Err(err).Msgf("can't set dataplane global config")
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
		config.Conf.N3AdvertisedAddress,
		config.Conf.N9AdvertisedAddress,
		bpfObjects,
		resourceManager,
		dumper,
		nil,
		core.DefaultProfile{},
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
		resourceManager,
		dumper,
		nil,
		core.DefaultProfile{},
	)

	if err != nil {
		log.Fatal().Msgf("Could not create Sxa connection: %s", err.Error())
	}
	sxaRemoteNodes := []core.AssociationConnector{}
	for _, remoteNode := range config.Conf.SxaRemoteNode {
		if config.Conf.HuaweiSupport {
			connector, err := core.NewSxaAssociationConnector(remoteNode, config.Conf.S1UAddress, config.Conf.S5S8Address)
			if err != nil {
				log.Warn().Msgf("failed to create sxa association connector: %v", err)
				continue
			}
			sxaRemoteNodes = append(sxaRemoteNodes, connector)
		} else {
			connector, err := core.NewDefaultAssociationConnector(remoteNode)
			if err != nil {
				log.Warn().Msgf("failed to create sxa association connector: %v", err)
				continue
			}
			sxaRemoteNodes = append(sxaRemoteNodes, connector)
		}
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
		resourceManager,
		dumper,
		sdfNotifier.GetNotificationChannel(),
		core.DefaultProfile{},
	)

	if err != nil {
		log.Fatal().Msgf("Could not create Sxb connection: %s", err.Error())
	}
	sxbRemoteNodes := []core.AssociationConnector{}
	for _, remoteNode := range config.Conf.SxbRemoteNode {
		if config.Conf.HuaweiSupport {
			connector, err := core.NewSxbAssociationConnector(remoteNode, config.Conf.PAAddress)
			if err != nil {
				log.Warn().Msgf("failed to create sxb association connector: %v", err)
				continue
			}
			sxbRemoteNodes = append(sxbRemoteNodes, connector)
		} else {
			connector, err := core.NewDefaultAssociationConnector(remoteNode)
			if err != nil {
				log.Warn().Msgf("failed to create sxb association connector: %v", err)
				continue
			}
			sxbRemoteNodes = append(sxbRemoteNodes, connector)
		}
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
		resetAssociationChan,
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assocReleaseFunc := func() {
		pfcpConn.SendAssociationReleaseRequest()
		sxaConn.SendAssociationReleaseRequest()
		sxbConn.SendAssociationReleaseRequest()

		waitForAllAssocReleases(ctx, pfcpConn, sxaConn, sxbConn)

		pfcpConn.DeleteAllAssociations()
		sxaConn.DeleteAllAssociations()
		sxbConn.DeleteAllAssociations()
	}

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
			log.Info().Msg("OS signal received. Initiating graceful shutdown...")

			assocReleaseFunc()

			gtpPathManager.Stop()

			if err := apiSrv.Stop(ctx); err != nil {
				log.Error().Err(err).Msg("Error while stopping API server")
			}
			if err := metricsSrv.Stop(ctx); err != nil {
				log.Error().Err(err).Msg("Error while stopping metrics server")
			}

			log.Info().Msg("Service has been shut down successfully.")

			return
		case <-resetAssociationChan:
			assocReleaseFunc()
		}
	}
}
