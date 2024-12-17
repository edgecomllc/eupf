package ebpf

import (
	"github.com/rs/zerolog/log"
)

type UpfXdpActionStatistic struct {
	BpfObjects *BpfObjects
	previous   UpfCounters
}

type UpfCounters struct {
	RxArp      uint64
	RxIcmp     uint64
	RxIcmp6    uint64
	RxIp4      uint64
	RxIp6      uint64
	RxTcp      uint64
	RxUdp      uint64
	RxOther    uint64
	RxGtpEcho  uint64
	RxGtpPdu   uint64
	RxGtpOther uint64
	RxGtpUnexp uint64
}

type UpfStatistic struct {
	Counters UpfCounters
	XdpStats [5]uint64
}

func (current *UpfCounters) Add(new UpfCounters) {
	current.RxArp += new.RxArp
	current.RxIcmp += new.RxIcmp
	current.RxIcmp6 += new.RxIcmp6
	current.RxIp4 += new.RxIp4
	current.RxIp6 += new.RxIp6
	current.RxTcp += new.RxTcp
	current.RxUdp += new.RxUdp
	current.RxOther += new.RxOther
	current.RxGtpEcho += new.RxGtpEcho
	current.RxGtpPdu += new.RxGtpPdu
	current.RxGtpOther += new.RxGtpOther
}

func (current *UpfCounters) Delta(new UpfCounters) UpfCounters {

	delta := UpfCounters{}

	delta.RxArp = new.RxArp - current.RxArp
	delta.RxIcmp = new.RxIcmp - current.RxIcmp
	delta.RxIcmp6 = new.RxIcmp6 - current.RxIcmp6
	delta.RxIp4 = new.RxIp4 - current.RxIp4
	delta.RxIp6 = new.RxIp6 - current.RxIp6
	delta.RxTcp = new.RxTcp - current.RxTcp
	delta.RxUdp = new.RxUdp - current.RxUdp
	delta.RxOther = new.RxOther - current.RxOther
	delta.RxGtpEcho = new.RxGtpEcho - current.RxGtpEcho
	delta.RxGtpPdu = new.RxGtpPdu - current.RxGtpPdu
	delta.RxGtpOther = new.RxGtpOther - current.RxGtpOther
	return delta
}

// Getters for the upf_xdp_statistic (xdp_action)

func (stat *UpfXdpActionStatistic) getUpfXdpStatisticField(field uint32) uint64 {

	var statistics []IpEntrypointUpfStatistic
	err := stat.BpfObjects.UpfExtStat.Lookup(uint32(0), &statistics)
	if err != nil {
		log.Info().Msg(err.Error())
		return 0
	}

	var totalValue uint64 = 0
	for _, statistic := range statistics {
		totalValue += statistic.XdpActions[field]
	}

	return totalValue
}

func (stat *UpfXdpActionStatistic) GetAborted() uint64 {
	return stat.getUpfXdpStatisticField(uint32(0))
}

func (stat *UpfXdpActionStatistic) GetDrop() uint64 {
	return stat.getUpfXdpStatisticField(uint32(1))
}

func (stat *UpfXdpActionStatistic) GetPass() uint64 {
	return stat.getUpfXdpStatisticField(uint32(2))
}

func (stat *UpfXdpActionStatistic) GetTx() uint64 {
	return stat.getUpfXdpStatisticField(uint32(3))
}

func (stat *UpfXdpActionStatistic) GetRedirect() uint64 {
	return stat.getUpfXdpStatisticField(uint32(4))
}

// Getters for the upf_ext_stat (upf_counters)
// #TODO: Do not retrieve the whole struct each time.
func (stat *UpfXdpActionStatistic) GetUpfExtStat() UpfCounters {

	var statistics []IpEntrypointUpfStatistic
	var counters UpfCounters
	err := stat.BpfObjects.UpfExtStat.Lookup(uint32(0), &statistics)
	if err != nil {
		log.Info().Msg(err.Error())
		return counters
	}

	for _, statistic := range statistics {
		counters.Add(statistic.UpfCounters)
	}

	delta := stat.previous.Delta(counters)
	stat.previous = counters
	return delta
}
