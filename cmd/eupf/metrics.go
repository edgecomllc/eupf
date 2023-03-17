package main

import (
	"log"
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

	// PacketCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	// 	Name: "upf_packet_counter",
	// 	Help: "The total number of packets",
	// }, []string{"label"}) // here we can add more labels to the metric
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
func RegisterMetrics(bpfObjects *BpfObjects) {
	aborted, drop, pass, tx, redirect := CreateEbpfGetStats(bpfObjects)
	UpfXdpAborted = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_xdp_aborted",
		Help: "The total number of aborted packets",
	}, func() float64 {
		res, err := aborted()
		if err != nil {
			log.Println("XDP Stats: aborted", err)
			return 0
		}
		return float64(res)
	})

	UpfXdpDrop = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_xdp_drop",
		Help: "The total number of dropped packets",
	}, func() float64 {
		res, err := drop()
		if err != nil {
			log.Println("XDP Stats: drop", err)
			return 0
		}
		return float64(res)
	})

	UpfXdpPass = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_xdp_pass",
		Help: "The total number of passed packets",
	}, func() float64 {
		res, err := pass()
		if err != nil {
			log.Println("XDP Stats: pass", err)
			return 0
		}
		return float64(res)
	})

	UpfXdpTx = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_xdp_tx",
		Help: "The total number of transmitted packets",
	}, func() float64 {
		res, err := tx()
		if err != nil {
			log.Println("XDP Stats: tx", err)
			return 0
		}
		return float64(res)
	})

	UpfXdpRedirect = prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name: "upf_xdp_redirect",
		Help: "The total number of redirected packets",
	}, func() float64 {
		res, err := redirect()
		if err != nil {
			log.Println("XDP Stats: redirect", err)
			return 0
		}
		return float64(res)
	})

	prometheus.MustRegister(UpfXdpAborted)
	prometheus.MustRegister(UpfXdpDrop)
	prometheus.MustRegister(UpfXdpPass)
	prometheus.MustRegister(UpfXdpTx)
	prometheus.MustRegister(UpfXdpRedirect)
}
