package main

import (
	"testing"
)

func TestIdTranslator(t *testing.T) {
	trans := NewIdTranslator()

	translatedId11 := trans.GetId(1, 1)
	if trans.GetId(1, 1) != translatedId11 {
		t.Errorf("Expected original ID to be %d, but it was %d", translatedId11, trans.GetId(1, 1))
	}

	translatedId11Check := trans.GetId(1, 1)
	if trans.GetId(1, 1) != translatedId11 || translatedId11Check != translatedId11 {
		t.Errorf("Expected original ID to be %d, but it was %d", translatedId11, trans.GetId(1, 1))
	}

	translatedId12 := trans.GetId(1, 2)
	if trans.GetId(1, 2) != translatedId12 {
		t.Errorf("Expected original ID to be %d, but it was %d", translatedId12, trans.GetId(1, 2))
	}

	translatedId23 := trans.GetId(2, 3)
	if trans.GetId(2, 3) != translatedId23 {
		t.Errorf("Expected original ID to be %d, but it was %d", translatedId23, trans.GetId(2, 3))
	}

	translatedId22 := trans.GetId(2, 2)
	if trans.GetId(2, 2) != translatedId22 {
		t.Errorf("Expected original ID to be %d, but it was %d", translatedId22, trans.GetId(2, 1))
	}

	removedGlobalId := trans.RemoveId(1, 1)
	if _, exists := trans.BucketMappingTables[1][1]; exists {
		t.Errorf("Expected ID to be removed from bucket, but it still exists")
	}
	if _, exists := trans.GlobalMappingTable[removedGlobalId]; exists {
		t.Errorf("Expected ID to be removed from global map, but it still exists")
	}
}
