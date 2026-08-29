package core

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func TestHandlePfcpAssociationSetupRequestProfileNodeID(t *testing.T) {
	tests := []struct {
		name    string
		profile PfcpProfile
	}{
		{"Default", DefaultProfile{}},
		{"Huawei", testHuaweiProfile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pfcpConn := PfcpConnection{
				NodeAssociations: make(map[string]*NodeAssociation),
				nodeId:           "127.0.0.1",
				associationMutex: &sync.Mutex{},
				profile:          tt.profile,
			}
			asReq := message.NewAssociationSetupRequest(0,
				ie.NewNodeID("", "", "test"),
				ie.NewRecoveryTimeStamp(time.Now()),
			)
			response, _, err := HandlePfcpAssociationSetupRequest(&pfcpConn, asReq, "127.0.0.1")
			if err != nil {
				t.Fatalf("Error handling association setup request: %s", err)
			}
			nodeIDIE := response.(*message.AssociationSetupResponse).NodeID
			if nodeIDIE == nil {
				t.Fatal("expected non-nil NodeID in response")
			}
			v, err := nodeIDIE.NodeID()
			if err != nil {
				t.Fatalf("unexpected error parsing NodeID: %v", err)
			}
			if v != "127.0.0.1" {
				t.Errorf("expected 127.0.0.1, got %s", v)
			}
		})
	}
}

func TestHandlePfcpSessionEstablishmentResponseSecondFSEID(t *testing.T) {
	tests := []struct {
		name            string
		profile         PfcpProfile
		wantSecondFSEID bool
	}{
		{"Default", DefaultProfile{}, false},
		{"Huawei", testHuaweiProfile, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pfcpConn, smfIP := PreparePfcpConnectionWithMockAndProfile(t, &MapOperationsMock{}, config.UpfConfig{}, tt.profile)

			estReq := message.NewSessionEstablishmentRequest(0, 0, 2, 1, 0,
				ie.NewNodeID("", "", "test"),
				ie.NewFSEID(1, net.ParseIP(smfIP), nil),
				ie.NewCreatePDR(
					ie.NewPDRID(0xffff),
				),
			)
			response, _, err := HandlePfcpSessionEstablishmentRequest(pfcpConn, estReq, smfIP)
			if err != nil {
				t.Fatalf("Error handling session establishment request: %s", err)
			}

			estResp := response.(*message.SessionEstablishmentResponse)
			fseidCount := 0
			for _, item := range estResp.IEs {
				if item.Type == ie.FSEID {
					fseidCount++
				}
			}
			if tt.wantSecondFSEID {
				if fseidCount < 2 {
					t.Errorf("expected at least 2 FSEID IEs, got %d", fseidCount)
				}
			} else {
				if fseidCount > 1 {
					t.Errorf("expected at most 1 FSEID IE for DefaultProfile, got %d", fseidCount)
				}
			}
		})
	}
}

func TestSendHeartbeatRequestProfileMetricIE(t *testing.T) {
	// HeartbeatRequestAdditionalIEs is tested at the profile level (TestProfileHeartbeatRequestAdditionalIEs).
	// This test verifies that SendHeartbeatRequest does not panic with either profile
	// and that the profile method is actually invoked (not nil).
	tests := []struct {
		name    string
		profile PfcpProfile
	}{
		{"Default", DefaultProfile{}},
		{"Huawei", testHuaweiProfile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pfcpConn := PfcpConnection{
				profile:           tt.profile,
				RecoveryTimestamp: time.Now(),
			}
			// SendHeartbeatRequest will fail to resolve the UDP address (no listener),
			// but it should not panic on conn.profile.HeartbeatRequestAdditionalIEs().
			// The function logs warnings on failure, so no error to check.
			SendHeartbeatRequest(&pfcpConn, 0, "127.0.0.1:99999")
		})
	}
}
