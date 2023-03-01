package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Association setup requests
	AsrSucsess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "eupf_asr_sucsess",
		Help: "The total number of successful association setup requests",
	})
	AsrReject = promauto.NewCounter(prometheus.CounterOpts{
		Name: "eupf_asr_reject",
		Help: "The total number of rejected association setup requests",
	})

	// Session establishment requests
	SerSucsess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "eupf_ser_sucsess",
		Help: "The total number of successful session establishment requests",
	})
	SerReject = promauto.NewCounter(prometheus.CounterOpts{
		Name: "eupf_ser_reject",
		Help: "The total number of rejected session establishment requests",
	})
)

func StartMetrics(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(addr, nil)
}
