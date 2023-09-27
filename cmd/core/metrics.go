package core

import (
	"net/http"
	"time"

	"github.com/edgecomllc/eupf/cmd/ebpf"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	PfcpMessageRx = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "upf_pfcp_rx",
		Help: "The total number of received PFCP messages",
	}, []string{"message_name"})

	PfcpMessageTx = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "upf_pfcp_tx",
		Help: "The total number of transmitted PFCP messages",
	}, []string{"message_name"})

	PfcpMessageRxErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "upf_pfcp_rx_errors",
		Help: "The total number of received PFCP messages with cause code",
	}, []string{"message_name", "cause_code"})

	UpfXdpAborted       prometheus.CounterFunc
	UpfXdpDrop          prometheus.CounterFunc
	UpfXdpPass          prometheus.CounterFunc
	UpfXdpTx            prometheus.CounterFunc
	UpfXdpRedirect      prometheus.CounterFunc
	UpfPfcpSessions     prometheus.GaugeFunc
	UpfPfcpAssociations prometheus.GaugeFunc

	UpfRx = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "upf_rx",
		Help: "The total number of received packets",
	}, []string{"packet_type"})

	UpfMessageRxLatency = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "upf_pfcp_rx_latency",
		Subsystem:  "pfcp",
		Help:       "Duration of the PFCP message processing",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"message_type"})
)

func StartMetrics(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(addr, nil)
	return err
}

func RegisterMetrics(stats ebpf.UpfXdpActionStatistic, conn *PfcpConnection) {
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

	UpfPfcpSessions = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "upf_pfcp_sessions",
		Help: "The current number of PFCP sessions",
	}, func() float64 {
		return float64(conn.GetSessionCount())
	})

	UpfPfcpAssociations = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "upf_pfcp_associations",
		Help: "The current number of PFCP associations",
	}, func() float64 {
		return float64(conn.GetAssiciationCount())
	})

	// Register metrics
	prometheus.MustRegister(UpfXdpAborted)
	prometheus.MustRegister(UpfXdpDrop)
	prometheus.MustRegister(UpfXdpPass)
	prometheus.MustRegister(UpfXdpTx)
	prometheus.MustRegister(UpfXdpRedirect)
	prometheus.MustRegister(UpfPfcpSessions)
	prometheus.MustRegister(UpfPfcpAssociations)

	// Used for getting difference between two counters to increment the prometheus counter (counters cannot be written only incremented)
	var prevUpfCounters ebpf.UpfCounters
	go func() {
		time.Sleep(2 * time.Second)
		RxPacketCounters := stats.GetUpfExtStatField()
		UpfRx.WithLabelValues("Arp").Add(float64(RxPacketCounters.RxArp - prevUpfCounters.RxArp))
		UpfRx.WithLabelValues("Icmp").Add(float64(RxPacketCounters.RxIcmp - prevUpfCounters.RxIcmp))
		UpfRx.WithLabelValues("Icmp6").Add(float64(RxPacketCounters.RxIcmp6 - prevUpfCounters.RxIcmp6))
		UpfRx.WithLabelValues("Ip4").Add(float64(RxPacketCounters.RxIp4 - prevUpfCounters.RxIp4))
		UpfRx.WithLabelValues("Ip6").Add(float64(RxPacketCounters.RxIp6 - prevUpfCounters.RxIp6))
		UpfRx.WithLabelValues("Tcp").Add(float64(RxPacketCounters.RxTcp - prevUpfCounters.RxTcp))
		UpfRx.WithLabelValues("Udp").Add(float64(RxPacketCounters.RxUdp - prevUpfCounters.RxUdp))
		UpfRx.WithLabelValues("Other").Add(float64(RxPacketCounters.RxOther - prevUpfCounters.RxOther))
		UpfRx.WithLabelValues("GtpEcho").Add(float64(RxPacketCounters.RxGtpEcho - prevUpfCounters.RxGtpEcho))
		UpfRx.WithLabelValues("GtpPdu").Add(float64(RxPacketCounters.RxGtpPdu - prevUpfCounters.RxGtpPdu))
		UpfRx.WithLabelValues("GtpOther").Add(float64(RxPacketCounters.RxGtpOther - prevUpfCounters.RxGtpOther))
		UpfRx.WithLabelValues("GtpUnexp").Add(float64(RxPacketCounters.RxGtpUnexp - prevUpfCounters.RxGtpUnexp))

		prevUpfCounters = RxPacketCounters
	}()
}
