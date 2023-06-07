package main

import (
	"testing"
)

func TestIdTracker(t *testing.T) {
	Idtracker := NewIdTracker(1024)
	for i := uint32(0); i < 1024; i++ {
		id, _ := Idtracker.GetNext()
		if id != i {
			t.Errorf("IdTracker.GetId() = %d, want %d", id, i)
		}
	}
	Idtracker.Release(15)
	id, _ := Idtracker.GetNext()
	if id != 15 {
		t.Errorf("IdTracker.GetId() = %d, want %d", id, 15)
	}
}
