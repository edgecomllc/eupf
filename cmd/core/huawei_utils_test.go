package core

import (
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
