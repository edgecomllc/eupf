package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf ip_entrypoint 	../xdp/ip_entrypoint.c -- -I.. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf gtp_entrypoint 	../xdp/gtp_entrypoint.c -- -I.. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf qer_program 		../xdp/qer_program.c -- -I.. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf far_program 		../xdp/far_program.c -- -I.. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf upf_xdp 			../xdp/upf_program.c -- -I.. -O2 -Wall

var iface = flag.String("iface", "", "Interface to bind XDP program to")

func main() {
	flag.Parse()

	if err := increaseResourceLimits(); err != nil {
		panic(err)
	}

	objs := upf_xdpObjects{}
	if err := loadUpf_xdpObjects(&objs, nil); err != nil {
		panic(err)
	}
	defer objs.Close()

	//fmt.Printf("Attached XDP program to iface %q (index %d)", iface.Name, iface.Index)
	fmt.Printf("Press Ctrl-C to exit and remove the program")

	// Print the contents of the BPF hash map (source IP address -> packet count).
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s, err := formatMapContents(objs.UpfPipeline)
		if err != nil {
			fmt.Printf("Error reading map: %s", err)
			continue
		}
		fmt.Printf("Pipeline map contents:\n%s", s)
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
