package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/wmnsk/go-pfcp/message"
)

func main() {
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	if LoadConfig() != nil {
		log.Fatalf("Unable to load config")
	}

	if err := IncreaseResourceLimits(); err != nil {
		log.Fatalf("Can't increase resourse limits: %s", err)
	}

	bpfObjects := &BpfObjects{}
	if err := bpfObjects.Load(); err != nil {
		log.Fatalf("Loading bpf objects failed: %s", err)
	}

	defer bpfObjects.Close()

	bpfObjects.buildPipeline()

	for _, ifaceName := range config.InterfaceName {
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			log.Fatalf("Lookup network iface %q: %s", ifaceName, err)
		}

		// Attach the program.
		l, err := link.AttachXDP(link.XDPOptions{
			Program:   bpfObjects.UpfIpEntrypointFunc,
			Interface: iface.Index,
			Flags:     StringToXDPAttachMode(config.XDPAttachMode),
		})
		if err != nil {
			log.Fatalf("Could not attach XDP program: %s", err)
		}
		defer l.Close()

		log.Printf("Attached XDP program to iface %q (index %d)", iface.Name, iface.Index)
	}

	// Create PFCP connection
	var pfcpHandlers PfcpHanderMap = PfcpHanderMap{
		message.MsgTypeHeartbeatRequest:            handlePfcpHeartbeatRequest,
		message.MsgTypeAssociationSetupRequest:     handlePfcpAssociationSetupRequest,
		message.MsgTypeSessionEstablishmentRequest: handlePfcpSessionEstablishmentRequest,
		message.MsgTypeSessionDeletionRequest:      handlePfcpSessionDeletionRequest,
		message.MsgTypeSessionModificationRequest:  handlePfcpSessionModificationRequest,
	}

	pfcp_conn, err := CreatePfcpConnection(config.PfcpAddress, pfcpHandlers, config.PfcpNodeId, config.N3Address, bpfObjects)

	if err != nil {
		log.Printf("Could not create PFCP connection: %s", err)
	}
	go pfcp_conn.Run()
	defer pfcp_conn.Close()

	ForwardPlaneStats := UpfXdpActionStatistic{
		bpfObjects: bpfObjects,
	}

	// Start api server
	api := CreateApiServer(bpfObjects, pfcp_conn, ForwardPlaneStats)
	go api.Run(config.ApiAddress)

	RegisterMetrics(ForwardPlaneStats)
	go StartMetrics(config.MetricsAddress)
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
