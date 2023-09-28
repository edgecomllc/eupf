package core

import (
	"net"
	"testing"
	"time"

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

func SdfFilterStorePreSetup(t *testing.T) (PfcpConnection, string) {
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
	return pfcpConn, smfIP
}

func TestSdfFilterStoreValid(t *testing.T) {

	pfcpConn, smfIP := SdfFilterStorePreSetup(t)

	ip1, _ := net.ResolveIPAddr("ip", "1.1.1.1")
	ip2, _ := net.ResolveIPAddr("ip", "2.2.2.2")

	fd := SdfFilterTestStruct{FlowDescription: "permit out ip from 10.62.0.1 to 8.8.8.8/32", Protocol: 1,
		SrcType: 1, SrcAddress: "10.62.0.1", SrcMask: "<nil>", SrcPortLower: 0, SrcPortUpper: 65535,
		DstType: 1, DstAddress: "8.8.8.8", DstMask: "ffffffff", DstPortLower: 0, DstPortUpper: 65535}

	// Request with UEIP Address
	seReq1 := message.NewSessionEstablishmentRequest(0, 0,
		1, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, net.ParseIP(smfIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(1),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				//ie.NewFTEID(0, 0, ip1.IP, nil, 0),
				ie.NewUEIPAddress(2, ip1.IP.String(), "", 0, 0),
				ie.NewSDFFilter(fd.FlowDescription, "", "", "", 0),
			),
		),
	)

	// Request with TEID
	seReq2 := message.NewSessionEstablishmentRequest(0, 0,
		2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(2, net.ParseIP(smfIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(1),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				ie.NewFTEID(0, 0, ip2.IP, nil, 0),
				// ie.NewUEIPAddress(2, ip2.IP.String(), "", 0, 0),
				ie.NewSDFFilter(fd.FlowDescription, "", "", "", 0),
			),
		),
	)

	var err error
	_, err = HandlePfcpSessionEstablishmentRequest(&pfcpConn, seReq1, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	_, err = HandlePfcpSessionEstablishmentRequest(&pfcpConn, seReq2, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	// Check that session PDRs are correct
	if pfcpConn.NodeAssociations[smfIP].Sessions[2].PDRs[1].Ipv4.String() != "1.1.1.1" {
		t.Errorf("Session 1, got broken")
	}
	if pfcpConn.NodeAssociations[smfIP].Sessions[3].PDRs[1].Teid != 0 {
		t.Errorf("Session 2, got broken")
	}

	// Check that SDF filter is stored inside session
	err = CheckSdfFilterEquality(pfcpConn.NodeAssociations[smfIP].Sessions[2].PDRs[1].PdrInfo.SdfFilter, fd)
	if err != nil {
		t.Error(err)
	}
	err = CheckSdfFilterEquality(pfcpConn.NodeAssociations[smfIP].Sessions[2].PDRs[1].PdrInfo.SdfFilter, fd)
	if err != nil {
		t.Error(err)
	}
}

func TestSdfFilterStoreInvalid(t *testing.T) {

	pfcpConn, smfIP := SdfFilterStorePreSetup(t)

	ip1, _ := net.ResolveIPAddr("ip", "1.1.1.1")

	// Request with bad/unsuported SDF
	seReq1 := message.NewSessionEstablishmentRequest(0, 0,
		1, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(2, net.ParseIP(smfIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(1),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				ie.NewFTEID(0, 0, ip1.IP, nil, 0),
				// ie.NewUEIPAddress(2, ip2.IP.String(), "", 0, 0),
				ie.NewSDFFilter("deny out ip from 10.62.0.1 to 8.8.8.8/32", "", "", "", 0),
			),
		),
	)

	var err error
	_, err = HandlePfcpSessionEstablishmentRequest(&pfcpConn, seReq1, smfIP)
	if err != nil {
		t.Errorf("No error should appear while handling session establishment request. PDR with bad SDF should be skipped.")
	}

	// Check that session PDR wasn't stored
	if len(pfcpConn.NodeAssociations[smfIP].Sessions[2].PDRs) != 0 {
		t.Errorf("Session 1, PDR with bad SDF shouldn't be stored")
	}
}
