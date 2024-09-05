package core

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/wmnsk/go-pfcp/ie"
)

func has8thBit(f uint8) bool {
	return (f&0x80)>>7 == 1
}

func has7thBit(f uint8) bool {
	return (f&0x40)>>6 == 1
}

func has6thBit(f uint8) bool {
	return (f&0x20)>>5 == 1
}

func has5thBit(f uint8) bool {
	return (f&0x010)>>4 == 1
}

func has4thBit(f uint8) bool {
	return (f&0x08)>>3 == 1
}

func has3rdBit(f uint8) bool {
	return (f&0x04)>>2 == 1
}

func has2ndBit(f uint8) bool {
	return (f&0x02)>>1 == 1
}

func has1stBit(f uint8) bool {
	return (f & 0x01) == 1
}

// OuterHeaderCreationFields represents a fields contained in OuterHeaderCreation IE.
type OuterHeaderCreationFields struct {
	OuterHeaderCreationDescription uint16
	TEID                           uint32
	IPv4Address                    net.IP
	IPv6Address                    net.IP
	PortNumber                     uint16
	CTag                           uint32
	STag                           uint32
}

// NewOuterHeaderCreationFields creates a new OuterHeaderCreationFields.
func NewOuterHeaderCreationFields(desc uint16, teid uint32, v4, v6 string, port uint16, ctag, stag uint32) *OuterHeaderCreationFields {
	f := &OuterHeaderCreationFields{OuterHeaderCreationDescription: desc}

	oct5 := uint8((desc & 0xff00) >> 8)

	if has1stBit(oct5) || has2ndBit(oct5) {
		f.TEID = teid
	}

	if has1stBit(oct5) || has3rdBit(oct5) || has5thBit(oct5) {
		f.IPv4Address = net.ParseIP(v4).To4()
	}

	if has2ndBit(oct5) || has4thBit(oct5) || has6thBit(oct5) {
		f.IPv6Address = net.ParseIP(v6).To16()
	}

	if has3rdBit(oct5) || has4thBit(oct5) {
		f.PortNumber = port
	}

	if has7thBit(oct5) {
		f.CTag = ctag
	}

	if has8thBit(oct5) {
		f.STag = stag
	}

	return f
}

// UnmarshalBinary parses b into IE.
func (f *OuterHeaderCreationFields) UnmarshalBinary(b []byte) error {
	l := len(b)
	if l < 2 {
		return io.ErrUnexpectedEOF
	}

	f.OuterHeaderCreationDescription = uint16(b[0])
	offset := 1

	oct5 := b[0]

	if oct5 != 0 {
		return io.ErrUnexpectedEOF
	}

	if l < offset+4 {
		return io.ErrUnexpectedEOF
	}
	f.TEID = binary.BigEndian.Uint32(b[offset : offset+4])
	offset += 4

	if l < offset+4 {
		return io.ErrUnexpectedEOF
	}
	f.IPv4Address = net.IP(b[offset : offset+4]).To4()
	offset += 4
	return nil
}

// Marshal returns the serialized bytes of OuterHeaderCreationFields.
func (f *OuterHeaderCreationFields) Marshal() ([]byte, error) {
	b := make([]byte, f.MarshalLen())
	if err := f.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (f *OuterHeaderCreationFields) MarshalTo(b []byte) error {
	l := len(b)
	if l < 2 {
		return io.ErrUnexpectedEOF
	}

	binary.BigEndian.PutUint16(b[0:2], f.OuterHeaderCreationDescription)
	offset := 2

	oct5 := uint8((f.OuterHeaderCreationDescription & 0xff00) >> 8)

	if has1stBit(oct5) || has2ndBit(oct5) {
		binary.BigEndian.PutUint32(b[offset:offset+4], f.TEID)
		offset += 4
	}

	if has1stBit(oct5) || has3rdBit(oct5) || has5thBit(oct5) {
		copy(b[offset:offset+4], f.IPv4Address)
		offset += 4
	}

	if has2ndBit(oct5) || has4thBit(oct5) || has6thBit(oct5) {
		copy(b[offset:offset+16], f.IPv6Address)
		offset += 16
	}

	if has3rdBit(oct5) || has4thBit(oct5) {
		binary.BigEndian.PutUint16(b[offset:offset+2], f.PortNumber)
	}

	if has7thBit(oct5) {
		p := make([]byte, 4)
		binary.BigEndian.PutUint32(p, f.CTag)
		copy(b[offset:offset+3], p[1:4])
		offset += 3
	}

	if has8thBit(oct5) {
		p := make([]byte, 4)
		binary.BigEndian.PutUint32(p, f.STag)
		copy(b[offset:offset+3], p[1:4])
	}

	return nil
}

// MarshalLen returns field length in integer.
func (f *OuterHeaderCreationFields) MarshalLen() int {
	l := 2

	if f.HasTEID() {
		l += 4
	}
	if f.HasIPv4() {
		l += 4
	}
	if f.HasIPv6() {
		l += 16
	}
	if f.HasPortNumber() {
		l += 2
	}
	if f.HasCTag() {
		l += 3
	}
	if f.HasSTag() {
		l += 3
	}

	return l
}

// HasTEID reports wether TEID field is set.
func (f *OuterHeaderCreationFields) HasTEID() bool {
	// The TEID field shall be present
	// if the Outer Header Creation Description requests
	// the creation of aGTP-U header. Otherwise it shall not be present.
	//desc := uint8((f.OuterHeaderCreationDescription & 0xff00) >> 8)
	//return has1stBit(desc) || has2ndBit(desc)
	desc := uint8((f.OuterHeaderCreationDescription & 0xff))
	return desc == 0
}

// HasIPv4 reports wether IPv4 Address field is set.
func (f *OuterHeaderCreationFields) HasIPv4() bool {
	// The IPv4 Address field shall be present
	// if the Outer Header Creation Description requests
	// the creation of an IPv4 header. Otherwise it shall not be present.
	//desc := uint8((f.OuterHeaderCreationDescription & 0xff00) >> 8)
	//return has1stBit(desc) || has3rdBit(desc) || has5thBit(desc)
	desc := uint8((f.OuterHeaderCreationDescription & 0xff))
	return desc == 0
}

// HasIPv6 reports wether IPv6 Address field is set.
func (f *OuterHeaderCreationFields) HasIPv6() bool {
	// The IPv6 Address field shall be present
	// if the Outer Header Creation Description requests
	// the creation of an IPv6 header. Otherwise it shall not be present.
	//desc := uint8((f.OuterHeaderCreationDescription & 0xff00) >> 8)
	//return has2ndBit(desc) || has4thBit(desc) || has6thBit(desc)
	return false
}

// HasPortNumber reports wether Port Number field is set.
func (f *OuterHeaderCreationFields) HasPortNumber() bool {
	// The Port Number field shall be present
	// if the Outer Header Creation Description requests
	// the creation of a UDP/IP header. Otherwise it shall not be present.
	// desc := uint8((f.OuterHeaderCreationDescription & 0xff00) >> 8)
	// return has3rdBit(desc) || has4thBit(desc)
	return false
}

// HasCTag reports wether C-TAG field is set.
func (f *OuterHeaderCreationFields) HasCTag() bool {
	// The C-TAG field shall be present
	// if the Outer Header Creation Description requests
	// the setting of the C-Tag in Ethernet packet. Otherwise it shall not be present.
	// desc := uint8((f.OuterHeaderCreationDescription & 0xff00) >> 8)
	// return has7thBit(desc)
	return false
}

// HasSTag reports wether S-TAG field is set.
func (f *OuterHeaderCreationFields) HasSTag() bool {
	// The S-TAG field shall be present
	// if the Outer Header Creation Description requests
	// the setting of the S-Tag in Ethernet packet. Otherwise it shall not be present.
	// desc := uint8((f.OuterHeaderCreationDescription & 0xff00) >> 8)
	// return has8thBit(desc)
	return false
}

// IsN19 reports wether Outer Header Creation Description has N19 Indication.
func (f *OuterHeaderCreationFields) IsN19() bool {
	// desc := uint8(f.OuterHeaderCreationDescription & 0x00FF)
	// return has1stBit(desc)
	return false
}

// IsN6 reports wether Outer Header Creation Description has N9 Indication
func (f *OuterHeaderCreationFields) IsN6() bool {
	desc := uint8(f.OuterHeaderCreationDescription & 0x00FF)
	return has2ndBit(desc)
}

// IsLLSSMCTEID reports wether Outer Header Creation Description has Low Layer SSM and C-TEID
// This bit has been introduced in release 17.2
func (f *OuterHeaderCreationFields) IsLLSSMCTEID() bool {
	// desc := uint8(f.OuterHeaderCreationDescription & 0x00FF)
	// return has3rdBit(desc)
	return false
}

// ParseOuterHeaderCreationFields parses b into OuterHeaderCreationFields.
func ParseOuterHeaderCreationFields(b []byte) (*OuterHeaderCreationFields, error) {
	f := &OuterHeaderCreationFields{}
	if err := f.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return f, nil
}

// ..................forwarding-parameters
// ....................CHOICE
// ......................destination-interface
// 294>   00   0000****   ........................spare --- 0x0(0)
// ****0000   ........................interface-value --- access(0)
// ....................CHOICE
// ......................outer-header-creation
// 299>   00   00000000   ........................outer-header-creation --- gtpu-udp-ipv4(0)
// 300>   02   00000010
// 301>   44   01000100
// 302>   09   00001001
// 303>   C0   11000000   ........................teid --- 0x24409c0(38013376)
// ........................ipv4-address
// 304>   0A   00001010   ..........................uladdr1 --- 0xa(10)
// 305>   A9   10101001   ..........................uladdr2 --- 0xa9(169)
// 306>   70   01110000   ..........................uladdr3 --- 0x70(112)
// 307>   91   10010001   ..........................uladdr4 --- 0x91(145)

// update-forwarding-parameters
// CHOICE
//    outer-header-creation
// 	  length: ---- 0x9(9)
// 	  outer-header-creation-old-version: ---- gtpu-udp-ipv4(0)
// 	  teid: ---- 0xca632b5(212218549)
// 	  ipv4-address
// 		 uladdr1: ---- 0xa(10)
// 		 uladdr2: ---- 0xa9(169)
// 		 uladdr3: ---- 0xfa(250)
// 		 uladdr4: ---- 0x2e(46)

func HuaweiOuterHeaderCreation(i *ie.IE) (*OuterHeaderCreationFields, error) {
	switch i.Type {
	case ie.OuterHeaderCreation:
		f, err := ParseOuterHeaderCreationFields(i.Payload)
		if err != nil {
			return nil, err
		}
		return f, nil
	case ie.ForwardingParameters:
		ies, err := i.ForwardingParameters()
		if err != nil {
			return nil, err
		}
		for _, x := range ies {
			if x.Type == ie.OuterHeaderCreation {
				return HuaweiOuterHeaderCreation(x)
			}
		}
		return nil, ie.ErrIENotFound
	case ie.UpdateForwardingParameters:
		ies, err := i.UpdateForwardingParameters()
		if err != nil {
			return nil, err
		}
		for _, x := range ies {
			if x.Type == ie.OuterHeaderCreation {
				return HuaweiOuterHeaderCreation(x)
			}
		}
		return nil, ie.ErrIENotFound
	case ie.DuplicatingParameters:
		ies, err := i.DuplicatingParameters()
		if err != nil {
			return nil, err
		}
		for _, x := range ies {
			if x.Type == ie.OuterHeaderCreation {
				return HuaweiOuterHeaderCreation(x)
			}
		}
		return nil, ie.ErrIENotFound
	case ie.UpdateDuplicatingParameters:
		ies, err := i.UpdateDuplicatingParameters()
		if err != nil {
			return nil, err
		}
		for _, x := range ies {
			if x.Type == ie.OuterHeaderCreation {
				return HuaweiOuterHeaderCreation(x)
			}
		}
		return nil, ie.ErrIENotFound
	case ie.RedundantTransmissionParameters:
		ies, err := i.RedundantTransmissionParameters()
		if err != nil {
			return nil, err
		}
		for _, x := range ies {
			if x.Type == ie.OuterHeaderCreation {
				return HuaweiOuterHeaderCreation(x)
			}
		}
		return nil, ie.ErrIENotFound
	default:
		return nil, &ie.InvalidTypeError{Type: i.Type}
	}
}
