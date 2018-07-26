package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	inFlightGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "requests_in_flight",
			Help: "A gauge of requests currently being served by the wrapped handler.",
		},
	)

	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)

	duration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: prometheus.ExponentialBuckets(.05, 2, 9),
		},
		[]string{"code"},
	)

	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "response_size_bytes",
			Help:    "A histogram of response sizes for requests.",
			Buckets: prometheus.ExponentialBuckets(256, 2, 14),
		},
		[]string{},
	)
)

func prometheusInit() {
	prometheus.MustRegister(inFlightGauge, counter, duration, responseSize)
}

func prometheusWrapHandler(handler http.Handler) http.Handler {
	handler = promhttp.InstrumentHandlerInFlight(inFlightGauge, handler)
	handler = promhttp.InstrumentHandlerCounter(counter, handler)
	handler = promhttp.InstrumentHandlerDuration(duration, handler)
	handler = promhttp.InstrumentHandlerResponseSize(responseSize, handler)
	return handler
}
