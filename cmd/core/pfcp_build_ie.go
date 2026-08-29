package core

import (
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
