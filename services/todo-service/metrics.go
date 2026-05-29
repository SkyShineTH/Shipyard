package main

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "shipyard",
			Subsystem: "todo_service",
			Name:      "http_requests_total",
			Help:      "Total HTTP requests handled by todo-service.",
		},
		[]string{"method", "route", "status"},
	)
	httpRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "shipyard",
			Subsystem: "todo_service",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds for todo-service.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDurationSeconds)
}

func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())
		labels := []string{c.Request.Method, route, status}
		httpRequestsTotal.WithLabelValues(labels...).Inc()
		httpRequestDurationSeconds.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
	}
}
