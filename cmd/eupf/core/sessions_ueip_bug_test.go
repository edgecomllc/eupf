package core

import (
	"fmt"
	"github.com/edgecomllc/eupf/cmd/eupf/ebpf"
	"net"
	"testing"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

// Test for the bug where the session's UE IP address is overwritten because it is a pointer to the buffer for incoming UDP packets.
func TestSessionUEIpOverwrite(t *testing.T) {

	bpfObjects := &ebpf.BpfObjects{
		FarIdTracker: ebpf.NewIdTracker(100),
		QerIdTracker: ebpf.NewIdTracker(100),
	}
	if err := bpfObjects.Load(); err != nil {
		t.Errorf("Loading bpf objects failed: %s", err.Error())
	}

	var pfcpHandlers = PfcpHandlerMap{
		message.MsgTypeHeartbeatRequest:            HandlePfcpHeartbeatRequest,
		message.MsgTypeAssociationSetupRequest:     HandlePfcpAssociationSetupRequest,
		message.MsgTypeSessionEstablishmentRequest: HandlePfcpSessionEstablishmentRequest,
		message.MsgTypeSessionDeletionRequest:      HandlePfcpSessionDeletionRequest,
		message.MsgTypeSessionModificationRequest:  HandlePfcpSessionModificationRequest,
	}
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:35655")
	if err != nil {
		t.Errorf("Error resolving UDP address: %s", err)
	}
	udpConn, _ := net.ListenUDP("udp", udpAddr)
	pfcpConn := PfcpConnection{
		NodeAssociations: NodeAssociationMap{},
		nodeId:           "test-node",
		mapOperations:    bpfObjects,
		pfcpHandlerMap:   pfcpHandlers,
		udpConn:          udpConn,
	}
	asReq := message.NewAssociationSetupRequest(0,
		ie.NewNodeID("", "", "test"),
	)
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

	//uip1, _ := net.ResolveUDPAddr("ip", "1.1.1.1")
	//uip2, _ := net.ResolveUDPAddr("ip", "2.2.2.2")

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

	buf := make([]byte, 1500)
	bytes1, _ := seReq1.Marshal()
	copy(buf, bytes1)
	pfcpConn.Handle(buf, udpAddr)

	bytes2, _ := seReq2.Marshal()
	copy(buf, bytes2)
	pfcpConn.Handle(buf, udpAddr)

	// Check that session PDRs are correct
	if pfcpConn.NodeAssociations[udpAddr.String()].Sessions[2].DownlinkPDRs[1].Ipv4.String() != "1.1.1.1" {
		t.Errorf("Session 1, got broken")
	}
	if pfcpConn.NodeAssociations[udpAddr.String()].Sessions[3].DownlinkPDRs[1].Ipv4.String() != "2.2.2.2" {
		t.Errorf("Session 2, got broken")
	}
	fmt.Printf("%+v", pfcpConn.NodeAssociations)

}
