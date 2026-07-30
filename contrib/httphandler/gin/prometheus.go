package gin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bpcoder16/Chestnut/v4/appconfig"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	httpWebSocketConnectionsName   = "websocket_connections"
	httpRecoveredPanicsMetricName  = "recovered_panics_total"

	serviceLabelName     = "service"
	methodLabelName      = "method"
	routeLabelName       = "route"
	statusClassLabelName = "status_class"
)

// 保持默认 bucket 数量不变，用 800ms 监控边界替换对 API 延迟价值较低的 5ms 边界。
var httpRequestDurationBuckets = []float64{
	0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 0.8, 1, 2.5, 5, 10,
}

type prometheusHTTPMetrics struct {
	excludedPaths        map[string]struct{}
	excludedStatusCodes  map[int]struct{}
	requests             *prometheus.CounterVec
	requestDuration      *prometheus.HistogramVec
	requestsInFlight     prometheus.Gauge
	webSocketConnections *prometheus.GaugeVec
	recoveredPanics      *prometheus.CounterVec
	gatherer             prometheus.Gatherer
}

func newPrometheusHTTPMetrics(config appconfig.Prometheus) *prometheusHTTPMetrics {
	constLabels := prometheus.Labels{serviceLabelName: config.ServiceName}
	httpRequestLabels := []string{methodLabelName, routeLabelName, statusClassLabelName}
	httpRequestDurationLabels := []string{methodLabelName, routeLabelName}
	registry := prometheus.NewRegistry()
	metrics := &prometheusHTTPMetrics{
		excludedPaths:       make(map[string]struct{}, len(config.ExcludedPaths)+1),
		excludedStatusCodes: make(map[int]struct{}, len(config.ExcludedStatusCodes)),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   prometheusNamespace,
			Subsystem:   prometheusHTTPSubsystem,
			Name:        httpRequestsMetricName,
			Help:        "Total number of completed HTTP requests handled by the Chestnut server, grouped by HTTP status class.",
			ConstLabels: constLabels,
		}, httpRequestLabels),
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   prometheusNamespace,
			Subsystem:   prometheusHTTPSubsystem,
			Name:        httpRequestDurationMetricName,
			Help:        "Duration in seconds of completed HTTP requests handled by the Chestnut server, excluding WebSocket upgrade requests.",
			ConstLabels: constLabels,
			Buckets:     httpRequestDurationBuckets,
		}, httpRequestDurationLabels),
		requestsInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   prometheusNamespace,
			Subsystem:   prometheusHTTPSubsystem,
			Name:        httpRequestsInFlightMetricName,
			Help:        "Current number of matched HTTP requests being handled by the Chestnut server, including active WebSocket handlers.",
			ConstLabels: constLabels,
		}),
		webSocketConnections: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   prometheusNamespace,
			Subsystem:   prometheusHTTPSubsystem,
			Name:        httpWebSocketConnectionsName,
			Help:        "Current number of active WebSocket connections and in-progress upgrade requests handled by the Chestnut server, grouped by route.",
			ConstLabels: constLabels,
		}, []string{routeLabelName}),
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
		metrics.webSocketConnections,
		metrics.recoveredPanics,
	)
	metrics.gatherer = prometheus.Gatherers{registry, prometheus.DefaultGatherer}
	return metrics
}

func (m *prometheusHTTPMetrics) initializeRecoveredPanicRoutes(routes gin.RoutesInfo) {
	for _, route := range routes {
		if m.isExcludedPath(route.Path) {
			continue
		}
		// 在服务开始接收请求前暴露零值，为 Prometheus 抓取首次 Panic 的 0→1 增量提供基线。
		m.recoveredPanics.WithLabelValues(route.Method, route.Path)
	}
}

func (m *prometheusHTTPMetrics) middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		route := ctx.FullPath()
		if route == "" || m.isExcludedPath(ctx.Request.URL.Path) {
			ctx.Next()
			return
		}

		webSocketUpgrade := websocket.IsWebSocketUpgrade(ctx.Request)
		startedAt := time.Now()
		m.requestsInFlight.Inc()
		if webSocketUpgrade {
			m.webSocketConnections.WithLabelValues(route).Inc()
		}
		defer func() {
			m.requestsInFlight.Dec()
			if webSocketUpgrade {
				m.webSocketConnections.WithLabelValues(route).Dec()
			}

			// Recovery 是独立异常事实，即使最终状态被排除也必须保留该信号。
			if recovered, exists := ctx.Get(recoveryMetricContextKey); exists && recovered == true {
				m.recoveredPanics.WithLabelValues(ctx.Request.Method, route).Inc()
			}

			statusCode := ctx.Writer.Status()
			if m.isExcludedStatusCode(statusCode) {
				return
			}

			statusClass := strconv.Itoa(statusCode/100) + "xx"
			m.requests.WithLabelValues(ctx.Request.Method, route, statusClass).Inc()
			// WebSocket Handler 在连接关闭后才返回，记录该耗时会把连接存活时间误当成 HTTP 响应延迟。
			if webSocketUpgrade {
				return
			}
			m.requestDuration.WithLabelValues(ctx.Request.Method, route).Observe(time.Since(startedAt).Seconds())
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
