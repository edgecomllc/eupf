package core

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func TestHuaweiParseOuterHeaderCreationFields(t *testing.T) {

	buffer := []byte{0x00, 0x74, 0x03, 0x30, 0x0b, 0x0a, 0xa9, 0x70, 0xae}
	result, err := ParseOuterHeaderCreationFields(buffer)
	if err != nil {
		t.Errorf("Error")
	}

	if !result.HasTEID() {
		t.Errorf("Error")
	}

	if !result.HasIPv4() {
		t.Errorf("Error")
	}

	t.Logf("TEID: %v, IP: %v ", result.TEID, result.IPv4Address)
}

func TestDecodeDigitsFromBytes(t *testing.T) {

	buffer1 := []byte{0x97, 0x89, 0x03, 0x00, 0x20, 0xf3}
	result := DecodeDigitsFromBytes(buffer1)
	if result != "79983000023" {
		t.Errorf("error result: %s", result)
		return
	}
	t.Logf("Decoded digits: %s", result)

	buffer2, _ := hex.DecodeString("52500300000020f3")
	result = DecodeDigitsFromBytes(buffer2)
	if result != "250530000000023" {
		t.Errorf("error result: %s", result)
		return
	}
	t.Logf("Decoded digits: %s", result)
}

func TestEncodeFQDN(t *testing.T) {
	type args struct {
		fqdn string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "dgw4",
			args: args{fqdn: "dgw4"},
			want: []byte{0x64, 0x67, 0x77, 0x34},
		},
		{
			name: "nactech.dgw",
			args: args{fqdn: "nactech.dgw"},
			want: []byte{0x6e, 0x61, 0x63, 0x74, 0x65, 0x63, 0x68, 0x2e, 0x64, 0x67, 0x77},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeFQDNHuawei(tt.args.fqdn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeFQDN() = %v, want %v", hex.EncodeToString(got), hex.EncodeToString(tt.want))
			}
		})
	}
}
