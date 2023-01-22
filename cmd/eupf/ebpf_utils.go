package main

import (
	"fmt"
	"strings"

	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"
)

//increaseResourceLimits https://prototype-kernel.readthedocs.io/en/latest/bpf/troubleshooting.html#memory-ulimits
func IncreaseResourceLimits() error {
	return unix.Setrlimit(unix.RLIMIT_MEMLOCK, &unix.Rlimit{
		Cur: unix.RLIM_INFINITY,
		Max: unix.RLIM_INFINITY,
	})
}

func FormatMapContents(m *ebpf.Map) (string, error) {
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
