package core

import (
	"fmt"
	"net"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
)

// buildSessionReportUsageIEs builds the IEs for SendSessionReportUsage
// (volume threshold report) without sending. Exposed for testability.
func buildSessionReportUsageIEs(profile PfcpProfile, urrid, sessionSeq, reportSeq uint32, uplink, downlink uint64) []*ie.IE {
	usageReportIEs := append([]*ie.IE{
		ie.NewURRID(urrid),
		profile.URSEQN(sessionSeq, reportSeq),
		ie.NewUsageReportTrigger(1<<1, 0, 0), //Volume Threshold
		ie.NewEndTime(time.Now()),
		ie.NewVolumeMeasurement(0x6, 0, uplink, downlink, 0, 0, 0),
		ie.NewTimeOfFirstPacket(time.Now()),
		ie.NewTimeOfLastPacket(time.Now()),
	}, profile.UsageReportSessionReportVendorIEs()...)
	return []*ie.IE{
		ie.NewReportType(0, 0, 1, 0),
		ie.NewUsageReportWithinSessionReportRequest(usageReportIEs...),
	}
}

// buildSessionReportADCIEs builds the IEs for SendSessionReportADC
// (ADC report) without sending. Exposed for testability.
func buildSessionReportADCIEs(profile PfcpProfile, urrid, sessionSeq, reportSeq uint32, sdfFilter string) []*ie.IE {
	adcReportIEs := append([]*ie.IE{
		ie.NewURRID(urrid),
		profile.URSEQN(sessionSeq, reportSeq),
	}, profile.UsageReportADCVendorIEs(sdfFilter)...)
	return []*ie.IE{
		ie.NewReportType(0, 0, 1, 0),
		ie.NewUsageReportWithinSessionReportRequest(adcReportIEs...),
	}
}

// buildSessionReportReleaseIEs builds the IEs for SendSessionReportSessionRelease
// without sending. Exposed for testability.
func buildSessionReportReleaseIEs(profile PfcpProfile, pdrList []uint16) []*ie.IE {
	ies := []*ie.IE{
		ie.NewReportType(0, 0, 0, 0),
	}
	if vendorIE := profile.SessionReportReleaseVendorIE(pdrList); vendorIE != nil {
		ies = append(ies, vendorIE)
	}
	return ies
}

// sxaFeaturesOctets returns the UPFunctionFeatures octets used in Sxa/Sxb
// Association Setup Requests (Huawei SPGW-C dialect).
func sxaFeaturesOctets() []uint8 {
	featuresOctets := []uint8{0, 0}
	featuresOctets[0] = setBit(featuresOctets[0], 1)
	featuresOctets[0] = setBit(featuresOctets[0], 2)
	featuresOctets[0] = setBit(featuresOctets[0], 6)
	featuresOctets[0] = setBit(featuresOctets[0], 7)
	return featuresOctets
}

// buildDefaultSetupRequestIEs builds the IEs for a Default Association Setup
// Request without sending. Exposed for testability.
func buildDefaultSetupRequestIEs(profile PfcpProfile, nodeId string, recoveryTS time.Time, featuresOctets []uint8) []*ie.IE {
	return []*ie.IE{
		profile.NodeID(nodeId),
		ie.NewRecoveryTimeStamp(recoveryTS),
		ie.NewUPFunctionFeatures(featuresOctets...),
	}
}

// buildSxaSetupRequestIEs builds the IEs for an Sxa Association Setup Request
// (Huawei SPGW-C dialect) without sending. Exposed for testability.
func buildSxaSetupRequestIEs(profile PfcpProfile, nodeId string, recoveryTS time.Time, s1uAddress, s5s8Address string) ([]*ie.IE, error) {
	s1uIP := net.ParseIP(s1uAddress)
	if s1uIP == nil {
		return nil, fmt.Errorf("failed to parse S1-U IP address: %s", s1uAddress)
	}
	s5s8IP := net.ParseIP(s5s8Address)
	if s5s8IP == nil {
		return nil, fmt.Errorf("failed to parse S5/S8 IP address: %s", s5s8Address)
	}

	ipSuiteName := "0001" + nodeId
	ipsuitInfo := make([]byte, 0)
	ipsuitInfo = append(ipsuitInfo, 0xA8, 0x00)
	ipsuitInfo = append(ipsuitInfo, (byte)(len(ipSuiteName)))
	ipsuitInfo = append(ipsuitInfo, ipSuiteName...)
	ipsuitInfo = append(ipsuitInfo, s1uIP.To4()...)
	ipsuitInfo = append(ipsuitInfo, s5s8IP.To4()...)
	ipsuitInfo = append(ipsuitInfo, s1uIP.To4()...)

	featuresOctets := sxaFeaturesOctets()

	return []*ie.IE{
		profile.NodeID(nodeId),
		ie.NewRecoveryTimeStamp(recoveryTS),
		ie.NewUPFunctionFeatures(featuresOctets...),
		ie.NewVendorSpecificIE(32787, 2011, ipsuitInfo),
		ie.NewVendorSpecificIE(32803, 2011, []byte{1}),
		ie.NewVendorSpecificIE(32806, 2011, []byte{0}),
		ie.NewVendorSpecificIE(32857, 2011, []byte{0}),
		ie.NewVendorSpecificIE(32900, 2011, []byte{3}),
		ie.NewVendorSpecificIE(32901, 2011, []byte{1}),
	}, nil
}

// buildSxbSetupRequestIEs builds the IEs for an Sxb Association Setup Request
// (Huawei SPGW-C dialect) without sending. Exposed for testability.
func buildSxbSetupRequestIEs(profile PfcpProfile, nodeId string, recoveryTS time.Time, paAddress string) ([]*ie.IE, error) {
	paIP := net.ParseIP(paAddress)
	if paIP == nil {
		return nil, fmt.Errorf("failed to parse PA IP address: %s", paAddress)
	}

	ipSuiteName := "0001" + nodeId
	ipsuitInfo := make([]byte, 0)
	ipsuitInfo = append(ipsuitInfo, 0x02, 0x00)
	ipsuitInfo = append(ipsuitInfo, (byte)(len(ipSuiteName)))
	ipsuitInfo = append(ipsuitInfo, ipSuiteName...)
	ipsuitInfo = append(ipsuitInfo, paIP.To4()...)

	featuresOctets := sxaFeaturesOctets()

	return []*ie.IE{
		profile.NodeID(nodeId),
		ie.NewRecoveryTimeStamp(recoveryTS),
		ie.NewUPFunctionFeatures(featuresOctets...),
		ie.NewVendorSpecificIE(32787, 2011, ipsuitInfo),
		ie.NewVendorSpecificIE(32803, 2011, []byte{1}),
		ie.NewVendorSpecificIE(32806, 2011, []byte{0}),
		ie.NewVendorSpecificIE(32857, 2011, []byte{0}),
		ie.NewVendorSpecificIE(32900, 2011, []byte{3}),
		ie.NewVendorSpecificIE(32901, 2011, []byte{1}),
	}, nil
}
