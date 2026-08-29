package core

import (
	"net"

	"github.com/wmnsk/go-pfcp/ie"
)

// FarHeaderFields is the profile-agnostic result of parsing an Outer Header
// Creation IE. Profiles project their dialect-specific representation into
// this struct so that the downstream code can map it to ebpf.FarInfo uniformly.
type FarHeaderFields struct {
	// OuterHeaderCreation is the pre-computed ebpf.FarInfo.OuterHeaderCreation
	// value (profile-specific bit transformation of Description).
	OuterHeaderCreation uint8
	// Description is the raw OuterHeaderCreationDescription for display.
	Description uint16
	Teid        uint32
	IPv4Address net.IP
	IPv6Address net.IP
	PortNumber  uint16
	CTag        uint32
	STag        uint32
}

// PfcpProfile encapsulates the dialect of PFCP interaction with a particular
// control plane (CP). Implementations decide how connectors are built, how IEs
// are parsed/encoded, and how URR sequence numbering works.
//
// The profile is selected once at startup (see cmd/main.go) and injected into
// every PfcpConnection, replacing the scattered "if config.Conf.HuaweiSupport"
// branches that previously gated Huawei-specific behavior.
type PfcpProfile interface {
	// SxaConnector builds the AssociationConnector for an Sxa (S1U/S5S8) peer.
	SxaConnector(remoteNode string) (AssociationConnector, error)
	// SxbConnector builds the AssociationConnector for an Sxb (PA) peer.
	SxbConnector(remoteNode string) (AssociationConnector, error)

	// ParseOuterHeaderCreation extracts FarHeaderFields from an Outer Header
	// Creation IE (or a wrapping ForwardingParameters/UpdateForwardingParameters
	// IE) according to the profile's dialect.
	ParseOuterHeaderCreation(i *ie.IE) (FarHeaderFields, error)
	// ParseSubscriberData extracts IMSI and MSISDN from enterprise-specific
	// IEs if the profile supports them. Returns empty strings otherwise.
	ParseSubscriberData(ieArr []*ie.IE) (imsi, msisdn string)
	// ParseQFI extracts the QoS Flow Identifier from a QER IE. Returns false
	// if the QFI could not be determined.
	ParseQFI(qer *ie.IE) (uint8, bool)

	// NodeID encodes a Node ID IE for every PFCP message that carries one
	// (Association Setup/Update, Session Establishment Response, etc.).
	NodeID(nodeID string) *ie.IE

	// URSEQN builds a URSEQN IE for usage reports. The profile decides whether
	// to use the session-level sequence (Huawei) or the per-URR report sequence
	// (standard 3GPP).
	URSEQN(sessionSeq, reportSeq uint32) *ie.IE

	// HeartbeatRequestAdditionalIEs returns extra IEs to append to PFCP
	// Heartbeat Request messages, or nil if none. HuaweiProfile adds
	// ie.NewMetric(25); DefaultProfile adds none.
	HeartbeatRequestAdditionalIEs() []*ie.IE

	// SessionEstablishmentResponseAdditionalIEs returns extra IEs to append
	// to the Session Establishment Response after the standard IEs, or nil.
	// HuaweiProfile adds a second FSEID with a profile-specific IP;
	// DefaultProfile adds none.
	SessionEstablishmentResponseAdditionalIEs(localSEID uint64, nodeAddrV4 net.IP) []*ie.IE

	// UsageReportDeletionVendorIEs returns vendor-specific IEs for the
	// Usage Report within Session Deletion Response, or nil.
	// HuaweiProfile returns Huawei enterprise IEs (enterprise-id 2011);
	// DefaultProfile returns nil.
	UsageReportDeletionVendorIEs() []*ie.IE

	// UsageReportSessionReportVendorIEs returns vendor-specific IEs for
	// SendSessionReportUsage (volume threshold report), or nil.
	// HuaweiProfile returns Huawei enterprise IEs (enterprise-id 2011);
	// DefaultProfile returns nil.
	UsageReportSessionReportVendorIEs() []*ie.IE

	// UsageReportADCVendorIEs returns vendor-specific IEs for
	// SendSessionReportADC, or nil. sdfFilter is passed for the
	// conditional 36017 IE.
	// HuaweiProfile returns Huawei enterprise IEs (enterprise-id 2011);
	// DefaultProfile returns nil.
	UsageReportADCVendorIEs(sdfFilter string) []*ie.IE

	// SessionReportReleaseVendorIE returns the vendor-specific grouped IE
	// for Session Report Release (pdr-id-list), or nil. HuaweiProfile
	// returns ie.NewVendorSpecificGroupedIE(32799, 2011, pdrListIE...);
	// DefaultProfile returns nil.
	SessionReportReleaseVendorIE(pdrList []uint16) *ie.IE
}

// DefaultProfile implements the standard 3GPP PFCP dialect using the go-pfcp
// library. It does not parse enterprise-specific IEs and uses standard URR
// sequence numbering.
type DefaultProfile struct{}

func (DefaultProfile) SxaConnector(remoteNode string) (AssociationConnector, error) {
	return NewDefaultAssociationConnector(remoteNode)
}

func (DefaultProfile) SxbConnector(remoteNode string) (AssociationConnector, error) {
	return NewDefaultAssociationConnector(remoteNode)
}

func (DefaultProfile) ParseOuterHeaderCreation(i *ie.IE) (FarHeaderFields, error) {
	oc, err := i.OuterHeaderCreation()
	if err != nil {
		return FarHeaderFields{}, err
	}
	return FarHeaderFields{
		OuterHeaderCreation: uint8(oc.OuterHeaderCreationDescription >> 8),
		Description:         oc.OuterHeaderCreationDescription,
		Teid:                oc.TEID,
		IPv4Address:         oc.IPv4Address,
		IPv6Address:         oc.IPv6Address,
		PortNumber:          oc.PortNumber,
		CTag:                oc.CTag,
		STag:                oc.STag,
	}, nil
}

func (DefaultProfile) ParseSubscriberData(ieArr []*ie.IE) (imsi, msisdn string) {
	return "", ""
}

func (DefaultProfile) ParseQFI(qer *ie.IE) (uint8, bool) {
	qfi, err := qer.QFI()
	if err != nil {
		return 0, false
	}
	return qfi, true
}

func (DefaultProfile) NodeID(nodeID string) *ie.IE {
	return newIeNodeID(nodeID)
}

func (DefaultProfile) URSEQN(sessionSeq, reportSeq uint32) *ie.IE {
	return ie.NewURSEQN(reportSeq)
}

func (DefaultProfile) HeartbeatRequestAdditionalIEs() []*ie.IE {
	return nil
}

func (DefaultProfile) SessionEstablishmentResponseAdditionalIEs(localSEID uint64, nodeAddrV4 net.IP) []*ie.IE {
	return nil
}

func (DefaultProfile) UsageReportDeletionVendorIEs() []*ie.IE {
	return nil
}

func (DefaultProfile) UsageReportSessionReportVendorIEs() []*ie.IE {
	return nil
}

func (DefaultProfile) UsageReportADCVendorIEs(sdfFilter string) []*ie.IE {
	return nil
}

func (DefaultProfile) SessionReportReleaseVendorIE(pdrList []uint16) *ie.IE {
	return nil
}

// HuaweiProfile implements the Huawei SPGW-C PFCP dialect. It uses Huawei's
// Outer Header Creation encoding, enterprise-specific IEs for subscriber data
// and QFI, Huawei Node ID format, session-level URR sequence numbering, and a
// vendor-specific IE in usage reports.
type HuaweiProfile struct {
	S1UAddress  string
	S5S8Address string
	PAAddress   string
}

func (h HuaweiProfile) SxaConnector(remoteNode string) (AssociationConnector, error) {
	return NewSxaAssociationConnector(remoteNode, h.S1UAddress, h.S5S8Address)
}

func (h HuaweiProfile) SxbConnector(remoteNode string) (AssociationConnector, error) {
	return NewSxbAssociationConnector(remoteNode, h.PAAddress)
}

func (h HuaweiProfile) ParseOuterHeaderCreation(i *ie.IE) (FarHeaderFields, error) {
	oc, err := HuaweiOuterHeaderCreation(i)
	if err != nil {
		return FarHeaderFields{}, err
	}
	return FarHeaderFields{
		OuterHeaderCreation: uint8(1 << oc.OuterHeaderCreationDescription),
		Description:         oc.OuterHeaderCreationDescription,
		Teid:                oc.TEID,
		IPv4Address:         oc.IPv4Address,
		IPv6Address:         oc.IPv6Address,
		PortNumber:          oc.PortNumber,
		CTag:                oc.CTag,
		STag:                oc.STag,
	}, nil
}

func (h HuaweiProfile) ParseSubscriberData(ieArr []*ie.IE) (imsi, msisdn string) {
	imsiIdx := findEnterpriseSpecificIEindex(ieArr, 32769, 2011)
	if imsiIdx != -1 {
		imsi = DecodeDigitsFromBytes(ieArr[imsiIdx].Payload)
	}
	msisdnIdx := findEnterpriseSpecificIEindex(ieArr, 32770, 2011)
	if msisdnIdx != -1 {
		msisdn = DecodeDigitsFromBytes(ieArr[msisdnIdx].Payload)
	}
	return imsi, msisdn
}

func (h HuaweiProfile) ParseQFI(qer *ie.IE) (uint8, bool) {
	if qfi, err := qer.QFI(); err == nil {
		return qfi, true
	}
	if qfiId := findEnterpriseSpecificIEindex(qer.ChildIEs, 32785, 2011); qfiId != -1 {
		return qer.ChildIEs[qfiId].Payload[0], true
	}
	return 0, false
}

func (h HuaweiProfile) NodeID(nodeID string) *ie.IE {
	return newIeNodeIDHuawei(nodeID)
}

func (h HuaweiProfile) URSEQN(sessionSeq, reportSeq uint32) *ie.IE {
	return ie.NewURSEQN(sessionSeq)
}

func (h HuaweiProfile) HeartbeatRequestAdditionalIEs() []*ie.IE {
	return []*ie.IE{ie.NewMetric(25)}
}

func (h HuaweiProfile) SessionEstablishmentResponseAdditionalIEs(localSEID uint64, nodeAddrV4 net.IP) []*ie.IE {
	return []*ie.IE{ie.NewFSEID(localSEID, net.IPv4(10, 169, 26, 130), nil)}
}

func (h HuaweiProfile) UsageReportDeletionVendorIEs() []*ie.IE {
	return []*ie.IE{
		// CHOICE
		// urr-type
		//    enterprise-id: ---- 0x7db(2011)
		//    urr-level-type: ---- bearer(2)
		//    urr-function-type: ---- charging(1)
		//    urr-charging-type: ---- offlinepgw(3)
		ie.NewVendorSpecificIE(34000, 2011, []byte{0x02, 0x01, 0x03}),
		// CHOICE
		// bearer-sequence
		//    enterprise-id: ---- 0x7db(2011)
		//    bearer-sequence-value: ---- 0x1(1)
		ie.NewVendorSpecificIE(32843, 2011, []byte{1}),
		// CHOICE
		// private-stop-time
		//    enterprise-id: ---- 0x7db(2011)
		//    private-stop-time-value: ---- 0x000001917EEC1EEC
		ie.NewVendorSpecificIE(34010, 2011, []byte{0x00, 0x00, 0x01, 0x92, 0x1d, 0xca, 0x24, 0x8c}),
	}
}

func (h HuaweiProfile) UsageReportSessionReportVendorIEs() []*ie.IE {
	return []*ie.IE{
		// CHOICE
		// urr-type
		//    enterprise-id: ---- 0x7db(2011)
		//    urr-level-type: ---- bearer(2)
		//    urr-function-type: ---- charging(1)
		//    urr-charging-type: ---- offlinepgw(3)
		ie.NewVendorSpecificIE(34000, 2011, []byte{0x02, 0x01, 0x03}),
		// CHOICE
		// bearer-sequence
		//    enterprise-id: ---- 0x7db(2011)
		//    bearer-sequence-value: ---- 0x1(1)
		ie.NewVendorSpecificIE(32843, 2011, []byte{1}),
		// CHOICE
		// private-stop-time
		//    enterprise-id: ---- 0x7db(2011)
		//    private-stop-time-value: ---- 0x000001917EEC1EEC
		ie.NewVendorSpecificIE(34010, 2011, []byte{0x00, 0x00, 0x01, 0x92, 0x1d, 0xca, 0x24, 0x8c}),
		// CHOICE
		// private-time-of-first-packet
		//    enterprise-id: ---- 0x7db(2011)
		//    time-of-first-packet-value: ---- 0x000001917EEC1BE9
		ie.NewVendorSpecificIE(34011, 2011, []byte{0x00, 0x00, 0x01, 0x92, 0x1d, 0xca, 0x20, 0xad}),
		// CHOICE
		// private-time-of-last-packet
		//    enterprise-id: ---- 0x7db(2011)
		//    time-of-last-packet-value: ---- 0x000001917EEC1EEC
		ie.NewVendorSpecificIE(34012, 2011, []byte{0x00, 0x00, 0x01, 0x92, 0x1d, 0xca, 0x24, 0x8c}),
	}
}

func (h HuaweiProfile) UsageReportADCVendorIEs(sdfFilter string) []*ie.IE {
	ies := []*ie.IE{
		// CHOICE
		// urr-type
		//    enterprise-id: ---- 0x7db(2011)
		//    urr-level-type: ---- charging(1)
		//    urr-function-type: ---- charging(3)
		//    urr-charging-type: ---- onlinepgw(6)
		ie.NewVendorSpecificIE(34000, 2011, []byte{0x01, 0x03, 0x06}),
		// CHOICE
		// bearer-sequence
		//    enterprise-id: ---- 0x7db(2011)
		//    bearer-sequence-value: ---- 0x1(1)
		ie.NewVendorSpecificIE(32843, 2011, []byte{1}),
		// CHOICE
		// ???
		ie.NewVendorSpecificIE(36001, 2011, []byte{0x04}),
	}
	if len(sdfFilter) > 0 {
		// "ff2f00003e7065726d697420696e20362066726f6d203130302e38392e322e312f333220343531323820746f2031302e3136392e32302e3137382f33322031303635300001ff0040000001000000054101010000000000000000000000000000000000"
		unknownValue := []byte{0xff, 0x2f, 00, 00, (byte)(len(sdfFilter))}
		unknownValue = append(unknownValue, []byte(sdfFilter)...)
		unknownValue = append(unknownValue, []byte{0x0, 0x1, 0xff, 0x0, 0x40, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x5, 0x41, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}...)
		ies = append(ies, ie.NewVendorSpecificIE(36017, 2011, unknownValue))
	}
	return ies
}

func (h HuaweiProfile) SessionReportReleaseVendorIE(pdrList []uint16) *ie.IE {
	pdrListIE := []*ie.IE{}
	for _, pdrID := range pdrList {
		pdrListIE = append(pdrListIE, ie.NewPDRID(pdrID))
	}
	// CHOICE
	// delete-report-type
	// enterprise-id --- 0x7db(2011)
	// pdr-id-list
	// 	CHOICE
	// 	pdr-id --- 0x4(4)
	// 	CHOICE
	// 	pdr-id --- 0x5(5)
	return ie.NewVendorSpecificGroupedIE(32799, 2011, pdrListIE...)
}
