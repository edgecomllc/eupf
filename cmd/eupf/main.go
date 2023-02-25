package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/wmnsk/go-pfcp/message"
)

var ifaceName = flag.String("iface", "lo", "Interface to bind XDP program to")
var apiAddr = flag.String("aaddr", ":8080", "Address to bind api server to")
var pfcpAddr = flag.String("paddr", ":8805", "Address to bind PFCP server to")
var pfcpNodeId = flag.String("nodeid", "pfcp.somecore.internal", "PFCP Server Node ID")

func main() {
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	flag.Parse()

	if err := IncreaseResourceLimits(); err != nil {
		log.Fatalf("Can't increase resourse limits: %s", err)
	}

	bpfObjects := &BpfObjects{}
	if err := bpfObjects.Load(); err != nil {
		log.Fatalf("Loading bpf objects failed: %s", err)
	}

	defer bpfObjects.Close()

	bpfObjects.buildPipeline()

	iface, err := net.InterfaceByName(*ifaceName)
	if err != nil {
		log.Fatalf("Lookup network iface %q: %s", *ifaceName, err)
	}

	// Attach the program.
	l, err := link.AttachXDP(link.XDPOptions{
		Program:   bpfObjects.UpfIpEntrypointFunc,
		Interface: iface.Index,
	})
	if err != nil {
		log.Fatalf("Could not attach XDP program: %s", err)
	}
	defer l.Close()

	log.Printf("Attached XDP program to iface %q (index %d)", iface.Name, iface.Index)
	log.Printf("Press Ctrl-C to exit and remove the program")

	// Start api server
	api := CreateApiServer(bpfObjects)
	go api.Run(*apiAddr)

	// Create PFCP connection
	var pfcpHandlers PfcpHanderMap = PfcpHanderMap{
		message.MsgTypeHeartbeatRequest:        handlePfcpHeartbeatRequest,
		message.MsgTypeAssociationSetupRequest: handlePfcpAssociationSetupRequest,
		message.MsgTypeSessionEstablishmentRequest: handlePfcpSessionEstablishmentRequest,
	}

	pfcp_conn, err := CreatePfcpConnection(*pfcpAddr, pfcpHandlers, *pfcpNodeId)
	if err != nil {
		log.Printf("Could not create PFCP connection: %s", err)
	}
	go pfcp_conn.Run()
	defer pfcp_conn.Close()

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
