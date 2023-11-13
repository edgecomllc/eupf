package service

import (
	"github.com/rs/zerolog/log"
	"net"
	"testing"
)

func TestAllocateIP(t *testing.T) {
	ipam, err := NewIPAM("10.61.0.0/16")
	if err != nil {
		log.Err(err)
	}

	//IP TESTS
	result1, err := ipam.AllocateIP(12)
	expected1 := net.ParseIP("10.61.0.0")
	if result1.String() != expected1.String() {
		t.Errorf("Expected: %v, but got: %v", expected1, result1)
	}

	result2, err := ipam.AllocateIP(16)
	expected2 := net.ParseIP("10.61.0.1")
	if result2.String() != expected2.String() {
		t.Errorf("Expected: %v, but got: %v", expected2, result2)
	}

	ipam.ReleaseIP(12)
	expecctedLen1 := 65535
	if expecctedLen1 != len(ipam.freeIPs) {
		t.Errorf("Expected: %d, but got: %d", expecctedLen1, len(ipam.freeIPs))
	}

	ipam.ReleaseIP(16)
	expecctedLen2 := 65536
	if expecctedLen2 != len(ipam.freeIPs) {
		t.Errorf("Expected: %d, but got: %d", expecctedLen2, len(ipam.freeIPs))
	}

	//TEID TEST
	resultTEID1, err := ipam.AllocateTEID(12)
	expectedTEID1 := 1
	if result1.String() != expected1.String() {
		t.Errorf("Expected: %v, but got: %v", expectedTEID1, resultTEID1)
	}

	resultTEID2, err := ipam.AllocateTEID(16)
	expectedTEID2 := 2
	if result2.String() != expected2.String() {
		t.Errorf("Expected: %v, but got: %v", expectedTEID2, resultTEID2)
	}

	ipam.ReleaseTEID(12)
	expecctedTEIDLen1 := 65535
	if expecctedTEIDLen1 != len(ipam.freeTEIDs) {
		t.Errorf("Expected: %d, but got: %d", expecctedTEIDLen1, len(ipam.freeTEIDs))
	}

	ipam.ReleaseTEID(16)
	expecctedTEIDLen2 := 65536
	if expecctedTEIDLen2 != len(ipam.freeTEIDs) {
		t.Errorf("Expected: %d, but got: %d", expecctedTEIDLen2, len(ipam.freeTEIDs))
	}
}
