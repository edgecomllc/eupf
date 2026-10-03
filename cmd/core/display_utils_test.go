package core

import (
	"strings"
	"testing"

	"github.com/wmnsk/go-pfcp/ie"
)

// TestDisplayPdrMalformedSdfFilter pins displayPdr against a PDI whose SDF
// Filter claims a Flow-Description length far longer than the bytes actually
// present. go-pfcp's SDFFilterFields.UnmarshalBinary slices past the buffer
// in that case and panics; displayPdr is a logging path and must not take
// the UPF down with it.
func TestDisplayPdrMalformedSdfFilter(t *testing.T) {
	// Flags=0x01 (FD present), spare octet, FDLength=0xFFFE, one byte of
	// actual description.
	malformedSdfFilter := ie.New(ie.SDFFilter, []byte{0x01, 0x00, 0xFF, 0xFE, 0x41})
	pdi := ie.NewPDI(
		ie.NewSourceInterface(ie.SrcInterfaceAccess),
		malformedSdfFilter,
	)
	pdr := ie.NewCreatePDR(ie.NewPDRID(1), pdi)

	var sb strings.Builder
	displayPdr(&sb, pdr)
}
