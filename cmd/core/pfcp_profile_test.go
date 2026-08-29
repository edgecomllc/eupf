package core

import (
	"fmt"
	"testing"

	"github.com/wmnsk/go-pfcp/ie"
)

func TestDefaultProfileSxaConnector(t *testing.T) {
	p := DefaultProfile{}
	connector, err := p.SxaConnector("127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connector == nil {
		t.Fatal("expected non-nil connector")
	}
	if got := connector.getAddress(); got == "" {
		t.Error("expected non-empty address")
	}
}

func TestDefaultProfileSxbConnector(t *testing.T) {
	p := DefaultProfile{}
	connector, err := p.SxbConnector("127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connector == nil {
		t.Fatal("expected non-nil connector")
	}
}

func TestDefaultProfileN4Connector(t *testing.T) {
	p := DefaultProfile{}
	connector, err := p.N4Connector("127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connector == nil {
		t.Fatal("expected non-nil connector")
	}
	switch connector.(type) {
	case *DefaultAssociationConnector:
	default:
		t.Errorf("DefaultProfile.N4Connector should return *DefaultAssociationConnector, got %T", connector)
	}
}

func TestDefaultProfileNodeID(t *testing.T) {
	p := DefaultProfile{}
	nodeID := p.NodeID("127.0.0.1")
	if nodeID == nil {
		t.Fatal("expected non-nil NodeID IE")
	}
	v, err := nodeID.NodeID()
	if err != nil {
		t.Fatalf("unexpected error parsing NodeID: %v", err)
	}
	if v != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %s", v)
	}
}

func TestDefaultProfileURSEQN(t *testing.T) {
	p := DefaultProfile{}
	sessionSeq := uint32(42)
	reportSeq := uint32(7)
	nodeID := p.URSEQN(sessionSeq, reportSeq)
	if nodeID == nil {
		t.Fatal("expected non-nil URSEQN IE")
	}
	got, err := nodeID.URSEQN()
	if err != nil {
		t.Fatalf("unexpected error parsing URSEQN: %v", err)
	}
	if got != reportSeq {
		t.Errorf("DefaultProfile should use reportSeq: expected %d, got %d", reportSeq, got)
	}
}

func TestDefaultProfileParseSubscriberData(t *testing.T) {
	p := DefaultProfile{}
	imsi, msisdn := p.ParseSubscriberData(nil)
	if imsi != "" || msisdn != "" {
		t.Errorf("DefaultProfile should return empty subscriber data, got imsi=%s msisdn=%s", imsi, msisdn)
	}
}

func TestDefaultProfileParseQFI(t *testing.T) {
	p := DefaultProfile{}
	qer := ie.NewCreateQER(ie.NewQERID(1), ie.NewQFI(5))
	qfi, ok := p.ParseQFI(qer)
	if !ok {
		t.Fatal("expected QFI to be found")
	}
	if qfi != 5 {
		t.Errorf("expected QFI=5, got %d", qfi)
	}
}

func TestHuaweiProfileSxaConnector(t *testing.T) {
	p := HuaweiProfile{S1UAddress: "127.0.0.1", S5S8Address: "127.0.0.2", PAAddress: "127.0.0.3"}
	connector, err := p.SxaConnector("127.0.0.10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connector == nil {
		t.Fatal("expected non-nil connector")
	}
	if got := connector.getAddress(); got == "" {
		t.Error("expected non-empty address")
	}
}

func TestHuaweiProfileSxbConnector(t *testing.T) {
	p := HuaweiProfile{S1UAddress: "127.0.0.1", S5S8Address: "127.0.0.2", PAAddress: "127.0.0.3"}
	connector, err := p.SxbConnector("127.0.0.10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connector == nil {
		t.Fatal("expected non-nil connector")
	}
}

func TestHuaweiProfileN4Connector(t *testing.T) {
	p := HuaweiProfile{S1UAddress: "127.0.0.1", S5S8Address: "127.0.0.2", PAAddress: "127.0.0.3"}
	connector, err := p.N4Connector("127.0.0.10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connector == nil {
		t.Fatal("expected non-nil connector")
	}
	switch connector.(type) {
	case *DefaultAssociationConnector:
	default:
		t.Errorf("HuaweiProfile.N4Connector should return *DefaultAssociationConnector, got %T", connector)
	}
}

func TestProfileN4ConnectorEquivalence(t *testing.T) {
	dp := DefaultProfile{}
	hp := HuaweiProfile{S1UAddress: "127.0.0.1", S5S8Address: "127.0.0.2", PAAddress: "127.0.0.3"}

	defaultConn, err := dp.N4Connector("127.0.0.10")
	if err != nil {
		t.Fatalf("DefaultProfile.N4Connector error: %v", err)
	}
	huaweiConn, err := hp.N4Connector("127.0.0.10")
	if err != nil {
		t.Fatalf("HuaweiProfile.N4Connector error: %v", err)
	}

	defaultType := fmt.Sprintf("%T", defaultConn)
	huaweiType := fmt.Sprintf("%T", huaweiConn)
	if defaultType != huaweiType {
		t.Errorf("N4Connector should return same connector type for both profiles: default=%s huawei=%s", defaultType, huaweiType)
	}
}

func TestHuaweiProfileNodeID(t *testing.T) {
	p := HuaweiProfile{}
	nodeID := p.NodeID("127.0.0.1")
	if nodeID == nil {
		t.Fatal("expected non-nil NodeID IE")
	}
}

func TestHuaweiProfileURSEQN(t *testing.T) {
	p := HuaweiProfile{}
	sessionSeq := uint32(42)
	reportSeq := uint32(7)
	nodeID := p.URSEQN(sessionSeq, reportSeq)
	if nodeID == nil {
		t.Fatal("expected non-nil URSEQN IE")
	}
	got, err := nodeID.URSEQN()
	if err != nil {
		t.Fatalf("unexpected error parsing URSEQN: %v", err)
	}
	if got != sessionSeq {
		t.Errorf("HuaweiProfile should use sessionSeq: expected %d, got %d", sessionSeq, got)
	}
}

func TestHuaweiProfileParseSubscriberDataEmpty(t *testing.T) {
	p := HuaweiProfile{}
	imsi, msisdn := p.ParseSubscriberData(nil)
	if imsi != "" || msisdn != "" {
		t.Errorf("expected empty subscriber data for nil IE array, got imsi=%s msisdn=%s", imsi, msisdn)
	}
}

func TestHuaweiProfileParseQFIStandard(t *testing.T) {
	p := HuaweiProfile{}
	qer := ie.NewCreateQER(ie.NewQERID(1), ie.NewQFI(5))
	qfi, ok := p.ParseQFI(qer)
	if !ok {
		t.Fatal("expected QFI to be found via standard QFI IE")
	}
	if qfi != 5 {
		t.Errorf("expected QFI=5, got %d", qfi)
	}
}

func TestProfileNodeIDDifference(t *testing.T) {
	dp := DefaultProfile{}
	hp := HuaweiProfile{}
	nodeID := "10.0.0.1"

	defaultIE := dp.NodeID(nodeID)
	huaweiIE := hp.NodeID(nodeID)

	defaultNodeID, _ := defaultIE.NodeID()
	huaweiNodeID, _ := huaweiIE.NodeID()

	if defaultNodeID != huaweiNodeID {
		t.Errorf("both profiles should parse to same NodeID value: default=%s huawei=%s", defaultNodeID, huaweiNodeID)
	}

}

func TestProfileURSEQNDifference(t *testing.T) {
	dp := DefaultProfile{}
	hp := HuaweiProfile{}
	sessionSeq := uint32(100)
	reportSeq := uint32(200)

	defaultURSEQN, _ := dp.URSEQN(sessionSeq, reportSeq).URSEQN()
	huaweiURSEQN, _ := hp.URSEQN(sessionSeq, reportSeq).URSEQN()

	if defaultURSEQN != reportSeq {
		t.Errorf("DefaultProfile should use reportSeq (%d), got %d", reportSeq, defaultURSEQN)
	}
	if huaweiURSEQN != sessionSeq {
		t.Errorf("HuaweiProfile should use sessionSeq (%d), got %d", sessionSeq, huaweiURSEQN)
	}
}

func TestProfileParseOuterHeaderCreationStandard(t *testing.T) {
	dp := DefaultProfile{}
	oc := ie.NewOuterHeaderCreation(0x0100, 0x12345678, "10.0.0.1", "", 0, 0, 0)
	fields, err := dp.ParseOuterHeaderCreation(oc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fields.OuterHeaderCreation != 0x01 {
		t.Errorf("expected OuterHeaderCreation=0x01, got 0x%02x", fields.OuterHeaderCreation)
	}
	if fields.Teid != 0x12345678 {
		t.Errorf("expected TEID=0x12345678, got 0x%08x", fields.Teid)
	}
}

func TestProfileSxaConnectorTypeDifference(t *testing.T) {
	dp := DefaultProfile{}
	hp := HuaweiProfile{S1UAddress: "127.0.0.1", S5S8Address: "127.0.0.2"}

	defaultConn, _ := dp.SxaConnector("127.0.0.10")
	huaweiConn, _ := hp.SxaConnector("127.0.0.10")

	switch defaultConn.(type) {
	case *DefaultAssociationConnector:
	default:
		t.Errorf("DefaultProfile should return *DefaultAssociationConnector, got %T", defaultConn)
	}

	switch huaweiConn.(type) {
	case *SxaAssociationConnector:
	default:
		t.Errorf("HuaweiProfile should return *SxaAssociationConnector, got %T", huaweiConn)
	}
}

func TestProfileInterfaceCompliance(t *testing.T) {
	var _ PfcpProfile = DefaultProfile{}
	var _ PfcpProfile = HuaweiProfile{}
}

func TestDefaultProfileHeartbeatRequestAdditionalIEs(t *testing.T) {
	p := DefaultProfile{}
	ies := p.HeartbeatRequestAdditionalIEs()
	if ies != nil {
		t.Errorf("DefaultProfile should return nil additional IEs, got %d", len(ies))
	}
}

func TestHuaweiProfileHeartbeatRequestAdditionalIEs(t *testing.T) {
	p := HuaweiProfile{}
	ies := p.HeartbeatRequestAdditionalIEs()
	if len(ies) != 1 {
		t.Fatalf("HuaweiProfile should return 1 additional IE, got %d", len(ies))
	}
	if ies[0] == nil {
		t.Fatal("expected non-nil IE")
	}
	if ies[0].Type != ie.Metric {
		t.Errorf("expected Metric IE (type %d), got type %d", ie.Metric, ies[0].Type)
	}
}

func TestDefaultProfileSessionEstablishmentResponseAdditionalIEs(t *testing.T) {
	p := DefaultProfile{}
	ies := p.SessionEstablishmentResponseAdditionalIEs(42, nil)
	if ies != nil {
		t.Errorf("DefaultProfile should return nil, got %d IEs", len(ies))
	}
}

func TestHuaweiProfileSessionEstablishmentResponseAdditionalIEs(t *testing.T) {
	p := HuaweiProfile{}
	ies := p.SessionEstablishmentResponseAdditionalIEs(42, nil)
	if len(ies) != 1 {
		t.Fatalf("HuaweiProfile should return 1 IE, got %d", len(ies))
	}
	if ies[0] == nil {
		t.Fatal("expected non-nil IE")
	}
	if ies[0].Type != ie.FSEID {
		t.Errorf("expected FSEID IE, got type %d", ies[0].Type)
	}
}

func TestDefaultProfileUsageReportDeletionVendorIEs(t *testing.T) {
	p := DefaultProfile{}
	if ies := p.UsageReportDeletionVendorIEs(); ies != nil {
		t.Errorf("DefaultProfile should return nil, got %d IEs", len(ies))
	}
}

func TestDefaultProfileUsageReportSessionReportVendorIEs(t *testing.T) {
	p := DefaultProfile{}
	if ies := p.UsageReportSessionReportVendorIEs(); ies != nil {
		t.Errorf("DefaultProfile should return nil, got %d IEs", len(ies))
	}
}

func TestDefaultProfileUsageReportADCVendorIEs(t *testing.T) {
	p := DefaultProfile{}
	if ies := p.UsageReportADCVendorIEs("filter"); ies != nil {
		t.Errorf("DefaultProfile should return nil, got %d IEs", len(ies))
	}
}

func TestHuaweiProfileUsageReportDeletionVendorIEs(t *testing.T) {
	p := HuaweiProfile{}
	ies := p.UsageReportDeletionVendorIEs()
	if len(ies) != 3 {
		t.Fatalf("HuaweiProfile should return 3 IEs, got %d", len(ies))
	}
	expectedTypes := []uint16{34000, 32843, 34010}
	for i, et := range expectedTypes {
		if ies[i].Type != et {
			t.Errorf("IE[%d]: expected type %d, got %d", i, et, ies[i].Type)
		}
	}
}

func TestHuaweiProfileUsageReportSessionReportVendorIEs(t *testing.T) {
	p := HuaweiProfile{}
	ies := p.UsageReportSessionReportVendorIEs()
	if len(ies) != 5 {
		t.Fatalf("HuaweiProfile should return 5 IEs, got %d", len(ies))
	}
	expectedTypes := []uint16{34000, 32843, 34010, 34011, 34012}
	for i, et := range expectedTypes {
		if ies[i].Type != et {
			t.Errorf("IE[%d]: expected type %d, got %d", i, et, ies[i].Type)
		}
	}
}

func TestHuaweiProfileUsageReportADCVendorIEsWithoutSdfFilter(t *testing.T) {
	p := HuaweiProfile{}
	ies := p.UsageReportADCVendorIEs("")
	if len(ies) != 3 {
		t.Fatalf("HuaweiProfile should return 3 IEs without sdfFilter, got %d", len(ies))
	}
	expectedTypes := []uint16{34000, 32843, 36001}
	for i, et := range expectedTypes {
		if ies[i].Type != et {
			t.Errorf("IE[%d]: expected type %d, got %d", i, et, ies[i].Type)
		}
	}
}

func TestHuaweiProfileUsageReportADCVendorIEsWithSdfFilter(t *testing.T) {
	p := HuaweiProfile{}
	ies := p.UsageReportADCVendorIEs("permit in 6 from 100.89.2.1/32 45128 to 10.169.20.178/32 10650")
	if len(ies) != 4 {
		t.Fatalf("HuaweiProfile should return 4 IEs with sdfFilter, got %d", len(ies))
	}
	if ies[3].Type != 36017 {
		t.Errorf("IE[3]: expected type 36017, got %d", ies[3].Type)
	}
}

func TestDefaultProfileSessionReportReleaseVendorIE(t *testing.T) {
	p := DefaultProfile{}
	if ie := p.SessionReportReleaseVendorIE([]uint16{1, 2}); ie != nil {
		t.Errorf("DefaultProfile should return nil, got non-nil IE")
	}
}

func TestHuaweiProfileSessionReportReleaseVendorIE(t *testing.T) {
	p := HuaweiProfile{}
	vendorIE := p.SessionReportReleaseVendorIE([]uint16{4, 5})
	if vendorIE == nil {
		t.Fatal("HuaweiProfile should return non-nil IE")
	}
	if vendorIE.Type != 32799 {
		t.Errorf("expected type 32799, got %d", vendorIE.Type)
	}
	if vendorIE.EnterpriseID != 2011 {
		t.Errorf("expected enterprise-id 2011, got %d", vendorIE.EnterpriseID)
	}
}

func TestHuaweiProfileSessionReportReleaseVendorIEEmptyPdrList(t *testing.T) {
	p := HuaweiProfile{}
	vendorIE := p.SessionReportReleaseVendorIE([]uint16{})
	if vendorIE == nil {
		t.Fatal("HuaweiProfile should return non-nil IE even for empty pdrList")
	}
	if vendorIE.Type != 32799 {
		t.Errorf("expected type 32799, got %d", vendorIE.Type)
	}
}
