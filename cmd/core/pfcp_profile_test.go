package core

import (
	"fmt"
	"testing"

	"github.com/wmnsk/go-pfcp/ie"
)

// testHuaweiProfile is a shared HuaweiProfile fixture for tests that don't
// need address-specific assertions.
var testHuaweiProfile = HuaweiProfile{
	S1UAddress:  "127.0.0.1",
	S5S8Address: "127.0.0.2",
	PAAddress:   "127.0.0.3",
}

// testProfiles returns both profile implementations for table-driven tests
// that need to exercise Default and Huawei paths in parallel.
func testProfiles() []struct {
	name    string
	profile PfcpProfile
} {
	return []struct {
		name    string
		profile PfcpProfile
	}{
		{"Default", DefaultProfile{}},
		{"Huawei", testHuaweiProfile},
	}
}

func TestProfileInterfaceCompliance(t *testing.T) {
	var _ PfcpProfile = DefaultProfile{}
	var _ PfcpProfile = HuaweiProfile{}
}

func TestProfileConnectors(t *testing.T) {
	tests := []struct {
		name     string
		profile  PfcpProfile
		method   func(PfcpProfile, string) (AssociationConnector, error)
		wantType string
	}{
		{
			name:     "Default/Sxa",
			profile:  DefaultProfile{},
			method:   func(p PfcpProfile, n string) (AssociationConnector, error) { return p.SxaConnector(n) },
			wantType: "*core.DefaultAssociationConnector",
		},
		{
			name:     "Default/Sxb",
			profile:  DefaultProfile{},
			method:   func(p PfcpProfile, n string) (AssociationConnector, error) { return p.SxbConnector(n) },
			wantType: "*core.DefaultAssociationConnector",
		},
		{
			name:     "Default/N4",
			profile:  DefaultProfile{},
			method:   func(p PfcpProfile, n string) (AssociationConnector, error) { return p.N4Connector(n) },
			wantType: "*core.DefaultAssociationConnector",
		},
		{
			name:     "Huawei/Sxa",
			profile:  testHuaweiProfile,
			method:   func(p PfcpProfile, n string) (AssociationConnector, error) { return p.SxaConnector(n) },
			wantType: "*core.SxaAssociationConnector",
		},
		{
			name:     "Huawei/Sxb",
			profile:  testHuaweiProfile,
			method:   func(p PfcpProfile, n string) (AssociationConnector, error) { return p.SxbConnector(n) },
			wantType: "*core.SxbAssociationConnector",
		},
		{
			name:     "Huawei/N4",
			profile:  testHuaweiProfile,
			method:   func(p PfcpProfile, n string) (AssociationConnector, error) { return p.N4Connector(n) },
			wantType: "*core.DefaultAssociationConnector",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector, err := tt.method(tt.profile, "127.0.0.10")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if connector == nil {
				t.Fatal("expected non-nil connector")
			}
			if got := connector.getAddress(); got == "" {
				t.Error("expected non-empty address")
			}
			gotType := fmt.Sprintf("%T", connector)
			if gotType != tt.wantType {
				t.Errorf("expected %s, got %s", tt.wantType, gotType)
			}
		})
	}
}

func TestProfileNodeID(t *testing.T) {
	nodeIDStr := "127.0.0.1"
	for _, tt := range testProfiles() {
		t.Run(tt.name, func(t *testing.T) {
			nodeID := tt.profile.NodeID(nodeIDStr)
			if nodeID == nil {
				t.Fatal("expected non-nil NodeID IE")
			}
			v, err := nodeID.NodeID()
			if err != nil {
				t.Fatalf("unexpected error parsing NodeID: %v", err)
			}
			if v != nodeIDStr {
				t.Errorf("expected %s, got %s", nodeIDStr, v)
			}
		})
	}
}

func TestProfileURSEQN(t *testing.T) {
	sessionSeq := uint32(42)
	reportSeq := uint32(7)
	tests := []struct {
		name    string
		profile PfcpProfile
		wantSeq uint32
	}{
		{"Default", DefaultProfile{}, reportSeq},
		{"Huawei", testHuaweiProfile, sessionSeq},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urseqn := tt.profile.URSEQN(sessionSeq, reportSeq)
			if urseqn == nil {
				t.Fatal("expected non-nil URSEQN IE")
			}
			got, err := urseqn.URSEQN()
			if err != nil {
				t.Fatalf("unexpected error parsing URSEQN: %v", err)
			}
			if got != tt.wantSeq {
				t.Errorf("expected %d, got %d", tt.wantSeq, got)
			}
		})
	}
}

func TestProfileParseSubscriberData(t *testing.T) {
	tests := []struct {
		name       string
		profile    PfcpProfile
		wantIMSI   string
		wantMSISDN string
	}{
		{"Default", DefaultProfile{}, "", ""},
		{"Huawei", testHuaweiProfile, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imsi, msisdn := tt.profile.ParseSubscriberData(nil)
			if imsi != tt.wantIMSI {
				t.Errorf("expected imsi=%q, got %q", tt.wantIMSI, imsi)
			}
			if msisdn != tt.wantMSISDN {
				t.Errorf("expected msisdn=%q, got %q", tt.wantMSISDN, msisdn)
			}
		})
	}
}

func TestProfileParseQFI(t *testing.T) {
	qer := ie.NewCreateQER(ie.NewQERID(1), ie.NewQFI(5))
	for _, tt := range testProfiles() {
		t.Run(tt.name, func(t *testing.T) {
			qfi, ok := tt.profile.ParseQFI(qer)
			if !ok {
				t.Fatal("expected QFI to be found")
			}
			if qfi != 5 {
				t.Errorf("expected QFI=5, got %d", qfi)
			}
		})
	}
}

func TestProfileParseOuterHeaderCreationDefault(t *testing.T) {
	// Standard go-pfcp OuterHeaderCreation IE — only DefaultProfile can parse it.
	oc := ie.NewOuterHeaderCreation(0x0100, 0x12345678, "10.0.0.1", "", 0, 0, 0)
	fields, err := DefaultProfile{}.ParseOuterHeaderCreation(oc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fields.Teid != 0x12345678 {
		t.Errorf("expected TEID=0x12345678, got 0x%08x", fields.Teid)
	}
	if fields.IPv4Address == nil {
		t.Error("expected non-nil IPv4Address")
	}
}

func TestHuaweiProfileParseOuterHeaderCreationHuaweiEncoded(t *testing.T) {
	// Huawei-encoded OuterHeaderCreation (raw bytes: desc=0x00, teid=4 bytes, ipv4=4 bytes).
	huaweiOCPayload := []byte{0x00, 0x74, 0x03, 0x30, 0x0b, 0x0a, 0xa9, 0x70, 0xae}
	huaweiOC := ie.New(ie.OuterHeaderCreation, huaweiOCPayload)
	fields, err := HuaweiProfile{}.ParseOuterHeaderCreation(huaweiOC)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fields.Teid == 0 {
		t.Error("expected non-zero TEID")
	}
}

func TestProfileHeartbeatRequestAdditionalIEs(t *testing.T) {
	tests := []struct {
		name     string
		profile  PfcpProfile
		wantLen  int
		wantType uint16
	}{
		{"Default", DefaultProfile{}, 0, 0},
		{"Huawei", testHuaweiProfile, 1, ie.Metric},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ies := tt.profile.HeartbeatRequestAdditionalIEs()
			if tt.wantLen == 0 {
				if ies != nil {
					t.Errorf("expected nil, got %d IEs", len(ies))
				}
				return
			}
			if len(ies) != tt.wantLen {
				t.Fatalf("expected %d IEs, got %d", tt.wantLen, len(ies))
			}
			if ies[0].Type != tt.wantType {
				t.Errorf("expected type %d, got %d", tt.wantType, ies[0].Type)
			}
		})
	}
}

func TestProfileSessionEstablishmentResponseAdditionalIEs(t *testing.T) {
	tests := []struct {
		name     string
		profile  PfcpProfile
		wantLen  int
		wantType uint16
	}{
		{"Default", DefaultProfile{}, 0, 0},
		{"Huawei", testHuaweiProfile, 1, ie.FSEID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ies := tt.profile.SessionEstablishmentResponseAdditionalIEs(42, nil)
			if tt.wantLen == 0 {
				if ies != nil {
					t.Errorf("expected nil, got %d IEs", len(ies))
				}
				return
			}
			if len(ies) != tt.wantLen {
				t.Fatalf("expected %d IEs, got %d", tt.wantLen, len(ies))
			}
			if ies[0].Type != tt.wantType {
				t.Errorf("expected type %d, got %d", tt.wantType, ies[0].Type)
			}
		})
	}
}

func TestProfileUsageReportDeletionVendorIEs(t *testing.T) {
	tests := []struct {
		name      string
		profile   PfcpProfile
		wantLen   int
		wantTypes []uint16
	}{
		{"Default", DefaultProfile{}, 0, nil},
		{"Huawei", testHuaweiProfile, 3, []uint16{34000, 32843, 34010}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ies := tt.profile.UsageReportDeletionVendorIEs()
			if tt.wantLen == 0 {
				if ies != nil {
					t.Errorf("expected nil, got %d IEs", len(ies))
				}
				return
			}
			if len(ies) != tt.wantLen {
				t.Fatalf("expected %d IEs, got %d", tt.wantLen, len(ies))
			}
			for i, et := range tt.wantTypes {
				if ies[i].Type != et {
					t.Errorf("IE[%d]: expected type %d, got %d", i, et, ies[i].Type)
				}
			}
		})
	}
}

func TestProfileUsageReportSessionReportVendorIEs(t *testing.T) {
	tests := []struct {
		name      string
		profile   PfcpProfile
		wantLen   int
		wantTypes []uint16
	}{
		{"Default", DefaultProfile{}, 0, nil},
		{"Huawei", testHuaweiProfile, 5, []uint16{34000, 32843, 34010, 34011, 34012}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ies := tt.profile.UsageReportSessionReportVendorIEs()
			if tt.wantLen == 0 {
				if ies != nil {
					t.Errorf("expected nil, got %d IEs", len(ies))
				}
				return
			}
			if len(ies) != tt.wantLen {
				t.Fatalf("expected %d IEs, got %d", tt.wantLen, len(ies))
			}
			for i, et := range tt.wantTypes {
				if ies[i].Type != et {
					t.Errorf("IE[%d]: expected type %d, got %d", i, et, ies[i].Type)
				}
			}
		})
	}
}

func TestProfileUsageReportADCVendorIEs(t *testing.T) {
	tests := []struct {
		name      string
		profile   PfcpProfile
		sdfFilter string
		wantLen   int
		wantTypes []uint16
	}{
		{"Default/empty", DefaultProfile{}, "", 0, nil},
		{"Default/filter", DefaultProfile{}, "permit in 6", 0, nil},
		{"Huawei/empty", testHuaweiProfile, "", 3, []uint16{34000, 32843, 36001}},
		{"Huawei/filter", testHuaweiProfile, "permit in 6", 4, []uint16{34000, 32843, 36001, 36017}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ies := tt.profile.UsageReportADCVendorIEs(tt.sdfFilter)
			if tt.wantLen == 0 {
				if ies != nil {
					t.Errorf("expected nil, got %d IEs", len(ies))
				}
				return
			}
			if len(ies) != tt.wantLen {
				t.Fatalf("expected %d IEs, got %d", tt.wantLen, len(ies))
			}
			for i, et := range tt.wantTypes {
				if ies[i].Type != et {
					t.Errorf("IE[%d]: expected type %d, got %d", i, et, ies[i].Type)
				}
			}
		})
	}
}

func TestProfileSessionReportReleaseVendorIE(t *testing.T) {
	pdrList := []uint16{4, 5}
	tests := []struct {
		name         string
		profile      PfcpProfile
		wantNil      bool
		wantType     uint16
		wantVendorID uint16
	}{
		{"Default", DefaultProfile{}, true, 0, 0},
		{"Huawei", testHuaweiProfile, false, 32799, 2011},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendorIE := tt.profile.SessionReportReleaseVendorIE(pdrList)
			if tt.wantNil {
				if vendorIE != nil {
					t.Errorf("expected nil, got non-nil IE")
				}
				return
			}
			if vendorIE == nil {
				t.Fatal("expected non-nil IE")
			}
			if vendorIE.Type != tt.wantType {
				t.Errorf("expected type %d, got %d", tt.wantType, vendorIE.Type)
			}
			if vendorIE.EnterpriseID != tt.wantVendorID {
				t.Errorf("expected enterprise-id %d, got %d", tt.wantVendorID, vendorIE.EnterpriseID)
			}
		})
	}
}

func TestHuaweiProfileSessionReportReleaseVendorIEEmptyPdrList(t *testing.T) {
	p := testHuaweiProfile
	vendorIE := p.SessionReportReleaseVendorIE([]uint16{})
	if vendorIE == nil {
		t.Fatal("HuaweiProfile should return non-nil IE even for empty pdrList")
	}
	if vendorIE.Type != 32799 {
		t.Errorf("expected type 32799, got %d", vendorIE.Type)
	}
}
