package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dropbox/goebpf"
)

var iface = flag.String("iface", "", "Interface to bind XDP program to")
var elf = flag.String("elf", "ebpf_prog/xdp_fw.elf", "clang/llvm compiled binary file")
var programName = flag.String("program", "upf_far_program_func", "ebpf program name")

func main() {
	flag.Parse()

	if *elf == "" {
		fmt.Printf("--elf is required\n")
		os.Exit(1)
	}

	bpf := goebpf.NewDefaultEbpfSystem()

	//if err := bpf.LoadElf("../../build/CMakeFiles/eupf.dir/src/xdp/ip_entrypoint.c.o"); err != nil {
	if err := bpf.LoadElf(*elf); err != nil {
		fmt.Printf("Load elf %s failed: %v\n", *elf, err)
		os.Exit(1)
	}

	printBpfInfo(bpf)

	if *iface != "" && *programName != "" {
		program := bpf.GetProgramByName(*programName)
		if program == nil {
			fmt.Printf("No ebpf program %s\n", *programName)
			os.Exit(1)
		}

		if err := program.Load(); err != nil {
			fmt.Printf("Load ebpf to kernel failed: %v\n", err)
			os.Exit(1)

		}

		if err := program.Attach(*iface); err != nil {
			fmt.Printf("Can't attach ebpf program to interface %s: %v\n", *iface, err)
			os.Exit(1)
		}

		defer program.Detach()
	}

	// Interact with program is simply done through maps:
	upfPipeline := bpf.GetMapByName("upf_pipeline") // name also matches BPF_MAP_ADD(drops)
	val, err := upfPipeline.LookupInt(0)            // Get value from map at index 0
	if err == nil {
		fmt.Printf("Drops: %d\n", val)
	}
}

func printBpfInfo(bpf goebpf.System) {
	fmt.Println("ebpf maps:")
	for _, item := range bpf.GetMaps() {
		fmt.Printf("\t%s: %v, Fd %v\n", item.GetName(), item.GetType(), item.GetFd())
	}
	fmt.Println("\nebpf programs:")
	for _, prog := range bpf.GetPrograms() {
		fmt.Printf("\t%s: %v, size %d, license \"%s\"\n",
			prog.GetName(), prog.GetType(), prog.GetSize(), prog.GetLicense(),
		)

	}
	fmt.Println()
}
