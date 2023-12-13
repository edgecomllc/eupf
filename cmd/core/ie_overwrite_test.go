package core

import (
	"net"
	"testing"

	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type MapOperationsMock struct {
}

func (mapOps *MapOperationsMock) PutPdrUplink(teid uint32, pdrInfo ebpf.PdrInfo) error {
	return nil
}
func (mapOps *MapOperationsMock) PutPdrDownlink(ipv4 net.IP, pdrInfo ebpf.PdrInfo) error {
	return nil
}
func (mapOps *MapOperationsMock) UpdatePdrUplink(teid uint32, pdrInfo ebpf.PdrInfo) error {
	return nil
}
func (mapOps *MapOperationsMock) UpdatePdrDownlink(ipv4 net.IP, pdrInfo ebpf.PdrInfo) error {
	return nil
}
func (mapOps *MapOperationsMock) DeletePdrUplink(teid uint32) error {
	return nil
}
func (mapOps *MapOperationsMock) DeletePdrDownlink(ipv4 net.IP) error {
	return nil
}
func (mapOps *MapOperationsMock) PutDownlinkPdrIp6(ipv6 net.IP, pdrInfo ebpf.PdrInfo) error {
	return nil
}
func (mapOps *MapOperationsMock) UpdateDownlinkPdrIp6(ipv6 net.IP, pdrInfo ebpf.PdrInfo) error {
	return nil
}
func (mapOps *MapOperationsMock) DeleteDownlinkPdrIp6(ipv6 net.IP) error {
	return nil
}
func (mapOps *MapOperationsMock) NewFar(farInfo ebpf.FarInfo) (uint32, error) {
	return 0, nil
}
func (mapOps *MapOperationsMock) UpdateFar(internalId uint32, farInfo ebpf.FarInfo) error {
	return nil
}
func (mapOps *MapOperationsMock) DeleteFar(internalId uint32) error {
	return nil
}
func (mapOps *MapOperationsMock) NewQer(qerInfo ebpf.QerInfo) (uint32, error) {
	return 0, nil
}
func (mapOps *MapOperationsMock) UpdateQer(internalId uint32, qerInfo ebpf.QerInfo) error {
	return nil
}
func (mapOps *MapOperationsMock) DeleteQer(internalId uint32) error {
	return nil
}

func TestSessionOverwrite(t *testing.T) {

	mapOps := MapOperationsMock{}
	// Create pfcp connection struct
	pfcpConn := PfcpConnection{
		NodeAssociations: make(map[string]*NodeAssociation),
		nodeId:           "test-node",
		mapOperations:    &mapOps,
		n3Address:        net.ParseIP("127.0.0.1"),
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

	// Create two send two Session Establishment Requests with downlink PDRs
	// and check that the first session is not overwritten
	seReq1 := message.NewSessionEstablishmentRequest(0, 0,
		1, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, net.ParseIP(remoteIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(1),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				//ie.NewFTEID(0, 0, ip1.IP, nil, 0),
				ie.NewUEIPAddress(2, "1.1.1.1", "", 0, 0),
			),
		),
	)

	seReq2 := message.NewSessionEstablishmentRequest(0, 0,
		2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(2, net.ParseIP(remoteIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(1),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				//ie.NewFTEID(0, 0, ip2.IP, nil, 0),
				ie.NewUEIPAddress(2, "2.2.2.2", "", 0, 0),
			),
		),
	)

	// Send first request
	_, err = HandlePfcpSessionEstablishmentRequest(&pfcpConn, seReq1, remoteIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	// Send second request
	_, err = HandlePfcpSessionEstablishmentRequest(&pfcpConn, seReq2, remoteIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	// Check that session PDRs are correct
	if pfcpConn.NodeAssociations[remoteIP].Sessions[2].PDRs[1].Ipv4.String() != "1.1.1.1" {
		t.Errorf("Session 1, got broken")
	}
	if pfcpConn.NodeAssociations[remoteIP].Sessions[3].PDRs[1].Ipv4.String() != "2.2.2.2" {
		t.Errorf("Session 2, got broken")
	}

	// Send Session Modification Request, create FAR
	smReq := message.NewSessionModificationRequest(0, 0,
		2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(2, net.ParseIP(remoteIP), nil),
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
	_, err = HandlePfcpSessionModificationRequest(&pfcpConn, smReq, remoteIP)
	if err != nil {
		t.Errorf("Error handling session modification request: %s", err)
	}

	// Check that session PDRs are correct
	if pfcpConn.NodeAssociations[remoteIP].Sessions[2].PDRs[1].Ipv4.String() != "1.1.1.1" {
		t.Errorf("Session 1, got broken")
	}
	if pfcpConn.NodeAssociations[remoteIP].Sessions[3].PDRs[1].Ipv4.String() != "2.2.2.2" {
		t.Errorf("Session 2, got broken")
	}
}
