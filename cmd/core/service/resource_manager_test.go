package service

import (
	"github.com/rs/zerolog/log"
	"testing"
)

func TestAllocateIP(t *testing.T) {
	resourceManager, err := NewResourceManager(true, 65536)
	if err != nil {
		log.Err(err)
	}

	//TEID TEST
	resultTEID1, err := resourceManager.FTEIDM.AllocateTEID(12, 1)
	if err != nil {
		t.Errorf("resultTEID1 AllocateTEID err: %v", err)
	}
	var expectedTEID1 uint32 = 1
	if resultTEID1 != expectedTEID1 {
		t.Errorf("Expected: %v, but got: %v", expectedTEID1, resultTEID1)
	}

	resultTEID2, err := resourceManager.FTEIDM.AllocateTEID(12, 2)
	if err != nil {
		t.Errorf("resultTEID2 AllocateTEID err: %v", err)
	}
	var expectedTEID2 uint32 = 2
	if resultTEID2 != expectedTEID2 {
		t.Errorf("Expected: %v, but got: %v", expectedTEID2, resultTEID2)
	}

	resultTEID3, err := resourceManager.FTEIDM.AllocateTEID(16, 2)
	if err != nil {
		t.Errorf("resultTEID3 AllocateTEID err: %v", err)
	}
	var expectedTEID3 uint32 = 3
	if resultTEID3 != expectedTEID3 {
		t.Errorf("Expected: %v, but got: %v", expectedTEID3, resultTEID3)
	}

}
