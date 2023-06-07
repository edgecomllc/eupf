package main

import (
	"net"
	"testing"

	"github.com/wmnsk/go-pfcp/ie"
)

type MockController struct {
}

func (mock *MockController) PutPdrUpLink(teid uint32, pdrInfo PdrInfo) error {
	return nil
}
func (mock *MockController) PutPdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error {
	return nil
}
func (mock *MockController) UpdatePdrUpLink(teid uint32, pdrInfo PdrInfo) error {
	return nil
}
func (mock *MockController) UpdatePdrDownLink(ipv4 net.IP, pdrInfo PdrInfo) error {
	return nil
}
func (mock *MockController) DeletePdrUpLink(teid uint32) error {
	return nil
}
func (mock *MockController) DeletePdrDownLink(ipv4 net.IP) error {
	return nil
}
func (mock *MockController) NewFar(farInfo FarInfo) (uint32, error) {
	return 0, nil
}
func (mock *MockController) UpdateFar(internalId uint32, farInfo FarInfo) error {
	return nil
}
func (mock *MockController) DeleteFar(internalId uint32) error {
	return nil
}
func (mock *MockController) NewQer(qerInfo QerInfo) (uint32, error) {
	return 0, nil
}
func (mock *MockController) UpdateQer(internalId uint32, qerInfo QerInfo) error {
	return nil
}
func (mock *MockController) DeleteQer(internalId uint32) error {
	return nil
}

func Test_applyDownlinkPDR(t *testing.T) {
	type args struct {
		pdi           []*ie.IE
		spdrInfo      SPDRInfo
		pdrId         uint16
		session       *Session
		mapOperations ForwardingPlaneController
	}

	sessions := []Session{
		{
			LocalSEID:    1,
			RemoteSEID:   2,
			UplinkPDRs:   map[uint32]SPDRInfo{},
			DownlinkPDRs: map[uint32]SPDRInfo{},
			FARs:         map[uint32]SFarInfo{},
			QERs:         map[uint32]SQerInfo{},
		},
		{
			LocalSEID:    3,
			RemoteSEID:   4,
			UplinkPDRs:   map[uint32]SPDRInfo{},
			DownlinkPDRs: map[uint32]SPDRInfo{},
			FARs:         map[uint32]SFarInfo{},
			QERs:         map[uint32]SQerInfo{},
		},
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "First",
			args: struct {
				pdi           []*ie.IE
				spdrInfo      SPDRInfo
				pdrId         uint16
				session       *Session
				mapOperations ForwardingPlaneController
			}{
				pdi: ie.NewPDI(
					ie.NewUEIPAddress(0x02, "1.1.1.1", "", 0, 0),
				).ChildIEs,
				spdrInfo:      SPDRInfo{},
				pdrId:         0,
				session:       &sessions[0],
				mapOperations: &MockController{},
			},
			wantErr: false,
		},
		{
			name: "Second",
			args: struct {
				pdi           []*ie.IE
				spdrInfo      SPDRInfo
				pdrId         uint16
				session       *Session
				mapOperations ForwardingPlaneController
			}{
				pdi: ie.NewPDI(
					ie.NewUEIPAddress(0x02, "2.2.2.2", "", 0, 0),
				).ChildIEs,
				spdrInfo:      SPDRInfo{},
				pdrId:         0,
				session:       &sessions[1],
				mapOperations: &MockController{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := applyDownlinkPDR(tt.args.pdi, tt.args.spdrInfo, tt.args.pdrId, tt.args.session, tt.args.mapOperations); (err != nil) != tt.wantErr {
				t.Errorf("applyDownlinkPDR() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Log("The end")
}
