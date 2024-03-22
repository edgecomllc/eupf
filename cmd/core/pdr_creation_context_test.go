package core

import (
	"testing"

	"github.com/edgecomllc/eupf/cmd/core/service"
	"github.com/wmnsk/go-pfcp/ie"
)

func TestPDRCreationContext_extractPDR(t *testing.T) {
	type fields struct {
		Session         *Session
		ResourceManager *service.ResourceManager
		TEIDCache       map[uint8]uint32
	}
	type args struct {
		pdr      *ie.IE
		spdrInfo *SPDRInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "emptyFlowDescription",
			fields: fields{
				Session:         nil,
				ResourceManager: nil,
				//TEEIDCache: nil,
			},
			args: args{
				pdr: ie.NewCreatePDR(
					ie.NewPDRID(2),
					ie.NewPDI(
						ie.NewSourceInterface(ie.SrcInterfaceCore),
						ie.NewUEIPAddress(2, "192.168.0.1", "", 0, 0),
						ie.NewSDFFilter("", "ttc", "", "", 0),
					),
				),
				spdrInfo: &SPDRInfo{},
			},
			wantErr: false,
		},
		{
			name: "emptyFlowDescriptionAndFilterID",
			fields: fields{
				Session:         nil,
				ResourceManager: nil,
				//TEEIDCache: nil,
			},
			args: args{
				pdr: ie.NewCreatePDR(
					ie.NewPDRID(2),
					ie.NewPDI(
						ie.NewSourceInterface(ie.SrcInterfaceCore),
						ie.NewUEIPAddress(2, "192.168.0.1", "", 0, 0),
						ie.NewSDFFilter("", "", "", "", 4096),
					),
				),
				spdrInfo: &SPDRInfo{},
			},
			wantErr: false,
		},
		{
			name: "invalidFlowDescription",
			fields: fields{
				Session:         nil,
				ResourceManager: nil,
				//TEEIDCache: nil,
			},
			args: args{
				pdr: ie.NewCreatePDR(
					ie.NewPDRID(2),
					ie.NewPDI(
						ie.NewSourceInterface(ie.SrcInterfaceCore),
						ie.NewUEIPAddress(2, "192.168.0.1", "", 0, 0),
						ie.NewSDFFilter("123", "", "", "", 4096),
					),
				),
				spdrInfo: &SPDRInfo{},
			},
			wantErr: true,
		},
		{
			name: "validFlowDescription",
			fields: fields{
				Session:         nil,
				ResourceManager: nil,
				//TEEIDCache: nil,
			},
			args: args{
				pdr: ie.NewCreatePDR(
					ie.NewPDRID(2),
					ie.NewPDI(
						ie.NewSourceInterface(ie.SrcInterfaceCore),
						ie.NewUEIPAddress(2, "192.168.0.1", "", 0, 0),
						ie.NewSDFFilter("permit out ip from 10.62.0.1 to 8.8.8.8/32", "", "", "", 0),
					),
				),
				spdrInfo: &SPDRInfo{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdrContext := &PDRCreationContext{
				Session:         tt.fields.Session,
				ResourceManager: tt.fields.ResourceManager,
				TEIDCache:       tt.fields.TEIDCache,
			}
			if err := pdrContext.extractPDR(tt.args.pdr, tt.args.spdrInfo); (err != nil) != tt.wantErr {
				t.Errorf("PDRCreationContext.extractPDR() error: %v, expected error: %v", err, tt.wantErr)
			}
		})
	}
}
