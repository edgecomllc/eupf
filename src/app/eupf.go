package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"golang.org/x/sys/unix"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf ip_entrypoint 	../xdp/ip_entrypoint.c -- -I.. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf gtp_entrypoint 	../xdp/gtp_entrypoint.c -- -I.. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf qer_program 		../xdp/qer_program.c -- -I.. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf far_program 		../xdp/far_program.c -- -I.. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf upf_xdp 			../xdp/upf_program.c -- -I.. -O2 -Wall

type BpfObjects struct {
	upf_xdpObjects
	far_programObjects
	qer_programObjects
	ip_entrypointObjects
	gtp_entrypointObjects
}

func (o *BpfObjects) Load() error {

	pinPath := "/sys/fs/bpf/upf_pipeline"
	if err := os.MkdirAll(pinPath, os.ModePerm); err != nil {
		log.Fatalf("failed to create bpf fs subpath: %+v", err)
		return err
	}

	collectionOptions := ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			// Pin the map to the BPF filesystem and configure the
			// library to automatically re-write it in the BPF
			// program so it can be re-used if it already exists or
			// create it if not
			PinPath: pinPath,
		},
	}

	return LoadAllObjects(&collectionOptions,
		Loader{loadUpf_xdpObjects, &o.upf_xdpObjects},
		Loader{loadFar_programObjects, &o.far_programObjects},
		Loader{loadQer_programObjects, &o.qer_programObjects},
		Loader{loadIp_entrypointObjects, &o.ip_entrypointObjects},
		Loader{loadGtp_entrypointObjects, &o.gtp_entrypointObjects})
}

func (o *BpfObjects) Close() error {
	return CloseAllObjects(
		&o.upf_xdpObjects,
		&o.far_programObjects,
		&o.qer_programObjects,
		&o.ip_entrypointObjects,
		&o.gtp_entrypointObjects,
	)
}

func (bpfObjects *BpfObjects) buildPipeline() {
	upfPipeline := bpfObjects.upf_xdpObjects.UpfPipeline
	upfMainProgram := bpfObjects.UpfFunc
	farProgram := bpfObjects.UpfFarProgramFunc
	qerProgram := bpfObjects.UpfQerProgramFunc

	if err := upfPipeline.Put(uint32(0), upfMainProgram); err != nil {
		panic(err)
	}

	if err := upfPipeline.Put(uint32(1), farProgram); err != nil {
		panic(err)
	}

	if err := upfPipeline.Put(uint32(2), qerProgram); err != nil {
		panic(err)
	}

}

type LoaderFunc func(obj interface{}, opts *ebpf.CollectionOptions) error
type Loader struct {
	LoaderFunc
	object interface{}
}

func LoadAllObjects(opts *ebpf.CollectionOptions, loaders ...Loader) error {
	for _, loader := range loaders {
		if err := loader.LoaderFunc(loader.object, opts); err != nil {
			return err
		}
	}
	return nil
}

func CloseAllObjects(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

var ifaceName = flag.String("iface", "lo", "Interface to bind XDP program to")

func main() {
	flag.Parse()

	if err := increaseResourceLimits(); err != nil {
		panic(err)
	}

	bpfObjects := &BpfObjects{}
	if err := bpfObjects.Load(); err != nil {
		panic(err)
	}

	defer bpfObjects.Close()

	bpfObjects.buildPipeline()

	iface, err := net.InterfaceByName(*ifaceName)
	if err != nil {
		log.Fatalf("lookup network iface %q: %s", *ifaceName, err)
	}

	// Attach the program.
	l, err := link.AttachXDP(link.XDPOptions{
		Program:   bpfObjects.UpfIpEntrypointFunc,
		Interface: iface.Index,
	})
	if err != nil {
		log.Fatalf("could not attach XDP program: %s", err)
	}
	defer l.Close()

	log.Printf("Attached XDP program to iface %q (index %d)", iface.Name, iface.Index)
	log.Printf("Press Ctrl-C to exit and remove the program")

	// Print the contents of the BPF hash map (source IP address -> packet count).
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s, err := formatMapContents(bpfObjects.upf_xdpObjects.UpfPipeline)
		if err != nil {
			log.Printf("Error reading map: %s", err)
			continue
		}
		log.Printf("Pipeline map contents:\n%s", s)
	}
}

//increaseResourceLimits https://prototype-kernel.readthedocs.io/en/latest/bpf/troubleshooting.html#memory-ulimits
func increaseResourceLimits() error {
	return unix.Setrlimit(unix.RLIMIT_MEMLOCK, &unix.Rlimit{
		Cur: unix.RLIM_INFINITY,
		Max: unix.RLIM_INFINITY,
	})
}

func formatMapContents(m *ebpf.Map) (string, error) {
	var (
		sb  strings.Builder
		key []byte
		val uint32
	)
	iter := m.Iterate()
	for iter.Next(&key, &val) {
		programId := key
		programRef := val
		sb.WriteString(fmt.Sprintf("\t%d => %d\n", programId, programRef))
	}
	return sb.String(), iter.Err()
}
