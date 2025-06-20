package main

import (
	"encoding/binary"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/edgecomllc/eupf/cmd/core/service"

	"github.com/edgecomllc/eupf/cmd/api/rest"
	"github.com/edgecomllc/eupf/cmd/server"

	"github.com/edgecomllc/eupf/cmd/core"
	"github.com/edgecomllc/eupf/cmd/ebpf"

	"github.com/cilium/ebpf/link"
	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:generate swag init --parseDependency --parseInternal --parseDepth 1 -g api/rest/handler.go

func main() {
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	config.Init()

	// Warning: inefficient log writing.
	// As zerolog docs says: "Pretty logging on the console is made possible using the provided (but inefficient) zerolog.ConsoleWriter."
	core.InitLogger()
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

	n3AddressUint32 := binary.LittleEndian.Uint32(net.ParseIP(config.Conf.N3Address).To4())
	n9AddressUint32 := binary.LittleEndian.Uint32(net.ParseIP(config.Conf.N9Address).To4())

	entrypointConfig := ebpf.IpEntrypointDataplaneConfig{N3Ipv4Address: n3AddressUint32, N9Ipv4Address: n9AddressUint32}
	if err := bpfObjects.GlobalConfig.Set(entrypointConfig); err != nil {
		log.Fatal().Err(err).Msgf("can't set dataplane global config")
	}

	defer bpfObjects.Close()

	for _, ifaceName := range config.Conf.InterfaceName {
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			log.Fatal().Msgf("Lookup network iface %q: %s", ifaceName, err.Error())
		}

		// Attach the program.
		l, err := link.AttachXDP(link.XDPOptions{
			Program:   bpfObjects.UpfIpEntrypointFunc,
			Interface: iface.Index,
			Flags:     StringToXDPAttachMode(config.Conf.XDPAttachMode),
		})
		if err != nil {
			log.Fatal().Msgf("Could not attach XDP program: %s", err.Error())
		}
		defer l.Close()

		log.Info().Msgf("Attached XDP program to iface %q (index %d)", iface.Name, iface.Index)
	}

	log.Info().Msgf("Initialize resources: UEIP pool (CIDR: \"%s\"), TEID pool (size: %d)", config.Conf.UEIPPool, config.Conf.FTEIDPool)
	var err error
	resourceManager, err := service.NewResourceManager(config.Conf.UEIPPool, config.Conf.FTEIDPool)
	if err != nil {
		log.Error().Msgf("failed to create ResourceManager - err: %v", err)
	}

	// Create PFCP connection
	pfcpConn, err := core.NewPfcpConnection(config.Conf.PfcpAddress, config.Conf.PfcpNodeId,
		config.Conf.N3Address, config.Conf.N9Address,
		bpfObjects, resourceManager)
	if err != nil {
		log.Fatal().Msgf("Could not create PFCP connection: %s", err.Error())
	}

	remoteNodes := []core.AssociationConnector{}
	for _, remoteNode := range config.Conf.PfcpRemoteNode {
		remoteNodes = append(remoteNodes, core.NewDefaultAssociationConnector(remoteNode))
	}
	pfcpConn.SetRemoteNodes(remoteNodes)
	go pfcpConn.Run()
	defer pfcpConn.Close()

	ForwardPlaneStats := ebpf.UpfXdpActionStatistic{
		BpfObjects: bpfObjects,
	}

	h := rest.NewApiHandler(bpfObjects, pfcpConn, &ForwardPlaneStats, &config.Conf)

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

	gtpPathManager := core.NewGtpPathManager(config.Conf.N3Address+":2152", time.Duration(config.Conf.GtpEchoInterval)*time.Second)
	for _, peer := range config.Conf.GtpPeer {
		gtpPathManager.AddGtpPath(peer)
	}
	gtpPathManager.Run()
	defer gtpPathManager.Stop()

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
