package main

import (
	"encoding/binary"
	"fmt"
	"net"
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
	ProgramId              uint32 `json:"id"`
	ProgramRef             uint32 `json:"fd"`
	ProgramName            string `json:"name"`
	ProgramRunCount        uint32 `json:"run_count"`
	ProgramRunCountEnabled bool   `json:"run_count_enabled"`
	ProgramDuration        uint32 `json:"duration"`
	ProgramDurationEnabled bool   `json:"duration_enabled"`
}

func ListMapProgArrayContents(m *ebpf.Map) ([]BpfMapProgArrayMember, error) {
	if m.Type() != ebpf.ProgramArray {
		return nil, fmt.Errorf("map is not a program array")
	}
	var bpfMapProgArrayMember []BpfMapProgArrayMember
	var (
		key uint32
		val *ebpf.Program
	)

	iter := m.Iterate()
	for iter.Next(&key, &val) {
		programInfo, _ := val.Info()
		programID, _ := programInfo.ID()
		runCount, runCountEnabled := programInfo.RunCount()
		runDuration, runDurationEnabled := programInfo.Runtime()
		bpfMapProgArrayMember = append(bpfMapProgArrayMember,
			BpfMapProgArrayMember{
				ProgramId:              key,
				ProgramRef:             uint32(programID),
				ProgramName:            programInfo.Name,
				ProgramRunCount:        uint32(runCount),
				ProgramRunCountEnabled: runCountEnabled,
				ProgramDuration:        uint32(runDuration),
				ProgramDurationEnabled: runDurationEnabled,
			})
	}
	return bpfMapProgArrayMember, iter.Err()
}

type ContextMapElement struct {
	UeIpAddress        string `json:"ue_ip"`
	TEID               uint32 `json:"teid"`
	TunnelSrcIpAddress string `json:"tunnel_src_ip"`
	TunnelDstIpAddress string `json:"tunnel_dst_ip"`
	TunnelDstPort      uint16 `json:"tunnel_dst_port"`
}

func ListContextMapContents(m *ebpf.Map) ([]ContextMapElement, error) {
	if m.Type() != ebpf.Hash {
		return nil, fmt.Errorf("map %s is not a hash", m)
	}

	contextMap := []ContextMapElement{}

	type ContextMapValueStruct struct {
		Teid    uint32
		Srcip   uint32
		Dstip   uint32
		Dstport uint16
	}

	var key []byte
	var value ContextMapValueStruct

	iter := m.Iterate()
	for iter.Next(&key, &value) {
		ueIP := net.IP(key)
		contextMap = append(contextMap,
			ContextMapElement{
				UeIpAddress:        ueIP.String(),
				TEID:               value.Teid,
				TunnelSrcIpAddress: intToIP(value.Srcip).String(),
				TunnelDstIpAddress: intToIP(value.Dstip).String(),
				TunnelDstPort:      value.Dstport})

	}
	return contextMap, iter.Err()
}

// intToIP converts IPv4 number to net.IP
func intToIP(ipNum uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, ipNum)
	return ip
}

type QerMapElement struct {
	Id           uint32 `json:"id"`
	GateStatusUL uint8  `json:"gate_status_ul"`
	GateStatusDL uint8  `json:"gate_status_dl"`
	Qfi          uint8  `json:"qfi"`
	MaxBitrateUL uint64 `json:"max_bitrate_ul"`
	MaxBitrateDL uint64 `json:"max_bitrate_dl"`
}

func ListQerMapContents(m *ebpf.Map) ([]QerMapElement, error) {
	if m.Type() != ebpf.Hash {
		return nil, fmt.Errorf("map %s is not a hash", m)
	}

	contextMap := []QerMapElement{}

	var key uint32
	var value QerInfo

	iter := m.Iterate()
	for iter.Next(&key, &value) {
		id := key
		contextMap = append(contextMap,
			QerMapElement{
				Id:           id,
				GateStatusUL: value.GateStatusUL,
				GateStatusDL: value.GateStatusDL,
				Qfi:          value.Qfi,
				MaxBitrateUL: value.MaxBitrateUL,
				MaxBitrateDL: value.MaxBitrateDL,
			})

	}
	return contextMap, iter.Err()
}
