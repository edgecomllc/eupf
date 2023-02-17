package main

import (
	"fmt"
	"strings"

	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"
)

// increaseResourceLimits https://prototype-kernel.readthedocs.io/en/latest/bpf/troubleshooting.html#memory-ulimits
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

// https://man7.org/linux/man-pages/man2/bpf.2.html
// A program array map is a special kind of array map whose
// map values contain only file descriptors referring to
// other eBPF programs.  Thus, both the key_size and
// value_size must be exactly four bytes.
type BpfMapProgArrayMember struct {
	ProgramId  uint32 `json:"program_id"`
	ProgramRef uint32 `json:"program_ref"`
}

func ListMapProgArrayContents(m *ebpf.Map) ([]BpfMapProgArrayMember, error) {
	if m.Type() != ebpf.ProgramArray {
		return nil, fmt.Errorf("map is not a program array")
	}
	var bpfMapProgArrayMember []BpfMapProgArrayMember
	var (
		key uint32
		val uint32
	)
	iter := m.Iterate()
	for iter.Next(&key, &val) {
		bpfMapProgArrayMember = append(bpfMapProgArrayMember, BpfMapProgArrayMember{ProgramId: key, ProgramRef: val})
	}
	return bpfMapProgArrayMember, iter.Err()
}
