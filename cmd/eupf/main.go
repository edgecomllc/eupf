package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/edgecomllc/eupf/cmd/eupf/config"
	"github.com/wmnsk/go-pfcp/message"
)

func main() {
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	config.Init()

	if err := IncreaseResourceLimits(); err != nil {
		log.Fatalf("Can't increase resource limits: %s", err.Error())
	}

	bpfObjects := &BpfObjects{}
	if err := bpfObjects.Load(); err != nil {
		log.Fatalf("Loading bpf objects failed: %s", err.Error())
	}

	if err := bpfObjects.ResizeEbpfMapsFromConfig(config.Conf.QerMapSize, config.Conf.FarMapSize, config.Conf.PdrMapSize); err != nil {
		log.Fatalf("Failed to set ebpf map sizes: %s", err)
	}

	defer bpfObjects.Close()

	bpfObjects.buildPipeline()

	for _, ifaceName := range config.Conf.InterfaceName {
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			log.Fatalf("Lookup network iface %q: %s", ifaceName, err.Error())
		}

		// Attach the program.
		l, err := link.AttachXDP(link.XDPOptions{
			Program:   bpfObjects.UpfIpEntrypointFunc,
			Interface: iface.Index,
			Flags:     StringToXDPAttachMode(config.Conf.XDPAttachMode),
		})
		if err != nil {
			log.Fatalf("Could not attach XDP program: %s", err.Error())
		}
		defer l.Close()

		log.Printf("Attached XDP program to iface %q (index %d)", iface.Name, iface.Index)
	}

	// Create PFCP connection
	var pfcpHandlers = PfcpHandlerMap{
		message.MsgTypeHeartbeatRequest:            handlePfcpHeartbeatRequest,
		message.MsgTypeAssociationSetupRequest:     handlePfcpAssociationSetupRequest,
		message.MsgTypeSessionEstablishmentRequest: handlePfcpSessionEstablishmentRequest,
		message.MsgTypeSessionDeletionRequest:      handlePfcpSessionDeletionRequest,
		message.MsgTypeSessionModificationRequest:  handlePfcpSessionModificationRequest,
	}

	pfcpConn, err := CreatePfcpConnection(config.Conf.PfcpAddress, pfcpHandlers, config.Conf.PfcpNodeId, config.Conf.N3Address, bpfObjects)
	if err != nil {
		log.Fatalf("Could not create PFCP connection: %s", err.Error())
	}
	go pfcpConn.Run()
	defer pfcpConn.Close()

	ForwardPlaneStats := UpfXdpActionStatistic{
		bpfObjects: bpfObjects,
	}

	// Start api server
	api := CreateApiServer(bpfObjects, pfcpConn, ForwardPlaneStats)
	go func() {
		if err := api.Run(config.Conf.ApiAddress); err != nil {
			log.Fatalf("Could not start api server: %s", err.Error())
		}
	}()

	RegisterMetrics(ForwardPlaneStats)
	go func() {
		if err := StartMetrics(config.Conf.MetricsAddress); err != nil {
			log.Fatalf("Could not start metrics server: %s", err.Error())
		}
	}()

	// Print the contents of the BPF hash map (source IP address -> packet count).
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// s, err := FormatMapContents(bpfObjects.upf_xdpObjects.UpfPipeline)
			// if err != nil {
			// 	log.Printf("Error reading map: %s", err)
			// 	continue
			// }
			// log.Printf("Pipeline map contents:\n%s", s)
		case <-stopper:
			log.Println("Received signal, exiting program..")
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
