package core

import (
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
