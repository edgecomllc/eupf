package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/gin-gonic/gin"
)

var ifaceName = flag.String("iface", "lo", "Interface to bind XDP program to")
var webAddr = flag.String("waddr", ":8080", "Address to bind web server to")

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

	r := gin.Default()
	
	r.GET("/upf_pipeline", func(c *gin.Context) {
		elements, err := ListMapProgArrayContents(bpfObjects.upf_xdpObjects.UpfPipeline)
		if err != nil {
			log.Printf("Error reading map: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, elements)
	})

	// Start web server
	go r.Run(*webAddr)

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
