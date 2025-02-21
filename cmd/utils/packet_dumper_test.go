package utils

import (
	"net/netip"
	"testing"
)

func TestPacketDumper_RunWrite(t *testing.T) {

	dumper, _ := NewPacketDumper("")
	dumper.DumpRawIn(netip.MustParseAddrPort("1.2.3.4:8805"), netip.MustParseAddrPort("1.2.3.4:8806"), []byte{0, 0, 0})
	dumper.DumpRawOut(netip.MustParseAddrPort("1.2.3.4:8805"), netip.MustParseAddrPort("1.2.3.4:8806"), []byte{0, 0, 0})
	dumper.Close(true)
	dumper.Write()

	if dumper.GetPacketsWritten() != 2 {
		t.Errorf("Wrong number of written packets")
	}
}
