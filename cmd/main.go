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
				if conn == nil {
					continue
				}
				totalAssocs += len(conn.NodeAssociations)
			}

			if totalAssocs == 0 {
				log.Info().Msg("All associations released")
				return
			}
		}
	}
}

// safeReleaseConn sends an association release request and deletes all
// associations on the connection. It is a no-op when conn is nil, which
// happens when the corresponding Sxa/Sxb connection was not configured.
func safeReleaseConn(conn *core.PfcpConnection) {
	if conn == nil {
		return
	}
	conn.SendAssociationReleaseRequest()
	conn.DeleteAllAssociations()
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

	// Select PFCP profile based on the pfcp_profile config key.
	var profile core.PfcpProfile
	switch config.Conf.PfcpProfile {
	case "huawei":
		profile = core.HuaweiProfile{
			S1UAddress:  config.Conf.S1UAddress,
			S5S8Address: config.Conf.S5S8Address,
			PAAddress:   config.Conf.PAAddress,
		}
	default:
		profile = core.DefaultProfile{}
	}

	type connSpec struct {
		label       string
		address     string
		nodeId      string
		n3Ip        string
		n9Ip        string
		sdfNotifyC  <-chan ebpf.SdfFlowNotification
		remoteNodes []string
		connectorFn func(string) (core.AssociationConnector, error)
	}

	startConn := func(s connSpec) *core.PfcpConnection {
		conn, err := core.NewPfcpConnection(s.address, s.nodeId, s.n3Ip, s.n9Ip, bpfObjects, resourceManager, dumper, s.sdfNotifyC, profile)
		if err != nil {
			log.Fatal().Msgf("Could not create %s connection: %s", s.label, err)
		}
		connectors := make([]core.AssociationConnector, 0, len(s.remoteNodes))
		for _, rn := range s.remoteNodes {
			c, err := s.connectorFn(rn)
			if err != nil {
				log.Warn().Msgf("failed to create %s association connector: %v", s.label, err)
				continue
			}
			connectors = append(connectors, c)
		}
		conn.SetRemoteNodes(connectors)
		go conn.Run()
		return conn
	}

	// Create PFCP (N4) connection
	pfcpConn := startConn(connSpec{
		label:       "PFCP",
		address:     config.Conf.PfcpAddress,
		nodeId:      config.Conf.PfcpNodeId,
		n3Ip:        config.Conf.N3AdvertisedAddress,
		n9Ip:        config.Conf.N9AdvertisedAddress,
		sdfNotifyC:  nil,
		remoteNodes: config.Conf.PfcpRemoteNode,
		connectorFn: profile.N4Connector,
	})
	defer pfcpConn.Close()

	// Create Sxa connection (optional — only when sxa_address is configured).
	// Without an Sxa peer the S1U/S5S8 interfaces are unused, so we skip the
	// connection entirely rather than binding a UDP socket to an empty address.
	var sxaConn *core.PfcpConnection
	if config.Conf.SxaLocalAddress != "" {
		sxaConn = startConn(connSpec{
			label:       "Sxa",
			address:     config.Conf.SxaLocalAddress,
			nodeId:      config.Conf.SxaLocalNodeId,
			n3Ip:        config.Conf.S1UAddress,
			n9Ip:        config.Conf.S5S8Address,
			sdfNotifyC:  nil,
			remoteNodes: config.Conf.SxaRemoteNode,
			connectorFn: profile.SxaConnector,
		})
		defer sxaConn.Close()
	} else {
		log.Info().Msg("Sxa connection skipped (sxa_address not configured)")
	}

	// Create Sxb connection (optional — only when sxb_address is configured).
	var sxbConn *core.PfcpConnection
	if config.Conf.SxbLocalAddress != "" {
		sxbConn = startConn(connSpec{
			label:       "Sxb",
			address:     config.Conf.SxbLocalAddress,
			nodeId:      config.Conf.SxbLocalNodeId,
			n3Ip:        config.Conf.PAAddress,
			n9Ip:        config.Conf.PAAddress,
			sdfNotifyC:  sdfNotifier.GetNotificationChannel(),
			remoteNodes: config.Conf.SxbRemoteNode,
			connectorFn: profile.SxbConnector,
		})
		defer sxbConn.Close()
	} else {
		log.Info().Msg("Sxb connection skipped (sxb_address not configured)")
	}

	gtpPathManager := core.NewGtpPathManager(config.Conf.N3Address+core.GTPPortStr, time.Duration(config.Conf.GtpEchoInterval)*time.Second)
	for _, peer := range config.Conf.GtpPeer {
		gtpPathManager.AddGtpPath(peer)
	}
	gtpPathManager.Run()
	defer gtpPathManager.Stop()

	ForwardPlaneStats := ebpf.UpfXdpActionStatistic{
		BpfObjects: bpfObjects,
	}

	// Build the PFCP connection map exposed to the API. Sxa/Sxb entries are
	// only present when the corresponding connection was actually created; the
	// REST handlers in cmd/api/rest/config.go already guard with `exists`, so
	// omitting a key is the signal that the connection is unavailable.
	pfcpMap := map[string]*core.PfcpConnection{
		core.N4PFCPKeyName: pfcpConn,
	}
	if sxaConn != nil {
		pfcpMap[core.SxaPFCPKeyName] = sxaConn
	}
	if sxbConn != nil {
		pfcpMap[core.SxbPFCPKeyName] = sxbConn
	}

	h := rest.NewApiHandler(
		bpfObjects,
		pfcpMap,
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
		// Request release and clear associations. safeReleaseConn is a no-op
		// for Sxa/Sxb connections that were never created (nil).
		safeReleaseConn(pfcpConn)
		safeReleaseConn(sxaConn)
		safeReleaseConn(sxbConn)

		// Wait until every existing connection reports zero associations.
		// waitForAllAssocReleases skips nil entries internally.
		waitForAllAssocReleases(ctx, pfcpConn, sxaConn, sxbConn)
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
