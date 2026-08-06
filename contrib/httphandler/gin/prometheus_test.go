package gin

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bpcoder16/Chestnut/v4/appconfig"
	ginframework "github.com/gin-gonic/gin"
)

const (
	testMetricsServiceName = "test-api"
	testRequestTimeout     = 5 * time.Second
)

func TestHTTPHandlersKeepPrometheusDisabledByDefault(t *testing.T) {
	handlers := map[string]*ginframework.Engine{
		"legacy":                 HTTPHandler(),
		"legacy with middleware": HTTPHandlerWithMiddlewares(nil),
		"config aware":           HTTPHandlerWithConfig(&appconfig.AppConfig{}, nil),
	}

	for name, handler := range handlers {
		t.Run(name, func(t *testing.T) {
			recorder := performRequest(handler, http.MethodGet, "/metrics")
			if recorder.Code != http.StatusNotFound {
				t.Fatalf("GET /metrics status = %d, want %d", recorder.Code, http.StatusNotFound)
			}
		})
	}
}

func TestPrometheusRequestDurationBuckets(t *testing.T) {
	want := []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 0.8, 1, 2.5, 5, 10}
	if len(httpRequestDurationBuckets) != len(want) {
		t.Fatalf("request duration bucket count = %d, want %d", len(httpRequestDurationBuckets), len(want))
	}
	for index := range want {
		if httpRequestDurationBuckets[index] != want[index] {
			t.Fatalf("request duration bucket[%d] = %v, want %v", index, httpRequestDurationBuckets[index], want[index])
		}
	}
}

func TestPrometheusEndpointExposesEngineAndDefaultMetrics(t *testing.T) {
	handler := newPrometheusTestEngine(testPrometheusConfig(), func(router *ginframework.RouterGroup) {
		router.GET("/websocket-probe", func(ctx *ginframework.Context) {
			ctx.Status(http.StatusOK)
		})
	})
	performRequest(handler, http.MethodGet, "/test/123")
	performRequest(handler, http.MethodGet, "/panic")
	performWebSocketUpgrade(handler, "/websocket-probe")
	metrics := scrapeMetrics(t, handler)

	for _, metricHelp := range []string{
		"# HELP chestnut_http_server_requests_total",
		"# HELP chestnut_http_server_request_duration_seconds",
		"# HELP chestnut_http_server_requests_in_flight",
		"# HELP chestnut_http_server_websocket_connections",
		"# HELP chestnut_http_server_recovered_panics_total",
		"# HELP chestnut_http_server_business_responses_total",
		"# HELP go_goroutines",
		"# HELP go_memstats_alloc_bytes_total",
		"# HELP go_sched_gomaxprocs_threads",
		"# HELP process_cpu_seconds_total",
	} {
		if !strings.Contains(metrics, metricHelp) {
			t.Fatalf("GET /metrics body does not contain %q", metricHelp)
		}
	}
}

func TestPrometheusInitializesMetricsForRegisteredRoutes(t *testing.T) {
	handler := newPrometheusTestEngine(testPrometheusConfig(), func(router *ginframework.RouterGroup) {
		router.POST("/write", func(ctx *ginframework.Context) {
			ctx.Status(http.StatusNoContent)
		})
	})

	metrics := scrapeMetrics(t, handler)
	for _, route := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/panic"},
		{method: http.MethodGet, path: "/active-500"},
		{method: http.MethodGet, path: "/test/:id"},
		{method: http.MethodPost, path: "/write"},
	} {
		for _, statusClass := range httpStatusClasses {
			assertMetricValue(t, metrics, "chestnut_http_server_requests_total", 0,
				`service="test-api"`, `method="`+route.method+`"`, `route="`+route.path+`"`,
				`status_class="`+statusClass+`"`)
		}
		assertMetricValue(t, metrics, "chestnut_http_server_request_duration_seconds_count", 0,
			`service="test-api"`, `method="`+route.method+`"`, `route="`+route.path+`"`)
		assertMetricValue(t, metrics, "chestnut_http_server_recovered_panics_total", 0,
			`service="test-api"`, `method="`+route.method+`"`, `route="`+route.path+`"`)
		assertMetricValue(t, metrics, "chestnut_http_server_business_responses_total", 0,
			`service="test-api"`, `method="`+route.method+`"`, `route="`+route.path+`"`, `code="400"`)
	}
	for _, excludedRoute := range []string{"/metrics", "/health", "/ready"} {
		for _, metricName := range []string{
			"chestnut_http_server_requests_total",
			"chestnut_http_server_request_duration_seconds_count",
			"chestnut_http_server_recovered_panics_total",
			"chestnut_http_server_business_responses_total",
		} {
			assertMetricAbsent(t, metrics, metricName, `route="`+excludedRoute+`"`)
		}
	}
}

func TestPrometheusRecordsRouteTemplateHTTPStatusClassAndBusinessCode(t *testing.T) {
	handler := newPrometheusTestEngine(testPrometheusConfig(), nil)

	created := performRequest(handler, http.MethodGet, "/test/123")
	if created.Code != http.StatusCreated || created.Body.String() != "created" {
		t.Fatalf("GET /test/123 = (%d, %q), want (%d, %q)", created.Code, created.Body.String(), http.StatusCreated, "created")
	}
	ok := performRequest(handler, http.MethodGet, "/test/ok")
	if ok.Code != http.StatusOK || ok.Body.String() != "ok" {
		t.Fatalf("GET /test/ok = (%d, %q), want (%d, %q)", ok.Code, ok.Body.String(), http.StatusOK, "ok")
	}
	businessError := performRequest(handler, http.MethodGet, "/business-error")
	if businessError.Code != http.StatusOK {
		t.Fatalf("GET /business-error status = %d, want %d", businessError.Code, http.StatusOK)
	}

	metrics := scrapeMetrics(t, handler)
	assertMetricValue(t, metrics, "chestnut_http_server_requests_total", 2,
		`service="test-api"`, `method="GET"`, `route="/test/:id"`, `status_class="2xx"`)
	assertMetricValue(t, metrics, "chestnut_http_server_request_duration_seconds_count", 2,
		`service="test-api"`, `method="GET"`, `route="/test/:id"`)
	assertMetricAbsent(t, metrics, "chestnut_http_server_requests_total", `status=`)
	assertMetricAbsent(t, metrics, "chestnut_http_server_request_duration_seconds_count", `status=`)
	assertMetricAbsent(t, metrics, "chestnut_http_server_request_duration_seconds_count", `status_class=`)
	assertMetricValue(t, metrics, "chestnut_http_server_requests_total", 1,
		`service="test-api"`, `method="GET"`, `route="/business-error"`, `status_class="2xx"`)
	assertMetricValue(t, metrics, "chestnut_http_server_business_responses_total", 1,
		`service="test-api"`, `method="GET"`, `route="/business-error"`, `code="400"`)

	for _, forbidden := range []string{"/test/123", "/test/ok", "服务异常"} {
		if strings.Contains(metrics, forbidden) {
			t.Fatalf("GET /metrics body contains forbidden request/body value %q", forbidden)
		}
	}
}

func TestPrometheusTracksOnlyTopLevelNumericBusinessCode400WithinBodyLimit(t *testing.T) {
	handler := newPrometheusTestEngine(testPrometheusConfig(), func(router *ginframework.RouterGroup) {
		router.GET("/business-code-string", func(ctx *ginframework.Context) {
			ctx.JSON(http.StatusOK, ginframework.H{"code": "400"})
		})
		router.GET("/nested-business-code", func(ctx *ginframework.Context) {
			ctx.JSON(http.StatusOK, ginframework.H{"data": ginframework.H{"code": 400}})
		})
		router.GET("/other-business-code", func(ctx *ginframework.Context) {
			ctx.JSON(http.StatusOK, ginframework.H{"code": 401})
		})
		router.GET("/large-business-error", func(ctx *ginframework.Context) {
			ctx.JSON(http.StatusOK, ginframework.H{"code": 400, "data": strings.Repeat("x", prometheusResponseBodyLimit)})
		})
	})

	for _, route := range []string{
		"/business-code-string", "/nested-business-code", "/other-business-code", "/large-business-error",
	} {
		response := performRequest(handler, http.MethodGet, route)
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", route, response.Code, http.StatusOK)
		}
	}

	metrics := scrapeMetrics(t, handler)
	for _, route := range []string{
		"/business-code-string", "/nested-business-code", "/other-business-code", "/large-business-error",
	} {
		assertMetricValue(t, metrics, "chestnut_http_server_business_responses_total", 0,
			`service="test-api"`, `method="GET"`, `route="`+route+`"`, `code="400"`)
	}
}

func TestPrometheusTracksWebSocketConnectionsByRouteAndExcludesThemFromHTTPRequestDuration(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	handler := newPrometheusTestEngine(testPrometheusConfig(), func(router *ginframework.RouterGroup) {
		router.GET("/app/websocket/:id", func(ctx *ginframework.Context) {
			close(started)
			<-release
			ctx.Status(http.StatusOK)
		})
		router.GET("/admin/websocket/:id", func(ctx *ginframework.Context) {
			ctx.Status(http.StatusOK)
		})
	})

	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		done <- performWebSocketUpgrade(handler, "/app/websocket/123")
	}()

	select {
	case <-started:
	case <-time.After(testRequestTimeout):
		t.Fatal("WebSocket handler did not start before timeout")
	}
	adminResponse := performWebSocketUpgrade(handler, "/admin/websocket/456")
	if adminResponse.Code != http.StatusOK {
		t.Fatalf("GET /admin/websocket/456 status = %d, want %d", adminResponse.Code, http.StatusOK)
	}

	activeMetrics := scrapeMetrics(t, handler)
	assertMetricValue(t, activeMetrics, "chestnut_http_server_requests_in_flight", 1, `service="test-api"`)
	assertMetricValue(t, activeMetrics, "chestnut_http_server_websocket_connections", 1,
		`service="test-api"`, `route="/app/websocket/:id"`)
	assertMetricValue(t, activeMetrics, "chestnut_http_server_websocket_connections", 0,
		`service="test-api"`, `route="/admin/websocket/:id"`)

	close(release)
	select {
	case recorder := <-done:
		if recorder.Code != http.StatusOK {
			t.Fatalf("GET /app/websocket/123 status = %d, want %d", recorder.Code, http.StatusOK)
		}
	case <-time.After(testRequestTimeout):
		t.Fatal("WebSocket handler did not finish before timeout")
	}

	metrics := scrapeMetrics(t, handler)
	for _, route := range []string{"/app/websocket/:id", "/admin/websocket/:id"} {
		assertMetricValue(t, metrics, "chestnut_http_server_requests_total", 1,
			`service="test-api"`, `method="GET"`, `route="`+route+`"`, `status_class="2xx"`)
		assertMetricValue(t, metrics, "chestnut_http_server_request_duration_seconds_count", 0,
			`service="test-api"`, `method="GET"`, `route="`+route+`"`)
		assertMetricValue(t, metrics, "chestnut_http_server_websocket_connections", 0,
			`service="test-api"`, `route="`+route+`"`)
	}
	assertMetricValue(t, metrics, "chestnut_http_server_requests_in_flight", 0, `service="test-api"`)
	for _, rawPath := range []string{"/app/websocket/123", "/admin/websocket/456"} {
		if strings.Contains(metrics, rawPath) {
			t.Fatalf("GET /metrics body contains raw WebSocket path %q", rawPath)
		}
	}
}

func TestPrometheusExcludesUnmatchedConfiguredAndStatusRequests(t *testing.T) {
	config := testPrometheusConfig()
	// /metrics must stay excluded by the framework even if operators omit it.
	config.Prometheus.ExcludedPaths = []string{"/health", "/ready"}
	handler := newPrometheusTestEngine(config, nil)

	requests := []string{"/missing/123", "/matched-not-found", "/metrics", "/health", "/ready"}
	for _, requestPath := range requests {
		performRequest(handler, http.MethodGet, requestPath)
	}

	metrics := scrapeMetrics(t, handler)
	for _, forbidden := range []string{
		`route="/metrics"`,
		`route="/health"`,
		`route="/ready"`,
		`route="unknown"`,
		"/missing/123",
	} {
		if strings.Contains(metrics, forbidden) {
			t.Fatalf("GET /metrics body contains excluded route/path %q", forbidden)
		}
	}
	assertMetricValue(t, metrics, "chestnut_http_server_requests_total", 0,
		`service="test-api"`, `method="GET"`, `route="/matched-not-found"`, `status_class="4xx"`)
	assertMetricValue(t, metrics, "chestnut_http_server_request_duration_seconds_count", 0,
		`service="test-api"`, `method="GET"`, `route="/matched-not-found"`)
	assertMetricValue(t, metrics, "chestnut_http_server_recovered_panics_total", 0,
		`service="test-api"`, `method="GET"`, `route="/matched-not-found"`)
	assertMetricValue(t, metrics, "chestnut_http_server_requests_in_flight", 0, `service="test-api"`)
}

func TestPrometheusExcludesConfiguredRouteTemplate(t *testing.T) {
	config := testPrometheusConfig()
	config.Prometheus.ExcludedPaths = append(config.Prometheus.ExcludedPaths, "/api/*path")
	handler := newPrometheusTestEngine(config, func(router *ginframework.RouterGroup) {
		router.GET("/api/*path", func(ctx *ginframework.Context) {
			ctx.Status(http.StatusNoContent)
		})
	})

	response := performRequest(handler, http.MethodGet, "/api/users/123")
	if response.Code != http.StatusNoContent {
		t.Fatalf("GET /api/users/123 status = %d, want %d", response.Code, http.StatusNoContent)
	}

	metrics := scrapeMetrics(t, handler)
	for _, metricName := range []string{
		"chestnut_http_server_requests_total",
		"chestnut_http_server_request_duration_seconds_count",
		"chestnut_http_server_recovered_panics_total",
	} {
		assertMetricAbsent(t, metrics, metricName, `route="/api/*path"`)
	}
	if strings.Contains(metrics, "/api/users/123") {
		t.Fatal("GET /metrics body contains excluded raw request path")
	}
}

func TestPrometheusRecordsRecoveredPanicSeparatelyFromActiveHTTP500(t *testing.T) {
	handler := newPrometheusTestEngine(testPrometheusConfig(), nil)
	baselineMetrics := scrapeMetrics(t, handler)
	assertMetricValue(t, baselineMetrics, "chestnut_http_server_recovered_panics_total", 0,
		`service="test-api"`, `method="GET"`, `route="/panic"`)

	panicResponse := performRequest(handler, http.MethodGet, "/panic")
	if panicResponse.Code != http.StatusInternalServerError {
		t.Fatalf("GET /panic status = %d, want %d", panicResponse.Code, http.StatusInternalServerError)
	}
	activeResponse := performRequest(handler, http.MethodGet, "/active-500")
	if activeResponse.Code != http.StatusInternalServerError {
		t.Fatalf("GET /active-500 status = %d, want %d", activeResponse.Code, http.StatusInternalServerError)
	}

	metrics := scrapeMetrics(t, handler)
	for _, route := range []string{"/panic", "/active-500"} {
		assertMetricValue(t, metrics, "chestnut_http_server_requests_total", 1,
			`service="test-api"`, `method="GET"`, `route="`+route+`"`, `status_class="5xx"`)
		assertMetricValue(t, metrics, "chestnut_http_server_request_duration_seconds_count", 1,
			`service="test-api"`, `method="GET"`, `route="`+route+`"`)
	}
	assertMetricValue(t, metrics, "chestnut_http_server_recovered_panics_total", 1,
		`service="test-api"`, `method="GET"`, `route="/panic"`)
	assertMetricValue(t, metrics, "chestnut_http_server_recovered_panics_total", 0,
		`service="test-api"`, `method="GET"`, `route="/active-500"`)
	assertMetricValue(t, metrics, "chestnut_http_server_requests_in_flight", 0, `service="test-api"`)
}

func TestPrometheusPreservesCommittedResponseAfterRecoveredPanic(t *testing.T) {
	const (
		responseHeaderName  = "X-Recovery-Test"
		responseHeaderValue = "committed"
		responseBody        = "response already committed"
	)

	handler := newPrometheusTestEngine(testPrometheusConfig(), func(router *ginframework.RouterGroup) {
		router.GET("/committed-panic", func(ctx *ginframework.Context) {
			ctx.Header(responseHeaderName, responseHeaderValue)
			ctx.String(http.StatusOK, responseBody)
			panic("panic after committed response")
		})
	})

	response := performRequest(handler, http.MethodGet, "/committed-panic")
	if response.Code != http.StatusOK {
		t.Fatalf("GET /committed-panic status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get(responseHeaderName); got != responseHeaderValue {
		t.Fatalf("GET /committed-panic %s = %q, want %q", responseHeaderName, got, responseHeaderValue)
	}
	if got := response.Body.String(); got != responseBody {
		t.Fatalf("GET /committed-panic body = %q, want %q", got, responseBody)
	}

	metrics := scrapeMetrics(t, handler)
	assertMetricValue(t, metrics, "chestnut_http_server_requests_total", 1,
		`service="test-api"`, `method="GET"`, `route="/committed-panic"`, `status_class="2xx"`)
	assertMetricValue(t, metrics, "chestnut_http_server_request_duration_seconds_count", 1,
		`service="test-api"`, `method="GET"`, `route="/committed-panic"`)
	assertMetricValue(t, metrics, "chestnut_http_server_recovered_panics_total", 1,
		`service="test-api"`, `method="GET"`, `route="/committed-panic"`)
	assertMetricValue(t, metrics, "chestnut_http_server_requests_in_flight", 0, `service="test-api"`)
}

func TestPrometheusRecordsRecoveryForExcludedCommittedStatus(t *testing.T) {
	const responseBody = "not found before panic"

	handler := newPrometheusTestEngine(testPrometheusConfig(), func(router *ginframework.RouterGroup) {
		router.GET("/committed-not-found-panic", func(ctx *ginframework.Context) {
			ctx.String(http.StatusNotFound, responseBody)
			panic("panic after committed excluded status")
		})
	})

	response := performRequest(handler, http.MethodGet, "/committed-not-found-panic")
	if response.Code != http.StatusNotFound {
		t.Fatalf("GET /committed-not-found-panic status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if got := response.Body.String(); got != responseBody {
		t.Fatalf("GET /committed-not-found-panic body = %q, want %q", got, responseBody)
	}

	metrics := scrapeMetrics(t, handler)
	assertMetricValue(t, metrics, "chestnut_http_server_requests_total", 0,
		`service="test-api"`, `method="GET"`, `route="/committed-not-found-panic"`, `status_class="4xx"`)
	assertMetricValue(t, metrics, "chestnut_http_server_request_duration_seconds_count", 0,
		`service="test-api"`, `method="GET"`, `route="/committed-not-found-panic"`)
	assertMetricValue(t, metrics, "chestnut_http_server_recovered_panics_total", 1,
		`service="test-api"`, `method="GET"`, `route="/committed-not-found-panic"`)
	assertMetricValue(t, metrics, "chestnut_http_server_requests_in_flight", 0, `service="test-api"`)
}

func TestPrometheusTracksInFlightRequestUntilCompletion(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	handler := newPrometheusTestEngine(testPrometheusConfig(), func(router *ginframework.RouterGroup) {
		router.GET("/blocked", func(ctx *ginframework.Context) {
			close(started)
			<-release
			ctx.Status(http.StatusNoContent)
		})
	})

	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		done <- performRequest(handler, http.MethodGet, "/blocked")
	}()

	select {
	case <-started:
	case <-time.After(testRequestTimeout):
		t.Fatal("blocked handler did not start before timeout")
	}
	assertMetricValue(t, scrapeMetrics(t, handler), "chestnut_http_server_requests_in_flight", 1, `service="test-api"`)
	assertMetricAbsent(t, scrapeMetrics(t, handler), "chestnut_http_server_websocket_connections", `service="test-api"`)

	close(release)
	select {
	case response := <-done:
		if response.Code != http.StatusNoContent {
			t.Fatalf("GET /blocked status = %d, want %d", response.Code, http.StatusNoContent)
		}
	case <-time.After(testRequestTimeout):
		t.Fatal("blocked handler did not finish before timeout")
	}
	assertMetricValue(t, scrapeMetrics(t, handler), "chestnut_http_server_requests_in_flight", 0, `service="test-api"`)
}

func TestPrometheusRegistriesAreIsolatedPerEngine(t *testing.T) {
	engineA := newPrometheusTestEngine(testPrometheusConfig(), nil)
	engineB := newPrometheusTestEngine(testPrometheusConfig(), nil)

	response := performRequest(engineA, http.MethodGet, "/test/123")
	if response.Code != http.StatusCreated {
		t.Fatalf("engine A GET /test/123 status = %d, want %d", response.Code, http.StatusCreated)
	}

	assertMetricValue(t, scrapeMetrics(t, engineA), "chestnut_http_server_requests_total", 1,
		`service="test-api"`, `method="GET"`, `route="/test/:id"`, `status_class="2xx"`)
	assertMetricValue(t, scrapeMetrics(t, engineB), "chestnut_http_server_requests_total", 0,
		`service="test-api"`, `method="GET"`, `route="/test/:id"`, `status_class="2xx"`)
}

func testPrometheusConfig() *appconfig.AppConfig {
	return &appconfig.AppConfig{
		Prometheus: appconfig.Prometheus{
			Enabled:             true,
			ServiceName:         testMetricsServiceName,
			ExcludedPaths:       []string{"/metrics", "/health", "/ready"},
			ExcludedStatusCodes: []int{http.StatusNotFound},
		},
	}
}

func newPrometheusTestEngine(config *appconfig.AppConfig, extraRoutes func(*ginframework.RouterGroup)) *ginframework.Engine {
	return HTTPHandlerWithConfig(config, nil, func(router *ginframework.RouterGroup) {
		router.GET("/test/:id", func(ctx *ginframework.Context) {
			if ctx.Param("id") == "ok" {
				ctx.String(http.StatusOK, "ok")
				return
			}
			ctx.String(http.StatusCreated, "created")
		})
		router.GET("/business-error", func(ctx *ginframework.Context) {
			ctx.JSON(http.StatusOK, ginframework.H{"code": 400, "msg": "服务异常"})
		})
		router.GET("/matched-not-found", func(ctx *ginframework.Context) {
			ctx.Status(http.StatusNotFound)
		})
		router.GET("/panic", func(_ *ginframework.Context) {
			panic("prometheus recovery test")
		})
		router.GET("/active-500", func(ctx *ginframework.Context) {
			ctx.Status(http.StatusInternalServerError)
		})
		router.GET("/ready", func(ctx *ginframework.Context) {
			ctx.JSON(http.StatusOK, ginframework.H{"status": "ok"})
		})
		if extraRoutes != nil {
			extraRoutes(router)
		}
	})
}

func performRequest(handler http.Handler, method string, target string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, target, nil))
	return recorder
}

func performWebSocketUpgrade(handler http.Handler, target string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.Header.Set("Connection", "keep-alive, Upgrade")
	request.Header.Set("Upgrade", "websocket")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func scrapeMetrics(t *testing.T, handler http.Handler) string {
	t.Helper()
	recorder := performRequest(handler, http.MethodGet, "/metrics")
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want %d; body = %q", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.Contains(contentType, "text/plain") {
		t.Fatalf("GET /metrics Content-Type = %q, want Prometheus text format", contentType)
	}
	return recorder.Body.String()
}

func assertMetricValue(t *testing.T, metrics string, metricName string, want float64, labels ...string) {
	t.Helper()
	got, ok := metricValue(t, metrics, metricName, labels...)
	if !ok {
		t.Fatalf("metric %s with labels %v not found", metricName, labels)
	}
	if got != want {
		t.Fatalf("metric %s with labels %v = %v, want %v", metricName, labels, got, want)
	}
}

func assertMetricAbsent(t *testing.T, metrics string, metricName string, labels ...string) {
	t.Helper()
	if value, ok := metricValue(t, metrics, metricName, labels...); ok {
		t.Fatalf("metric %s with labels %v = %v, want absent", metricName, labels, value)
	}
}

func metricValue(t *testing.T, metrics string, metricName string, labels ...string) (float64, bool) {
	t.Helper()
	for _, line := range strings.Split(metrics, "\n") {
		if !strings.HasPrefix(line, metricName+"{") && !strings.HasPrefix(line, metricName+" ") {
			continue
		}
		matched := true
		for _, label := range labels {
			if !strings.Contains(line, label) {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			t.Fatalf("metric line %q has no value", line)
		}
		value, err := strconv.ParseFloat(fields[len(fields)-1], 64)
		if err != nil {
			t.Fatalf("parse metric value from %q: %v", line, err)
		}
		return value, true
	}
	return 0, false
}
