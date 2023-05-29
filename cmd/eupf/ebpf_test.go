// Copyright 2023 Edgecom LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func testArp(bpfObjects *BpfObjects) error {

	packetArp := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(packetArp, gopacket.SerializeOptions{},
		&layers.Ethernet{
			SrcMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 10},
			DstMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 20},
			EthernetType: layers.EthernetTypeARP,
		},
		&layers.ARP{},
	); err != nil {
		return fmt.Errorf("serializing input packet failed: %v", err)
	}

	bpfRet, bufOut, err := bpfObjects.ip_entrypointPrograms.UpfIpEntrypointFunc.Test(packetArp.Bytes())
	if err != nil {
		return fmt.Errorf("test run failed: %v", err)
	}
	if bpfRet != 2 { // XDP_PASS
		return fmt.Errorf("unexpected return value: %d", bpfRet)
	}
	if !reflect.DeepEqual(bufOut, packetArp.Bytes()) {
		return errors.New("unexpected data modification")
	}

	return nil
}

func testGtpEcho(bpfObjects *BpfObjects) error {

	packetArp := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(packetArp, gopacket.SerializeOptions{},
		&layers.Ethernet{
			SrcMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 10},
			DstMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 20},
			EthernetType: layers.EthernetTypeIPv4,
		},
		&layers.IPv4{
			Version:  4,
			DstIP:    net.IP{10, 3, 0, 10},
			SrcIP:    net.IP{10, 3, 0, 20},
			Protocol: layers.IPProtocolUDP,
			IHL:      5,
		},
		&layers.UDP{
			DstPort: 2152,
			SrcPort: 2152,
		},
		&layers.GTPv1U{
			Version:        1,
			MessageType:    1, // GTPU_ECHO_REQUEST
			TEID:           0,
			SequenceNumber: 0,
		},
	); err != nil {
		return fmt.Errorf("serializing input packet failed: %v", err)
	}

	bpfRet, bufOut, err := bpfObjects.ip_entrypointPrograms.UpfIpEntrypointFunc.Test(packetArp.Bytes())
	if err != nil {
		return fmt.Errorf("test run failed: %v", err)
	}
	if bpfRet != 3 { // XDP_TX
		return fmt.Errorf("unexpected return value: %d", bpfRet)
	}

	response := gopacket.NewPacket(bufOut, layers.LayerTypeEthernet, gopacket.Default)
	if gtpLayer := response.Layer(layers.LayerTypeGTPv1U); gtpLayer != nil {
		gtp, _ := gtpLayer.(*layers.GTPv1U)

		if gtp.MessageType != 2 { //GTPU_ECHO_RESPONSE
			return fmt.Errorf("unexpected gtp response: %d", gtp.MessageType)
		}
		if gtp.SequenceNumber != 0 {
			return fmt.Errorf("unexpected gtp sequence: %d", gtp.SequenceNumber)
		}
		if gtp.TEID != 0 {
			return fmt.Errorf("unexpected gtp TEID: %d", gtp.TEID)
		}
	} else {
		return errors.New("unexpected response")
	}

	return nil
}

func TestEntrypoint(t *testing.T) {

	if err := IncreaseResourceLimits(); err != nil {
		t.Fatalf("Can't increase resource limits: %s", err.Error())
	}

	bpfObjects := &BpfObjects{}

	if err := bpfObjects.Load(); err != nil {
		t.Fatalf("Loading bpf objects failed: %s", err.Error())
	}

	defer bpfObjects.Close()

	t.Run("Arp test", func(t *testing.T) {
		err := testArp(bpfObjects)
		if err != nil {
			t.Fatalf("test failed: %s", err)
		}
	})

	t.Run("GTP-U Echo test", func(t *testing.T) {
		err := testGtpEcho(bpfObjects)
		if err != nil {
			t.Fatalf("test failed: %s", err)
		}
	})
}
