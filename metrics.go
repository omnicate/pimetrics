package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	isUpMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pimetrics_is_up",
		Help: "Is pimetrics system is up",
	})
)
