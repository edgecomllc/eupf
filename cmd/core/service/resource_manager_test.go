package service

import (
	"github.com/rs/zerolog/log"
	"net"
	"testing"
)

func TestAllocateIP(t *testing.T) {
	resourceManager, err := NewResourceManager(true, true, "10.61.0.0/16", 65536)
	if err != nil {
		log.Err(err)
	}

	//IP TESTS
	result1, err := resourceManager.IPAM.AllocateIP(12)
	expected1 := net.ParseIP("10.61.0.0")
	if result1.String() != expected1.String() {
		t.Errorf("Expected: %v, but got: %v", expected1, result1)
	}

	result2, err := resourceManager.IPAM.AllocateIP(16)
	expected2 := net.ParseIP("10.61.0.1")
	if result2.String() != expected2.String() {
		t.Errorf("Expected: %v, but got: %v", expected2, result2)
	}

	//resourceManager.IPAM.ReleaseIP(12)
	resourceManager.ReleaseResources(12)
	expecctedLen1 := 65535
	if expecctedLen1 != len(resourceManager.IPAM.freeIPs) {
		t.Errorf("Expected: %d, but got: %d", expecctedLen1, len(resourceManager.IPAM.freeIPs))
	}

	//resourceManager.IPAM.ReleaseIP(16)
	resourceManager.ReleaseResources(16)
	expecctedLen2 := 65536
	if expecctedLen2 != len(resourceManager.IPAM.freeIPs) {
		t.Errorf("Expected: %d, but got: %d", expecctedLen2, len(resourceManager.IPAM.freeIPs))
	}

	//TEID TEST
	resultTEID1, err := resourceManager.FTEIDM.AllocateTEID(12, 1)
	expectedTEID1 := 1
	if result1.String() != expected1.String() {
		t.Errorf("Expected: %v, but got: %v", expectedTEID1, resultTEID1)
	}

	resultTEID2, err := resourceManager.FTEIDM.AllocateTEID(12, 2)
	expectedTEID2 := 2
	if result2.String() != expected2.String() {
		t.Errorf("Expected: %v, but got: %v", expectedTEID2, resultTEID2)
	}

	resultTEID3, err := resourceManager.FTEIDM.AllocateTEID(16, 2)
	expectedTEID3 := 2
	if result2.String() != expected2.String() {
		t.Errorf("Expected: %v, but got: %v", expectedTEID3, resultTEID3)
	}

	//resourceManager.FTEIDM.ReleaseTEID(12)
	resourceManager.ReleaseResources(12)
	expecctedTEIDLen1 := 65535
	if expecctedTEIDLen1 != len(resourceManager.FTEIDM.freeTEIDs) {
		t.Errorf("Expected: %d, but got: %d", expecctedTEIDLen1, len(resourceManager.FTEIDM.freeTEIDs))
	}

	//resourceManager.FTEIDM.ReleaseTEID(16)
	resourceManager.ReleaseResources(16)
	expecctedTEIDLen2 := 65536
	if expecctedTEIDLen2 != len(resourceManager.FTEIDM.freeTEIDs) {
		t.Errorf("Expected: %d, but got: %d", expecctedTEIDLen2, len(resourceManager.FTEIDM.freeTEIDs))
	}
}
