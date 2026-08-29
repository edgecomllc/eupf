package core

import (
	"testing"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
)

func TestBuildSessionReportUsageIEs(t *testing.T) {
	tests := []struct {
		name            string
		profile         PfcpProfile
		wantVendorCount int
		wantURSEQNVal   uint32
	}{
		{"Default", DefaultProfile{}, 0, 7},
		{"Huawei", testHuaweiProfile, 5, 42},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ies := buildSessionReportUsageIEs(tt.profile, 1, 42, 7, 100, 200)
			if len(ies) != 2 {
				t.Fatalf("expected 2 top-level IEs (ReportType + UsageReport), got %d", len(ies))
			}
			if ies[0].Type != ie.ReportType {
				t.Errorf("expected ReportType, got %d", ies[0].Type)
			}
			if ies[1].Type != ie.UsageReportWithinSessionReportRequest {
				t.Errorf("expected UsageReport, got %d", ies[1].Type)
			}
			// UsageReport is a grouped IE — check children
			childIEs := ies[1].ChildIEs
			// Find URSEQN
			var urseqnVal uint32
			vendorCount := 0
			for _, c := range childIEs {
				if c.Type == ie.URSEQN {
					v, _ := c.URSEQN()
					urseqnVal = v
				}
				if c.EnterpriseID == 2011 {
					vendorCount++
				}
			}
			if urseqnVal != tt.wantURSEQNVal {
				t.Errorf("URSEQN: expected %d, got %d", tt.wantURSEQNVal, urseqnVal)
			}
			if vendorCount != tt.wantVendorCount {
				t.Errorf("vendor IEs: expected %d, got %d", tt.wantVendorCount, vendorCount)
			}
		})
	}
}

func TestBuildSessionReportADCIEs(t *testing.T) {
	tests := []struct {
		name            string
		profile         PfcpProfile
		sdfFilter       string
		wantVendorCount int
		wantURSEQNVal   uint32
	}{
		{"Default/empty", DefaultProfile{}, "", 0, 7},
		{"Huawei/empty", testHuaweiProfile, "", 3, 42},
		{"Huawei/filter", testHuaweiProfile, "permit in 6", 4, 42},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ies := buildSessionReportADCIEs(tt.profile, 1, 42, 7, tt.sdfFilter)
			if len(ies) != 2 {
				t.Fatalf("expected 2 top-level IEs, got %d", len(ies))
			}
			childIEs := ies[1].ChildIEs
			var urseqnVal uint32
			vendorCount := 0
			for _, c := range childIEs {
				if c.Type == ie.URSEQN {
					v, _ := c.URSEQN()
					urseqnVal = v
				}
				if c.EnterpriseID == 2011 {
					vendorCount++
				}
			}
			if urseqnVal != tt.wantURSEQNVal {
				t.Errorf("URSEQN: expected %d, got %d", tt.wantURSEQNVal, urseqnVal)
			}
			if vendorCount != tt.wantVendorCount {
				t.Errorf("vendor IEs: expected %d, got %d", tt.wantVendorCount, vendorCount)
			}
		})
	}
}

func TestBuildSessionReportReleaseIEs(t *testing.T) {
	pdrList := []uint16{4, 5}
	tests := []struct {
		name         string
		profile      PfcpProfile
		wantLen      int
		wantVendorIE bool
	}{
		{"Default", DefaultProfile{}, 1, false},
		{"Huawei", testHuaweiProfile, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ies := buildSessionReportReleaseIEs(tt.profile, pdrList)
			if len(ies) != tt.wantLen {
				t.Fatalf("expected %d IEs, got %d", tt.wantLen, len(ies))
			}
			if ies[0].Type != ie.ReportType {
				t.Errorf("expected ReportType, got %d", ies[0].Type)
			}
			if tt.wantVendorIE {
				if ies[1].Type != 32799 {
					t.Errorf("expected vendor IE type 32799, got %d", ies[1].Type)
				}
				if ies[1].EnterpriseID != 2011 {
					t.Errorf("expected enterprise-id 2011, got %d", ies[1].EnterpriseID)
				}
			}
		})
	}
}

func TestBuildDefaultSetupRequestIEs(t *testing.T) {
	tests := []struct {
		name    string
		profile PfcpProfile
	}{
		{"Default", DefaultProfile{}},
		{"Huawei", testHuaweiProfile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ies := buildDefaultSetupRequestIEs(tt.profile, "10.0.0.1", time.Now(), []uint8{0, 0, 0})
			if len(ies) != 3 {
				t.Fatalf("expected 3 IEs, got %d", len(ies))
			}
			if ies[0].Type != ie.NodeID {
				t.Errorf("expected NodeID, got %d", ies[0].Type)
			}
			if ies[1].Type != ie.RecoveryTimeStamp {
				t.Errorf("expected RecoveryTimeStamp, got %d", ies[1].Type)
			}
			if ies[2].Type != ie.UPFunctionFeatures {
				t.Errorf("expected UPFunctionFeatures, got %d", ies[2].Type)
			}
			for _, item := range ies {
				if item.EnterpriseID == 2011 {
					t.Errorf("Default setup should have no vendor IEs, found type %d with enterprise-id 2011", item.Type)
				}
			}
		})
	}
}

func TestBuildSxaSetupRequestIEs(t *testing.T) {
	tests := []struct {
		name          string
		profile       PfcpProfile
		wantVendorIEs int
	}{
		{"Default", DefaultProfile{}, 6},
		{"Huawei", testHuaweiProfile, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ies, err := buildSxaSetupRequestIEs(tt.profile, "10.0.0.1", time.Now(), "127.0.0.1", "127.0.0.2")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(ies) != 9 {
				t.Fatalf("expected 9 IEs (NodeID+TS+Features+6 vendor), got %d", len(ies))
			}
			if ies[0].Type != ie.NodeID {
				t.Errorf("expected NodeID, got %d", ies[0].Type)
			}
			vendorTypes := []uint16{32787, 32803, 32806, 32857, 32900, 32901}
			vendorCount := 0
			for _, item := range ies {
				if item.EnterpriseID == 2011 {
					vendorCount++
				}
			}
			if vendorCount != tt.wantVendorIEs {
				t.Errorf("expected %d vendor IEs, got %d", tt.wantVendorIEs, vendorCount)
			}
			for i, vt := range vendorTypes {
				found := false
				for _, item := range ies {
					if item.Type == vt && item.EnterpriseID == 2011 {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected vendor IE type %d not found (index %d)", vt, i)
				}
			}
		})
	}
}

func TestBuildSxbSetupRequestIEs(t *testing.T) {
	tests := []struct {
		name          string
		profile       PfcpProfile
		wantVendorIEs int
	}{
		{"Default", DefaultProfile{}, 6},
		{"Huawei", testHuaweiProfile, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ies, err := buildSxbSetupRequestIEs(tt.profile, "10.0.0.1", time.Now(), "127.0.0.3")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(ies) != 9 {
				t.Fatalf("expected 9 IEs, got %d", len(ies))
			}
			if ies[0].Type != ie.NodeID {
				t.Errorf("expected NodeID, got %d", ies[0].Type)
			}
			vendorCount := 0
			for _, item := range ies {
				if item.EnterpriseID == 2011 {
					vendorCount++
				}
			}
			if vendorCount != tt.wantVendorIEs {
				t.Errorf("expected %d vendor IEs, got %d", tt.wantVendorIEs, vendorCount)
			}
		})
	}
}

func TestBuildSxaSetupRequestIEsInvalidIP(t *testing.T) {
	_, err := buildSxaSetupRequestIEs(testHuaweiProfile, "10.0.0.1", time.Now(), "invalid", "127.0.0.2")
	if err == nil {
		t.Error("expected error for invalid s1u address, got nil")
	}
}

func TestBuildSxbSetupRequestIEsInvalidIP(t *testing.T) {
	_, err := buildSxbSetupRequestIEs(testHuaweiProfile, "10.0.0.1", time.Now(), "invalid")
	if err == nil {
		t.Error("expected error for invalid pa address, got nil")
	}
}
