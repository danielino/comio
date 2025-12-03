package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "comio_requests_total",
			Help: "Total number of requests",
		},
		[]string{"method", "bucket", "status"},
	)
	
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "comio_request_duration_seconds",
			Help: "Request duration in seconds",
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(RequestsTotal)
	prometheus.MustRegister(RequestDuration)
}
