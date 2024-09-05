package core

import (
	"net"
	"strings"

	"github.com/wmnsk/go-pfcp/ie"
)

func EncodeFQDN(fqdn string) []byte {
	b := make([]byte, len(fqdn) /*+1*/)

	var offset = 0
	for _, label := range strings.Split(fqdn, ".") {
		l := len(label)
		//b[offset] = uint8(l)
		copy(b[offset: /*+1*/], label)
		offset += l /*+ 1*/
	}

	return b
}

func NewNodeIDHuawei(ipv4, ipv6, fqdn string) *ie.IE {
	var p []byte

	switch {
	case ipv4 != "":
		p = make([]byte, 5)
		p[0] = ie.NodeIDIPv4Address
		copy(p[1:], net.ParseIP(ipv4).To4())
	case ipv6 != "":
		p = make([]byte, 17)
		p[0] = ie.NodeIDIPv6Address
		copy(p[1:], net.ParseIP(ipv6).To16())
	case fqdn != "":
		p = make([]byte, 1+len([]byte(fqdn)))
		p[0] = ie.NodeIDFQDN
		copy(p[1:], EncodeFQDN(fqdn))
	default: // all params are empty
		return nil
	}

	return ie.New(ie.NodeID, p)
}

func newIeNodeIDHuawei(nodeID string) *ie.IE {
	ip := net.ParseIP(nodeID)
	if ip != nil {
		if ip.To4() != nil {
			return NewNodeIDHuawei(nodeID, "", "")
		}
		return NewNodeIDHuawei("", nodeID, "")
	}
	return NewNodeIDHuawei("", "", nodeID)
}
