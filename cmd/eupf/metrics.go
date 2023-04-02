package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Association setup requests
	AsrSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "upf_asr_success",
		Help: "The total number of successful association setup requests",
	})
	AsrReject = promauto.NewCounter(prometheus.CounterOpts{
		Name: "upf_asr_reject",
		Help: "The total number of rejected association setup requests",
	})

	// Session establishment requests
	SerSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "upf_ser_success",
		Help: "The total number of successful session establishment requests",
	})
	SerReject = promauto.NewCounter(prometheus.CounterOpts{
		Name: "upf_ser_reject",
		Help: "The total number of rejected session establishment requests",
	})

	// Session Deletion requests
	SdrSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "upf_sdr_success",
		Help: "The total number of successful session deletion requests",
	})
	SdrReject = promauto.NewCounter(prometheus.CounterOpts{
		Name: "upf_sdr_reject",
		Help: "The total number of rejected session deletion requests",
	})

	// Session modification requests
	SmrSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "upf_smr_success",
		Help: "The total number of successful session modification requests",
	})
	SmrReject = promauto.NewCounter(prometheus.CounterOpts{
		Name: "upf_smr_reject",
		Help: "The total number of rejected session modification requests",
	})

	UpfXdpAborted  prometheus.CounterFunc
	UpfXdpDrop     prometheus.CounterFunc
	UpfXdpPass     prometheus.CounterFunc
	UpfXdpTx       prometheus.CounterFunc
	UpfXdpRedirect prometheus.CounterFunc

	UpfRxArp      prometheus.CounterFunc
	UpfRxIcmp     prometheus.CounterFunc
	UpfRxIcmpv6   prometheus.CounterFunc
	UpfRxIp4      prometheus.CounterFunc
	UpfRxIp6      prometheus.CounterFunc
	UpfRxTcp      prometheus.CounterFunc
	UpfRxUdp      prometheus.CounterFunc
	UpfRxOther    prometheus.CounterFunc
	UpfRxGptEcho  prometheus.CounterFunc
	UpfRxGtpPdu   prometheus.CounterFunc
	UpfRxGtpOther prometheus.CounterFunc
	UpfRxGtpUnexp prometheus.CounterFunc

	UpfMessageProcessingDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "upf_message_processing_duration",
		Subsystem:  "pfcp",
		Help:       "Duration of the PFCP message processing",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"message_type"})
)

// PacketCounter usage
/*
	PacketCounter.WithLabelValues("SOMELABLEVALUE").Add(1)
	PacketCounter.WithLabelValues("SOMELABLEVALUE").Add(10)
	PacketCounter.WithLabelValues("SOMELABLEVALUE2").Add(1)
*/

func StartMetrics(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(addr, nil)
}

// Register eBPF metrics
func RegisterMetrics(stats UpfXdpActionStatistic) {

	// Metrics for the upf_xdp_statistic (xdp_action)

	UpfXdpAborted = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_xdp_aborted",
		Help: "The total number of aborted packets",
	}, func() float64 {
		return float64(stats.GetAborted())
	})

	UpfXdpDrop = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_xdp_drop",
		Help: "The total number of dropped packets",
	}, func() float64 {
		return float64(stats.GetDrop())
	})

	UpfXdpPass = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_xdp_pass",
		Help: "The total number of passed packets",
	}, func() float64 {
		return float64(stats.GetPass())
	})

	UpfXdpTx = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_xdp_tx",
		Help: "The total number of transmitted packets",
	}, func() float64 {
		return float64(stats.GetTx())
	})

	UpfXdpRedirect = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_xdp_redirect",
		Help: "The total number of redirected packets",
	}, func() float64 {
		return float64(stats.GetRedirect())
	})

	prometheus.MustRegister(UpfXdpAborted)
	prometheus.MustRegister(UpfXdpDrop)
	prometheus.MustRegister(UpfXdpPass)
	prometheus.MustRegister(UpfXdpTx)
	prometheus.MustRegister(UpfXdpRedirect)

	// Metrics for the upf_ext_stat (upf_counters)
	UpfRxArp = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_rx_arp",
		Help: "The total number of received ARP packets",
	}, func() float64 {
		return float64(stats.GetRxArp())
	})

	UpfRxIcmp = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_rx_icmp",
		Help: "The total number of received ICMP packets",
	}, func() float64 {
		return float64(stats.GetRxIcmp())
	})

	UpfRxIcmpv6 = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_rx_icmpv6",
		Help: "The total number of received ICMPv6 packets",
	}, func() float64 {
		return float64(stats.GetRxIcmp6())
	})

	UpfRxIp4 = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_rx_ip4",
		Help: "The total number of received IPv4 packets",
	}, func() float64 {
		return float64(stats.GetRxIp4())
	})

	UpfRxIp6 = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_rx_ip6",
		Help: "The total number of received IPv6 packets",
	}, func() float64 {
		return float64(stats.GetRxIp6())
	})

	UpfRxTcp = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_rx_tcp",
		Help: "The total number of received TCP packets",
	}, func() float64 {
		return float64(stats.GetRxTcp())
	})

	UpfRxUdp = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_rx_udp",
		Help: "The total number of received UDP packets",
	}, func() float64 {
		return float64(stats.GetRxUdp())
	})

	UpfRxOther = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_rx_other",
		Help: "The total number of received other packets",
	}, func() float64 {
		return float64(stats.GetRxOther())
	})

	UpfRxGptEcho = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_rx_gtp_echo",
		Help: "The total number of received GTP echo packets",
	}, func() float64 {
		return float64(stats.GetRxGtpEcho())
	})

	UpfRxGtpPdu = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_rx_gtp_pdu",
		Help: "The total number of received GTP PDU packets",
	}, func() float64 {
		return float64(stats.GetRxGtpPdu())
	})

	UpfRxGtpOther = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_rx_gtp_other",
		Help: "The total number of received GTP other packets",
	}, func() float64 {
		return float64(stats.GetRxGtpOther())
	})

	UpfRxGtpUnexp = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_rx_gtp_error",
		Help: "The total number of received GTP error packets",
	}, func() float64 {
		return float64(stats.GetRxGtpUnexp())
	})

	prometheus.MustRegister(UpfRxArp)
	prometheus.MustRegister(UpfRxIcmp)
	prometheus.MustRegister(UpfRxIcmpv6)
	prometheus.MustRegister(UpfRxIp4)
	prometheus.MustRegister(UpfRxIp6)
	prometheus.MustRegister(UpfRxTcp)
	prometheus.MustRegister(UpfRxUdp)
	prometheus.MustRegister(UpfRxOther)
	prometheus.MustRegister(UpfRxGptEcho)
	prometheus.MustRegister(UpfRxGtpPdu)
	prometheus.MustRegister(UpfRxGtpOther)
	prometheus.MustRegister(UpfRxGtpUnexp)
}
