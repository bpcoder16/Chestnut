package gin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bpcoder16/Chestnut/v4/appconfig"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	prometheusMetricsPath    = "/metrics"
	recoveryMetricContextKey = "chestnut.gin.recovered_panic"
	prometheusNamespace      = "chestnut"
	prometheusHTTPSubsystem  = "http_server"

	httpRequestsMetricName         = "requests_total"
	httpRequestDurationMetricName  = "request_duration_seconds"
	httpRequestsInFlightMetricName = "requests_in_flight"
	httpRecoveredPanicsMetricName  = "recovered_panics_total"

	serviceLabelName     = "service"
	methodLabelName      = "method"
	routeLabelName       = "route"
	statusLabelName      = "status"
	statusClassLabelName = "status_class"
)

type prometheusHTTPMetrics struct {
	excludedPaths       map[string]struct{}
	excludedStatusCodes map[int]struct{}
	requests            *prometheus.CounterVec
	requestDuration     *prometheus.HistogramVec
	requestsInFlight    prometheus.Gauge
	recoveredPanics     *prometheus.CounterVec
	gatherer            prometheus.Gatherer
}

func newPrometheusHTTPMetrics(config appconfig.Prometheus) *prometheusHTTPMetrics {
	constLabels := prometheus.Labels{serviceLabelName: config.ServiceName}
	httpRequestLabels := []string{methodLabelName, routeLabelName, statusLabelName}
	httpRequestDurationLabels := []string{methodLabelName, routeLabelName, statusClassLabelName}
	registry := prometheus.NewRegistry()
	metrics := &prometheusHTTPMetrics{
		excludedPaths:       make(map[string]struct{}, len(config.ExcludedPaths)+1),
		excludedStatusCodes: make(map[int]struct{}, len(config.ExcludedStatusCodes)),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   prometheusNamespace,
			Subsystem:   prometheusHTTPSubsystem,
			Name:        httpRequestsMetricName,
			Help:        "Total number of completed HTTP requests handled by the Chestnut server.",
			ConstLabels: constLabels,
		}, httpRequestLabels),
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   prometheusNamespace,
			Subsystem:   prometheusHTTPSubsystem,
			Name:        httpRequestDurationMetricName,
			Help:        "Duration in seconds of completed HTTP requests handled by the Chestnut server.",
			ConstLabels: constLabels,
			Buckets:     prometheus.DefBuckets,
		}, httpRequestDurationLabels),
		requestsInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   prometheusNamespace,
			Subsystem:   prometheusHTTPSubsystem,
			Name:        httpRequestsInFlightMetricName,
			Help:        "Current number of matched HTTP requests being handled by the Chestnut server.",
			ConstLabels: constLabels,
		}),
		recoveredPanics: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   prometheusNamespace,
			Subsystem:   prometheusHTTPSubsystem,
			Name:        httpRecoveredPanicsMetricName,
			Help:        "Total number of HTTP requests recovered from panics by Chestnut.",
			ConstLabels: constLabels,
		}, []string{methodLabelName, routeLabelName}),
	}

	metrics.excludedPaths[prometheusMetricsPath] = struct{}{}
	for _, excludedPath := range config.ExcludedPaths {
		metrics.excludedPaths[excludedPath] = struct{}{}
	}
	for _, statusCode := range config.ExcludedStatusCodes {
		metrics.excludedStatusCodes[statusCode] = struct{}{}
	}

	registry.MustRegister(
		metrics.requests,
		metrics.requestDuration,
		metrics.requestsInFlight,
		metrics.recoveredPanics,
	)
	metrics.gatherer = prometheus.Gatherers{registry, prometheus.DefaultGatherer}
	return metrics
}

func (m *prometheusHTTPMetrics) middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		route := ctx.FullPath()
		if route == "" || m.isExcludedPath(ctx.Request.URL.Path) {
			ctx.Next()
			return
		}

		startedAt := time.Now()
		m.requestsInFlight.Inc()
		defer func() {
			m.requestsInFlight.Dec()

			// Recovery 是独立异常事实，即使最终状态被排除也必须保留该信号。
			if recovered, exists := ctx.Get(recoveryMetricContextKey); exists && recovered == true {
				m.recoveredPanics.WithLabelValues(ctx.Request.Method, route).Inc()
			}

			statusCode := ctx.Writer.Status()
			if m.isExcludedStatusCode(statusCode) {
				return
			}

			status := strconv.Itoa(statusCode)
			m.requests.WithLabelValues(ctx.Request.Method, route, status).Inc()
			// 延迟只保留状态类别，避免每个精确状态码复制整套 Histogram buckets。
			statusClass := strconv.Itoa(statusCode/100) + "xx"
			m.requestDuration.WithLabelValues(ctx.Request.Method, route, statusClass).Observe(time.Since(startedAt).Seconds())
		}()

		ctx.Next()
	}
}

func (m *prometheusHTTPMetrics) handler() http.Handler {
	return promhttp.HandlerFor(m.gatherer, promhttp.HandlerOpts{})
}

func (m *prometheusHTTPMetrics) isExcludedPath(requestPath string) bool {
	_, excluded := m.excludedPaths[requestPath]
	return excluded
}

func (m *prometheusHTTPMetrics) isExcludedStatusCode(statusCode int) bool {
	_, excluded := m.excludedStatusCodes[statusCode]
	return excluded
}
