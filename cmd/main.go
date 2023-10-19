package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/edgecomllc/eupf/cmd/core"
	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/fsnotify/fsnotify"

	"github.com/cilium/ebpf/link"
	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/message"
)

//go:generate swag init --parseDependency

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
	if err := bpfObjects.Load(); err != nil {
		log.Fatal().Msgf("Loading bpf objects failed: %s", err.Error())
	}

	if config.Conf.EbpfMapResize {
		if err := bpfObjects.ResizeAllMaps(config.Conf.QerMapSize, config.Conf.FarMapSize, config.Conf.PdrMapSize); err != nil {
			log.Fatal().Msgf("Failed to set ebpf map sizes: %s", err)
		}
	}

	defer bpfObjects.Close()

	bpfObjects.BuildPipeline()

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

	// Create PFCP connection
	var pfcpHandlers = core.PfcpHandlerMap{
		message.MsgTypeHeartbeatRequest:            core.HandlePfcpHeartbeatRequest,
		message.MsgTypeHeartbeatResponse:           core.HandlePfcpHeartbeatResponse,
		message.MsgTypeAssociationSetupRequest:     core.HandlePfcpAssociationSetupRequest,
		message.MsgTypeSessionEstablishmentRequest: core.HandlePfcpSessionEstablishmentRequest,
		message.MsgTypeSessionDeletionRequest:      core.HandlePfcpSessionDeletionRequest,
		message.MsgTypeSessionModificationRequest:  core.HandlePfcpSessionModificationRequest,
	}

	pfcpConn, err := core.CreatePfcpConnection(config.Conf.PfcpAddress, pfcpHandlers, config.Conf.PfcpNodeId, config.Conf.N3Address, bpfObjects)
	if err != nil {
		log.Fatal().Msgf("Could not create PFCP connection: %s", err.Error())
	}
	go pfcpConn.Run()
	defer pfcpConn.Close()

	ForwardPlaneStats := ebpf.UpfXdpActionStatistic{
		BpfObjects: bpfObjects,
	}

	// Start api server
	api := core.CreateApiServer(bpfObjects, pfcpConn, ForwardPlaneStats)
	go func() {
		if err := api.Run(config.Conf.ApiAddress); err != nil {
			log.Fatal().Msgf("Could not start api server: %s", err.Error())
		}
	}()

	core.RegisterMetrics(ForwardPlaneStats, pfcpConn)
	go func() {
		if err := core.StartMetrics(config.Conf.MetricsAddress); err != nil {
			log.Fatal().Msgf("Could not start metrics server: %s", err.Error())
		}
	}()

	// Handle config file change.
	config.WatchConfig(OnConfigFileChange)

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

func OnConfigFileChange(e fsnotify.Event) {
	if e.Has(fsnotify.Create) || e.Has(fsnotify.Write) {
		if conf, err := config.ReadConfig(); err == nil {
			if err := core.SetConfig(conf); err != nil {
				log.Error().Msgf("Error during config file change handling: %s", err.Error())
			}
		}
	}
}
