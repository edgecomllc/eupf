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
		Name: "upf_ser_sucsess",
		Help: "The total number of successful session establishment requests",
	})
	SerReject = promauto.NewCounter(prometheus.CounterOpts{
		Name: "upf_ser_reject",
		Help: "The total number of rejected session establishment requests",
	})
)

func StartMetrics(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(addr, nil)
}
