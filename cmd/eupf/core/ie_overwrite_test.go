package core

import (
	"fmt"
	"github.com/edgecomllc/eupf/cmd/eupf/ebpf"
	"net"
	"testing"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func TestSessionOverwrite(t *testing.T) {
	bpfObjects := &ebpf.BpfObjects{
		FarIdTracker: ebpf.NewIdTracker(100),
		QerIdTracker: ebpf.NewIdTracker(100),
	}
	if err := bpfObjects.Load(); err != nil {
		t.Errorf("Loading bpf objects failed: %s", err.Error())
	}
	// Create pfcp connection struct
	pfcpConn := PfcpConnection{
		NodeAssociations: NodeAssociationMap{},
		nodeId:           "test-node",
		mapOperations:    bpfObjects,
	}
	asReq := message.NewAssociationSetupRequest(0,
		ie.NewNodeID("", "", "test"),
	)
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:35655")
	if err != nil {
		t.Errorf("Error resolving UDP address: %s", err)
	}
	response, err := HandlePfcpAssociationSetupRequest(&pfcpConn, asReq, udpAddr)
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
	if _, ok := pfcpConn.NodeAssociations[udpAddr.String()]; !ok {
		t.Errorf("Association not created")
	}

	ip1, _ := net.ResolveIPAddr("ip", "1.1.1.1")
	ip2, _ := net.ResolveIPAddr("ip", "2.2.2.2")
	// Create two send two Session Establishment Requests with downlink PDRs
	// and check that the first session is not overwritten
	seReq1 := message.NewSessionEstablishmentRequest(0, 0,
		1, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, udpAddr.IP, nil),
		ie.NewCreatePDR(
			ie.NewPDRID(1),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				ie.NewFTEID(0, 0, ip1.IP, nil, 0),
				ie.NewUEIPAddress(2, ip1.IP.String(), "", 0, 0),
			),
		),
	)

	seReq2 := message.NewSessionEstablishmentRequest(0, 0,
		2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(2, udpAddr.IP, nil),
		ie.NewCreatePDR(
			ie.NewPDRID(1),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				ie.NewFTEID(0, 0, ip2.IP, nil, 0),
				ie.NewUEIPAddress(2, ip2.IP.String(), "", 0, 0),
			),
		),
	)

	// Send first request
	_, err = HandlePfcpSessionEstablishmentRequest(&pfcpConn, seReq1, udpAddr)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	// Send second request
	_, err = HandlePfcpSessionEstablishmentRequest(&pfcpConn, seReq2, udpAddr)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	// Check that session PDRs are correct
	if pfcpConn.NodeAssociations[udpAddr.String()].Sessions[2].DownlinkPDRs[1].Ipv4.String() != "1.1.1.1" {
		t.Errorf("Session 1, got broken")
	}
	if pfcpConn.NodeAssociations[udpAddr.String()].Sessions[3].DownlinkPDRs[1].Ipv4.String() != "2.2.2.2" {
		t.Errorf("Session 2, got broken")
	}

	// Send Session Modification Request, create FAR
	smReq := message.NewSessionModificationRequest(0, 0,
		2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(2, udpAddr.IP, nil),
		ie.NewCreateFAR(
			ie.NewFARID(1),
			ie.NewApplyAction(2),
			ie.NewForwardingParameters(
				ie.NewDestinationInterface(ie.DstInterfaceAccess),
				ie.NewNetworkInstance(""),
			),
		),
	)

	// Send modification request
	_, err = HandlePfcpSessionModificationRequest(&pfcpConn, smReq, udpAddr)
	if err != nil {
		t.Errorf("Error handling session modification request: %s", err)
	}

	// Check that session PDRs are correct
	if pfcpConn.NodeAssociations[udpAddr.String()].Sessions[2].DownlinkPDRs[1].Ipv4.String() != "1.1.1.1" {
		t.Errorf("Session 1, got broken")
	}
	if pfcpConn.NodeAssociations[udpAddr.String()].Sessions[3].DownlinkPDRs[1].Ipv4.String() != "2.2.2.2" {
		t.Errorf("Session 2, got broken")
	}
	fmt.Printf("%+v", pfcpConn.NodeAssociations)

}
