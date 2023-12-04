package core

import (
	"fmt"
	"testing"

	"github.com/edgecomllc/eupf/cmd/ebpf"
)

func TestSdfFilterParseValid(t *testing.T) {
	fds := [...]SdfFilterTestStruct{
		{FlowDescription: "permit out ip from 10.62.0.1 to 8.8.8.8/32", Protocol: 1,
			SrcType: 1, SrcAddress: "10.62.0.1", SrcMask: "ffffffff", SrcPortLower: 0, SrcPortUpper: 65535,
			DstType: 1, DstAddress: "8.8.8.8", DstMask: "ffffffff", DstPortLower: 0, DstPortUpper: 65535},
		{FlowDescription: "permit out tcp from 1.1.1.1/20 80 to 100.1.2.3 9121-10202", Protocol: 2,
			SrcType: 1, SrcAddress: "1.1.0.0", SrcMask: "fffff000", SrcPortLower: 80, SrcPortUpper: 80,
			DstType: 1, DstAddress: "100.1.2.3", DstMask: "ffffffff", DstPortLower: 9121, DstPortUpper: 10202},
		{FlowDescription: "permit out udp from 2001:db8:3333:4444:CCCC:DDDD:EEEE:FFFF 8080-8081 to 2001:0db8::42/30", Protocol: 3,
			SrcType: 2, SrcAddress: "2001:db8:3333:4444:cccc:dddd:eeee:ffff", SrcMask: "ffffffffffffffffffffffffffffffff", SrcPortLower: 8080, SrcPortUpper: 8081,
			DstType: 2, DstAddress: "2001:db8::", DstMask: "fffffffc000000000000000000000000", DstPortLower: 0, DstPortUpper: 65535},
		{FlowDescription: "permit out icmp from any 4-5 to ::1234:5678/2 2", Protocol: 0,
			SrcType: 0, SrcAddress: "<nil>", SrcMask: "<nil>", SrcPortLower: 4, SrcPortUpper: 5,
			DstType: 2, DstAddress: "::", DstMask: "c0000000000000000000000000000000", DstPortLower: 2, DstPortUpper: 2},
		{FlowDescription: "permit out 58 from ff02::2/128 to assigned", Protocol: 4,
			SrcType: 2, SrcAddress: "ff02::2", SrcMask: "ffffffffffffffffffffffffffffffff", SrcPortLower: 0, SrcPortUpper: 65535,
			DstType: 0, DstAddress: "<nil>", DstMask: "<nil>", DstPortLower: 0, DstPortUpper: 65535},
		{FlowDescription: "permit out ip from 10.60.0.0/16 to any", Protocol: 1,
			SrcType: 1, SrcAddress: "10.60.0.0", SrcMask: "ffff0000", SrcPortLower: 0, SrcPortUpper: 65535,
			DstType: 0, DstAddress: "<nil>", DstMask: "<nil>", DstPortLower: 0, DstPortUpper: 65535},
	}

	for i := 0; i < len(fds); i++ {
		if sdfFilter, err := ParseSdfFilter(fds[i].FlowDescription); err == nil {
			if err := CheckSdfFilterEquality(&sdfFilter, fds[i]); err != nil {
				t.Errorf("Iteration %d.\nFlowDescription: %s\nError: %s", i, fds[i].FlowDescription, err.Error())
			}
		} else {
			t.Errorf("Unexpected error while parsing SDF filter.\nFlowDescription: %s\nError: %s\n", fds[i].FlowDescription, err.Error())
		}
	}
}

func TestSdfFilterParseInvalid(t *testing.T) {
	fds := [...]string{
		// Unsupported (deny, in, option)
		"deny out ip from 10.62.0.1 to 8.8.8.8/32",
		"permit in tcp from 1.1.1.1/20 80 to 100.1.2.3 9121-10202",
		"permit out udp from 2001:db8:3333:4444:CCCC:DDDD:EEEE:FFFF 8080-8081 to 2001:0db8::42/30 option 2confidential",
		// Bad format
		"permit out icmp ? from any 4-5 to ::1234:5678/2 2",
		"permit out ip from 10.62.0.1:20 to 8.8.8.8/32",
		"permit out ip to 10.62.0.1 from 8.8.8.8/32",
		"permit out ip from 10.62.0.1 to 8.8.8.8 /32",
		"permit out ip from 10.62.0.1 to 8.8.8.8/",
		"permit out ip from 10.62.0.1/2-3 to 8.8.8.8",
		// Bad data
		"permit out ip from 10.62.0.1.1 to 8.8.8.8/32",
		"permit out ip from 10.0.1 to 8.8.8.8/32",
		"permit out ssl from 10.62.0.1 to 8.8.8.8/32",
		"permit out ip from 10.62.0.1 80-70 to 8.8.8.8/32",
		"permit out ip from 10.62.0.1 100500 to 8.8.8.8/32",
		"permit out ip from 10.62.0.1 100500 to 8.8.8.8/32",
		"permit out icmp from any 4-5 to :::1234:5678/2 2",
		"permit out icmp from any/23 to ::1234:5678/2 2",
		"permit out icmp from any 4-5 to ::abcd:efgh/2 2",
	}

	for i := 0; i < len(fds); i++ {
		if _, err := ParseSdfFilter(fds[i]); err == nil {
			t.Errorf("Iteration %d.\nFlowDescription: %s\nAn error should appear when parsing SDF", i, fds[i])
		}
	}
}

type SdfFilterTestStruct struct {
	FlowDescription string
	Protocol        uint8
	SrcType         uint8
	SrcAddress      string
	SrcMask         string
	SrcPortLower    uint16
	SrcPortUpper    uint16
	DstType         uint8
	DstAddress      string
	DstMask         string
	DstPortLower    uint16
	DstPortUpper    uint16
}

func CheckSdfFilterEquality(sdfFilter *ebpf.SdfFilter, fd SdfFilterTestStruct) error {
	if sdfFilter == nil {
		return fmt.Errorf("Wrong SdfFilter, expected: not nil, got: nil")
	}
	if sdfFilter.Protocol != fd.Protocol {
		return fmt.Errorf("Wrong Protocol, expected: %d, got: %d", fd.Protocol, sdfFilter.Protocol)
	}
	if sdfFilter.SrcAddress.Type != fd.SrcType {
		return fmt.Errorf("Wrong SrcType, expected: %d, got: %d", fd.SrcType, sdfFilter.SrcAddress.Type)
	}
	if sdfFilter.SrcAddress.Ip.String() != fd.SrcAddress {
		return fmt.Errorf("Wrong SrcAddress, expected: %s, got: %s", fd.SrcAddress, sdfFilter.SrcAddress.Ip.String())
	}
	if sdfFilter.SrcAddress.Mask.String() != fd.SrcMask {
		return fmt.Errorf("Wrong SrcMask, expected: %s, got: %s", fd.SrcMask, sdfFilter.SrcAddress.Mask.String())
	}
	if sdfFilter.SrcPortRange.LowerBound != fd.SrcPortLower {
		return fmt.Errorf("Wrong SrcPortLower, expected: %d, got: %d", fd.SrcPortLower, sdfFilter.SrcPortRange.LowerBound)
	}
	if sdfFilter.SrcPortRange.UpperBound != fd.SrcPortUpper {
		return fmt.Errorf("Wrong SrcPortUpper, expected: %d, got: %d", fd.SrcPortUpper, sdfFilter.SrcPortRange.UpperBound)
	}
	if sdfFilter.DstAddress.Type != fd.DstType {
		return fmt.Errorf("Wrong DstType, expected: %d, got: %d", fd.DstType, sdfFilter.DstAddress.Type)
	}
	if sdfFilter.DstAddress.Ip.String() != fd.DstAddress {
		return fmt.Errorf("Wrong DstAddress, expected: %s, got: %s", fd.DstAddress, sdfFilter.DstAddress.Ip.String())
	}
	if sdfFilter.DstAddress.Mask.String() != fd.DstMask {
		return fmt.Errorf("Wrong DstMask, expected: %s, got: %s", fd.DstMask, sdfFilter.DstAddress.Mask.String())
	}
	if sdfFilter.DstPortRange.LowerBound != fd.DstPortLower {
		return fmt.Errorf("Wrong DstPortLower, expected: %d, got: %d", fd.DstPortLower, sdfFilter.DstPortRange.LowerBound)
	}
	if sdfFilter.DstPortRange.UpperBound != fd.DstPortUpper {
		return fmt.Errorf("Wrong DstPortUpper, expected: %d, got: %d", fd.DstPortUpper, sdfFilter.DstPortRange.UpperBound)
	}
	return nil
}
