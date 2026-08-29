package core

import (
	"testing"

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
