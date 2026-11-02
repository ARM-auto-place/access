package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	CacheRequests      *prometheus.CounterVec
	RefreshRuns        *prometheus.CounterVec
	RefreshDuration    prometheus.Histogram
	RefreshLastSuccess prometheus.Gauge
	RequestLatency     *prometheus.HistogramVec
	HttpCollector      *prometheus.CounterVec
}

var (
	metricsCacheMu sync.Mutex
	metricsCache   = make(map[string]*Metrics)
)

func CreateMetrics(namespace, subsystem string, labels prometheus.Labels) *Metrics {
	cacheKey := namespace + "/" + subsystem

	metricsCacheMu.Lock()
	defer metricsCacheMu.Unlock()

	if cached, ok := metricsCache[cacheKey]; ok {
		return cached
	}

	requestLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "http_latency",
			Help:        "The API methods latency info",
			ConstLabels: labels,
		},
		[]string{"service", "method", "success"},
	)
	httpCollector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "http_collector",
			Help:        "The number of http requests",
			ConstLabels: labels,
		},
		[]string{"service", "method", "success"},
	)

	prometheus.MustRegister(httpCollector)
	prometheus.MustRegister(requestLatency)

	cacheRequests := prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: namespace, Subsystem: subsystem, Name: "cache_requests_total", Help: "Role cache lookups", ConstLabels: labels}, []string{"result"})
	refreshRuns := prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: namespace, Subsystem: subsystem, Name: "cache_refresh_total", Help: "Cache refresh runs", ConstLabels: labels}, []string{"success"})
	refreshDuration := prometheus.NewHistogram(prometheus.HistogramOpts{Namespace: namespace, Subsystem: subsystem, Name: "cache_refresh_seconds", Help: "Cache refresh duration", ConstLabels: labels})
	refreshLastSuccess := prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Subsystem: subsystem, Name: "cache_refresh_last_success_timestamp", Help: "Last published role snapshot", ConstLabels: labels})
	prometheus.MustRegister(cacheRequests, refreshRuns, refreshDuration, refreshLastSuccess)
	created := &Metrics{
		CacheRequests: cacheRequests, RefreshRuns: refreshRuns, RefreshDuration: refreshDuration, RefreshLastSuccess: refreshLastSuccess,
		HttpCollector:  httpCollector,
		RequestLatency: requestLatency,
	}
	metricsCache[cacheKey] = created
	return created
}

func NewMetrics(serviceName string) *Metrics {
	return CreateMetrics("arm", serviceName, nil)
}
