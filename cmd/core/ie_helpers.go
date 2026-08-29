package core

import (
	"strings"
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

func DecodeDigitsFromBytes(buffer []byte) string {
	decoded := make([]byte, len(buffer)*2)
	for i, b := range buffer {
		decoded[2*i], decoded[2*i+1] = hexDigit(b&0x0F), hexDigit((b&0xF0)>>4)
	}

	digits := string(decoded)
	if digits[len(digits)-1] == 'f' {
		return digits[:len(digits)-1]
	}
	return digits
}

func hexDigit(nibble byte) byte {
	if nibble < 10 {
		return '0' + nibble
	}
	return 'a' + (nibble - 10)
}

// EncodeFQDN encodes the given string as the Name Syntax defined
// in RFC 2181, RFC 1035 and RFC 1123.
func EncodeFQDN(fqdn string) []byte {
	b := make([]byte, len(fqdn)+1)

	var offset = 0
	for _, label := range strings.Split(fqdn, ".") {
		l := len(label)
		b[offset] = uint8(l)
		copy(b[offset+1:], label)
		offset += l + 1
	}

	return b
}
