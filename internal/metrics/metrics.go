// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2025 Babak Farrokhi

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// QueryDuration tracks the duration of DNS queries
	QueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_query_duration_seconds",
			Help:    "Duration of DNS queries",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"domain", "server", "protocol"},
	)

	// QuerySuccess counts successful DNS queries
	QuerySuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_query_success_total",
			Help: "Total successful DNS queries",
		},
		[]string{"domain", "server", "protocol"},
	)

	// QueryFailures counts failed DNS queries
	QueryFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_query_failures_total",
			Help: "Total failed DNS queries",
		},
		[]string{"domain", "server", "protocol"},
	)
)

func init() {
	prometheus.MustRegister(QueryDuration, QuerySuccess, QueryFailures)
}

// RecordQuery records metrics for a DNS query
func RecordQuery(domain, server, protocol string, duration float64, success bool) {
	QueryDuration.WithLabelValues(domain, server, protocol).Observe(duration)
	if success {
		QuerySuccess.WithLabelValues(domain, server, protocol).Inc()
	} else {
		QueryFailures.WithLabelValues(domain, server, protocol).Inc()
	}
}
