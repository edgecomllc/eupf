package core

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func TestHeartbeat(t *testing.T) {
	// Create pfcp connection struct
	pfcpConn := PfcpConnection{}
	hbReq := message.NewHeartbeatRequest(0,
		ie.NewRecoveryTimeStamp(time.Now()),
		nil,
	)
	response, err := HandlePfcpHeartbeatRequest(&pfcpConn, hbReq, "127.0.0.1")
	if err != nil {
		t.Errorf("Error handling heartbeat request: %s", err)
	}
	if response == nil {
		t.Errorf("No response from heartbeat request")
	}
	ts, err := response.(*message.HeartbeatResponse).RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		t.Errorf("Error getting timestamp from heartbeat response: %s", err)
	}
	t.Logf("Received response from heartbeat request with timestamp: %s", ts)
}

func TestAssociationSetup(t *testing.T) {
	// Create pfcp connection struct
	pfcpConn := PfcpConnection{
		NodeAssociations: make(map[string]*NodeAssociation),
		nodeId:           "test-node",
	}
	asReq := message.NewAssociationSetupRequest(0,
		ie.NewNodeID("", "", "test"),
	)

	remoteIP := "127.0.0.1"
	response, err := HandlePfcpAssociationSetupRequest(&pfcpConn, asReq, remoteIP)
	if err != nil {
		t.Errorf("Error handling association setup request: %s", err)
	}
	cause, err := response.(*message.AssociationSetupResponse).Cause.Cause()
	if err != nil {
		t.Errorf("Error getting cause from association setup response: %s", err)
	}
	if cause != ie.CauseRequestAccepted {
		t.Errorf("Unexpected cause in association setup response: %d", cause)
	}
	// Check nodeId in response
	nodeId, err := response.(*message.AssociationSetupResponse).NodeID.NodeID()
	if err != nil {
		t.Errorf("Error getting node ID from association setup response: %s", err)
	}
	if nodeId != "test-node" {
		t.Errorf("Unexpected node ID in association setup response: %s", nodeId)
	}
	if _, ok := pfcpConn.NodeAssociations[remoteIP]; !ok {
		t.Errorf("Association not created")
	}
}

func TestSdfFilterStoring(t *testing.T) {

	mapOps := MapOperationsMock{}

	var pfcpHandlers = PfcpHandlerMap{
		message.MsgTypeHeartbeatRequest:            HandlePfcpHeartbeatRequest,
		message.MsgTypeAssociationSetupRequest:     HandlePfcpAssociationSetupRequest,
		message.MsgTypeSessionEstablishmentRequest: HandlePfcpSessionEstablishmentRequest,
		message.MsgTypeSessionDeletionRequest:      HandlePfcpSessionDeletionRequest,
		message.MsgTypeSessionModificationRequest:  HandlePfcpSessionModificationRequest,
	}

	smfIP := "127.0.0.1"
	pfcpConn := PfcpConnection{
		NodeAssociations: make(map[string]*NodeAssociation),
		nodeId:           "test-node",
		mapOperations:    &mapOps,
		pfcpHandlerMap:   pfcpHandlers,
	}
	asReq := message.NewAssociationSetupRequest(0,
		ie.NewNodeID("", "", "test"),
	)
	response, err := HandlePfcpAssociationSetupRequest(&pfcpConn, asReq, smfIP)
	if err != nil {
		t.Errorf("Error handling association setup request: %s", err)
	}
	cause, err := response.(*message.AssociationSetupResponse).Cause.Cause()
	if err != nil {
		t.Errorf("Error getting cause from association setup response: %s", err)
	}
	if cause != ie.CauseRequestAccepted {
		t.Errorf("Unexpected cause in association setup response: %d", cause)
	}
	// Check nodeId in response
	nodeId, err := response.(*message.AssociationSetupResponse).NodeID.NodeID()
	if err != nil {
		t.Errorf("Error getting node ID from association setup response: %s", err)
	}
	if nodeId != "test-node" {
		t.Errorf("Unexpected node ID in association setup response: %s", nodeId)
	}
	if _, ok := pfcpConn.NodeAssociations[smfIP]; !ok {
		t.Errorf("Association not created")
	}

	ip1, _ := net.ResolveIPAddr("ip", "1.1.1.1")

	// Create various Sdf Filters.
	// And send a bunch of requests (with different seid-s).

	fds := [...]SdfFilterTestStruct{
		{FlowDescription: "permit out ip from 10.62.0.1 to 8.8.8.8/32", Protocol: 1,
			SrcType: 1, SrcAddress: "10.62.0.1", SrcMask: "<nil>", SrcPortLower: 0, SrcPortUpper: 65535,
			DstType: 1, DstAddress: "8.8.8.8", DstMask: "ffffffff", DstPortLower: 0, DstPortUpper: 65535},
		{FlowDescription: "permit out tcp from 1.1.1.1/20 80 to 100.1.2.3 9121-10202", Protocol: 2,
			SrcType: 1, SrcAddress: "1.1.0.0", SrcMask: "fffff000", SrcPortLower: 80, SrcPortUpper: 80,
			DstType: 1, DstAddress: "100.1.2.3", DstMask: "<nil>", DstPortLower: 9121, DstPortUpper: 10202},
		{FlowDescription: "permit out udp from 2001:db8:3333:4444:CCCC:DDDD:EEEE:FFFF 8080-8081 to 2001:0db8::42/30", Protocol: 3,
			SrcType: 2, SrcAddress: "2001:db8:3333:4444:cccc:dddd:eeee:ffff", SrcMask: "<nil>", SrcPortLower: 8080, SrcPortUpper: 8081,
			DstType: 2, DstAddress: "2001:db8::", DstMask: "fffffffc000000000000000000000000", DstPortLower: 0, DstPortUpper: 65535},
		{FlowDescription: "permit out icmp from any 4-5 to ::1234:5678/2 2", Protocol: 0,
			SrcType: 0, SrcAddress: "<nil>", SrcMask: "<nil>", SrcPortLower: 4, SrcPortUpper: 5,
			DstType: 2, DstAddress: "::", DstMask: "c0000000000000000000000000000000", DstPortLower: 2, DstPortUpper: 2},
	}

	for i := 0; i < len(fds); i++ {
		seReq1 := message.NewSessionEstablishmentRequest(0, 0,
			uint64(i+1), 1, 0,
			ie.NewNodeID("", "", "test"),
			ie.NewFSEID(uint64(i+1), net.ParseIP(smfIP), nil),
			ie.NewCreatePDR(
				ie.NewPDRID(1),
				ie.NewPDI(
					ie.NewSourceInterface(ie.SrcInterfaceCore),
					//ie.NewFTEID(0, 0, ip1.IP, nil, 0),
					ie.NewUEIPAddress(2, ip1.IP.String(), "", 0, 0),
					ie.NewSDFFilter(fds[i].FlowDescription, "", "", "", 0),
				),
			),
		)

		_, err = HandlePfcpSessionEstablishmentRequest(&pfcpConn, seReq1, smfIP)
		if err != nil {
			t.Errorf("Error handling session establishment request: %s", err)
		}

		// Check that session PDRs are correct
		if pfcpConn.NodeAssociations[smfIP].Sessions[uint64(i+1+1)].PDRs[1].Ipv4.String() != "1.1.1.1" {
			t.Errorf("Iteration 1, got broken")
		}

		err := CheckSdfFilterEquality(i+1, pfcpConn.NodeAssociations[smfIP].Sessions[uint64(i+1+1)].PDRs[1].PdrInfo.SdfFilter, fds[i])
		if err != "" {
			t.Error(err)
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

func CheckSdfFilterEquality(i int, sdfFilter ebpf.SdfFilter, fd SdfFilterTestStruct) string {
	if sdfFilter.Protocol != fd.Protocol {
		return fmt.Sprintf("Iteration %d, wrong Protocol, expected: %d, got: %d", i, fd.Protocol, sdfFilter.Protocol)
	}
	if sdfFilter.SrcAddress.Type != fd.SrcType {
		return fmt.Sprintf("Iteration %d, wrong SrcType, expected: %d, got: %d", i, fd.SrcType, sdfFilter.SrcAddress.Type)
	}
	if sdfFilter.SrcAddress.Ip.String() != fd.SrcAddress {
		return fmt.Sprintf("Iteration %d, wrong SrcAddress, expected: %s, got: %s", i, fd.SrcAddress, sdfFilter.SrcAddress.Ip.String())
	}
	if sdfFilter.SrcAddress.Mask.String() != fd.SrcMask {
		return fmt.Sprintf("Iteration %d, wrong SrcMask, expected: %s, got: %s", i, fd.SrcMask, sdfFilter.SrcAddress.Mask.String())
	}
	if sdfFilter.SrcPortRange.LowerBound != fd.SrcPortLower {
		return fmt.Sprintf("Iteration %d, wrong SrcPortLower, expected: %d, got: %d", i, fd.SrcPortLower, sdfFilter.SrcPortRange.LowerBound)
	}
	if sdfFilter.SrcPortRange.UpperBound != fd.SrcPortUpper {
		return fmt.Sprintf("Iteration %d, wrong SrcPortUpper, expected: %d, got: %d", i, fd.SrcPortUpper, sdfFilter.SrcPortRange.UpperBound)
	}
	if sdfFilter.DstAddress.Type != fd.DstType {
		return fmt.Sprintf("Iteration %d, wrong DstType, expected: %d, got: %d", i, fd.DstType, sdfFilter.DstAddress.Type)
	}
	if sdfFilter.DstAddress.Ip.String() != fd.DstAddress {
		return fmt.Sprintf("Iteration %d, wrong DstAddress, expected: %s, got: %s", i, fd.DstAddress, sdfFilter.DstAddress.Ip.String())
	}
	if sdfFilter.DstAddress.Mask.String() != fd.DstMask {
		return fmt.Sprintf("Iteration %d, wrong DstMask, expected: %s, got: %s", i, fd.DstMask, sdfFilter.DstAddress.Mask.String())
	}
	if sdfFilter.DstPortRange.LowerBound != fd.DstPortLower {
		return fmt.Sprintf("Iteration %d, wrong DstPortLower, expected: %d, got: %d", i, fd.DstPortLower, sdfFilter.DstPortRange.LowerBound)
	}
	if sdfFilter.DstPortRange.UpperBound != fd.DstPortUpper {
		return fmt.Sprintf("Iteration %d, wrong DstPortUpper, expected: %d, got: %d", i, fd.DstPortUpper, sdfFilter.DstPortRange.UpperBound)
	}
	return ""
}
