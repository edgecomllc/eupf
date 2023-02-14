package main

import (
	"flag"
	"log"
	"net"
	"time"

	"github.com/cilium/ebpf/link"
)

var ifaceName = flag.String("iface", "lo", "Interface to bind XDP program to")
var apiAddr = flag.String("aaddr", ":8080", "Address to bind api server to")

func main() {
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

	// Print the contents of the BPF hash map (source IP address -> packet count).
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s, err := FormatMapContents(bpfObjects.upf_xdpObjects.UpfPipeline)
		if err != nil {
			log.Printf("Error reading map: %s", err)
			continue
		}
		log.Printf("Pipeline map contents:\n%s", s)
	}
}
