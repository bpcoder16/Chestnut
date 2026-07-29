package gin

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/utils"
)

const (
	prometheusScrapeAssetPath                = "conf.example/prometheus/scrape.d/chestnut-api.yaml"
	grafanaOverviewDashboardAssetPath        = "conf.example/grafana/dashboards/chestnut-api/chestnut-api-overview-dashboard.json"
	grafanaRouteDashboardAssetPath           = "conf.example/grafana/dashboards/chestnut-api/chestnut-api-route-analysis-dashboard.json"
	grafanaInstanceRuntimeDashboardAssetPath = "conf.example/grafana/dashboards/chestnut-api/chestnut-api-instance-runtime-dashboard.json"
	prometheusDatasourceUID                  = "prometheus"
	minimumScrapeTargets                     = 3
)

type prometheusScrapeAsset struct {
	ScrapeConfigs []prometheusScrapeJob `mapstructure:"scrape_configs"`
}

type prometheusScrapeJob struct {
	JobName            string                   `mapstructure:"job_name"`
	ScrapeInterval     string                   `mapstructure:"scrape_interval"`
	ScrapeTimeout      string                   `mapstructure:"scrape_timeout"`
	Scheme             string                   `mapstructure:"scheme"`
	MetricsPath        string                   `mapstructure:"metrics_path"`
	ExtraScrapeMetrics bool                     `mapstructure:"extra_scrape_metrics"`
	StaticConfigs      []prometheusStaticConfig `mapstructure:"static_configs"`
}

type prometheusStaticConfig struct {
	Targets []string          `mapstructure:"targets"`
	Labels  map[string]string `mapstructure:"labels"`
}

type grafanaDashboardAsset struct {
	Title         string                 `json:"title"`
	UID           string                 `json:"uid"`
	Editable      bool                   `json:"editable"`
	LiveNow       *bool                  `json:"liveNow"`
	Preload       *bool                  `json:"preload"`
	Refresh       string                 `json:"refresh"`
	SchemaVersion int                    `json:"schemaVersion"`
	Style         string                 `json:"style"`
	Tags          []string               `json:"tags"`
	Timezone      string                 `json:"timezone"`
	WeekStart     string                 `json:"weekStart"`
	Links         []grafanaDashboardLink `json:"links"`
	Time          struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"time"`
	Timepicker struct {
		RefreshIntervals []string `json:"refresh_intervals"`
	} `json:"timepicker"`
	Templating struct {
		List []grafanaVariable `json:"list"`
	} `json:"templating"`
	Panels []grafanaPanel `json:"panels"`
}

type grafanaDashboardLink struct {
	AsDropdown  bool     `json:"asDropdown"`
	IncludeVars bool     `json:"includeVars"`
	KeepTime    bool     `json:"keepTime"`
	Tags        []string `json:"tags"`
	Type        string   `json:"type"`
}

type grafanaVariable struct {
	Name       string            `json:"name"`
	Definition string            `json:"definition"`
	AllValue   string            `json:"allValue"`
	IncludeAll bool              `json:"includeAll"`
	Multi      bool              `json:"multi"`
	Datasource grafanaDatasource `json:"datasource"`
	Query      struct {
		Query string `json:"query"`
	} `json:"query"`
}

type grafanaDatasource struct {
	Type string `json:"type"`
	UID  string `json:"uid"`
}

type grafanaPanel struct {
	ID          int               `json:"id"`
	Title       string            `json:"title"`
	Type        string            `json:"type"`
	Datasource  grafanaDatasource `json:"datasource"`
	Description string            `json:"description"`
	GridPos     struct {
		Height int `json:"h"`
		Width  int `json:"w"`
		X      int `json:"x"`
		Y      int `json:"y"`
	} `json:"gridPos"`
	FieldConfig struct {
		Defaults struct {
			Color struct {
				FixedColor string `json:"fixedColor"`
				Mode       string `json:"mode"`
			} `json:"color"`
			Decimals   *float64 `json:"decimals"`
			NoValue    string   `json:"noValue"`
			Unit       string   `json:"unit"`
			Min        *float64 `json:"min"`
			Thresholds struct {
				Mode  string `json:"mode"`
				Steps []struct {
					Color string   `json:"color"`
					Value *float64 `json:"value"`
				} `json:"steps"`
			} `json:"thresholds"`
			Mappings []grafanaValueMapping `json:"mappings"`
		} `json:"defaults"`
	} `json:"fieldConfig"`
	Options struct {
		GraphMode string `json:"graphMode"`
	} `json:"options"`
	Targets []struct {
		Expr         string `json:"expr"`
		Format       string `json:"format"`
		Instant      bool   `json:"instant"`
		LegendFormat string `json:"legendFormat"`
	} `json:"targets"`
	Transformations []struct {
		ID      string `json:"id"`
		Options struct {
			ExcludeByName map[string]bool   `json:"excludeByName"`
			IndexByName   map[string]int    `json:"indexByName"`
			RenameByName  map[string]string `json:"renameByName"`
		} `json:"options"`
	} `json:"transformations"`
}

type grafanaValueMapping struct {
	Type    string `json:"type"`
	Options map[string]struct {
		Color string `json:"color"`
		Text  string `json:"text"`
	} `json:"options"`
}

func TestPrometheusScrapeAssetSupportsThreeNodeCluster(t *testing.T) {
	var asset prometheusScrapeAsset
	if err := utils.ParseFile(prometheusScrapeAssetPath, &asset); err != nil {
		t.Fatalf("parse Prometheus scrape asset: %v", err)
	}
	if len(asset.ScrapeConfigs) != 1 {
		t.Fatalf("len(scrape_configs) = %d, want 1", len(asset.ScrapeConfigs))
	}

	job := asset.ScrapeConfigs[0]
	if job.JobName != "chestnut-api" {
		t.Fatalf("job_name = %q, want chestnut-api", job.JobName)
	}
	if job.Scheme != "http" || job.MetricsPath != prometheusMetricsPath || !job.ExtraScrapeMetrics {
		t.Fatalf("scrape endpoint = (%q, %q, extra=%v), want (http, /metrics, true)", job.Scheme, job.MetricsPath, job.ExtraScrapeMetrics)
	}
	interval, err := time.ParseDuration(job.ScrapeInterval)
	if err != nil {
		t.Fatalf("parse scrape_interval %q: %v", job.ScrapeInterval, err)
	}
	timeout, err := time.ParseDuration(job.ScrapeTimeout)
	if err != nil {
		t.Fatalf("parse scrape_timeout %q: %v", job.ScrapeTimeout, err)
	}
	if timeout > interval {
		t.Fatalf("scrape_timeout %s exceeds scrape_interval %s", timeout, interval)
	}

	targetCount := 0
	nodes := make(map[string]struct{}, minimumScrapeTargets)
	targetNodes := make(map[string]string, minimumScrapeTargets)
	for _, staticConfig := range job.StaticConfigs {
		targetCount += len(staticConfig.Targets)
		node := strings.TrimSpace(staticConfig.Labels["node"])
		if node == "" {
			t.Fatalf("static config %v has empty node label", staticConfig.Targets)
		}
		if len(staticConfig.Targets) != 1 {
			t.Fatalf("node %q targets = %v, want exactly one target", node, staticConfig.Targets)
		}
		if _, exists := nodes[node]; exists {
			t.Fatalf("node %q appears in multiple static configs", node)
		}
		if _, exists := staticConfig.Labels["service"]; exists {
			t.Fatalf("static config %v must not define conflicting service target label", staticConfig.Targets)
		}
		nodes[node] = struct{}{}
		for _, target := range staticConfig.Targets {
			if previousNode, exists := targetNodes[target]; exists {
				t.Fatalf("target %q is assigned to both node %q and %q", target, previousNode, node)
			}
			targetNodes[target] = node
			if !strings.HasPrefix(target, "192.0.2.") {
				t.Fatalf("target %q must use RFC 5737 documentation address space", target)
			}
		}
	}
	if targetCount < minimumScrapeTargets || len(nodes) < minimumScrapeTargets {
		t.Fatalf("targets = %d, unique nodes = %d, want at least %d each", targetCount, len(nodes), minimumScrapeTargets)
	}
}

func TestGrafanaDashboardAssetsCoverAPIClusterOperations(t *testing.T) {
	wantVariables := []struct {
		name  string
		query string
	}{
		{name: "node", query: `label_values(up{job="chestnut-api"}, node)`},
	}

	dashboardAssets := []struct {
		path        string
		uid         string
		title       string
		refresh     string
		panelTitles []string
	}{
		{
			path:    grafanaOverviewDashboardAssetPath,
			uid:     "chestnut-api-overview",
			title:   "Chestnut API 监控概览",
			refresh: "30s",
			panelTitles: []string{
				"可用节点数", "不可用节点数", "节点可用率", "节点采集明细", "最近启动节点运行时长",
				"当前 HTTP 请求数", "当前 WebSocket 连接数",
				"集群总 QPS", "HTTP 状态类别趋势", "所选时段 HTTP 5xx 比例", "所选时段 Recovery Panic 次数", "Recovery Panic 趋势",
				"HTTP 请求延迟分位数", "HTTP 平均响应耗时",
			},
		},
		{
			path:    grafanaRouteDashboardAssetPath,
			uid:     "chestnut-api-route-analysis",
			title:   "Chestnut API 路由分析",
			refresh: "1m",
			panelTitles: []string{
				"P95 最慢路由 Top 10", "各路由 P95", "路由 QPS Top 10", "各路由 5xx 比例", "请求明细",
			},
		},
		{
			path:    grafanaInstanceRuntimeDashboardAssetPath,
			uid:     "chestnut-api-instance-runtime",
			title:   "Chestnut API 节点与运行时",
			refresh: "1m",
			panelTitles: []string{
				"各节点 QPS", "各节点 P95", "各节点 5xx", "各节点并发请求数", "进程 CPU", "进程 RSS",
				"Go Heap", "Goroutine", "GC 次数", "GC Pause",
			},
		},
	}

	panels := make(map[string]grafanaPanel)
	allExpressions := make([]string, 0)
	var allRaw strings.Builder
	for _, asset := range dashboardAssets {
		raw, err := os.ReadFile(asset.path)
		if err != nil {
			t.Fatalf("os.ReadFile(%q): %v", asset.path, err)
		}
		allRaw.Write(raw)

		var dashboard grafanaDashboardAsset
		if err := json.Unmarshal(raw, &dashboard); err != nil {
			t.Fatalf("json.Unmarshal dashboard %q: %v", asset.path, err)
		}
		if dashboard.UID != asset.uid || dashboard.Title != asset.title || dashboard.Editable {
			t.Fatalf("dashboard %q identity = (uid=%q, title=%q, editable=%v), want (%q, %q, false)",
				asset.path, dashboard.UID, dashboard.Title, dashboard.Editable, asset.uid, asset.title)
		}
		if dashboard.Refresh != asset.refresh || dashboard.Time.From != "now-1h" || dashboard.Time.To != "now" {
			t.Fatalf("dashboard %q time = (refresh=%q, from=%q, to=%q), want (%q, now-1h, now)",
				asset.path, dashboard.Refresh, dashboard.Time.From, dashboard.Time.To, asset.refresh)
		}
		if dashboard.SchemaVersion != 42 {
			t.Fatalf("dashboard %q schemaVersion = %d, want 42", asset.path, dashboard.SchemaVersion)
		}
		if dashboard.LiveNow == nil || *dashboard.LiveNow || dashboard.Preload == nil || *dashboard.Preload {
			t.Fatalf("dashboard %q must explicitly disable liveNow and preload", asset.path)
		}
		if dashboard.Style != "light" || dashboard.Timezone != "Asia/Shanghai" || dashboard.WeekStart != "monday" {
			t.Fatalf("dashboard %q display settings = (style=%q, timezone=%q, weekStart=%q), want (light, Asia/Shanghai, monday)",
				asset.path, dashboard.Style, dashboard.Timezone, dashboard.WeekStart)
		}
		if strings.Join(dashboard.Tags, ",") != "chestnut,api" {
			t.Fatalf("dashboard %q tags = %v, want [chestnut api]", asset.path, dashboard.Tags)
		}
		if strings.Join(dashboard.Timepicker.RefreshIntervals, ",") != "30s,1m,5m,15m,30m,1h,2h,1d" {
			t.Fatalf("dashboard %q refresh intervals = %v, want standard dashboard intervals",
				asset.path, dashboard.Timepicker.RefreshIntervals)
		}
		if len(dashboard.Links) != 1 || dashboard.Links[0].Type != "dashboards" ||
			!dashboard.Links[0].AsDropdown || !dashboard.Links[0].IncludeVars || !dashboard.Links[0].KeepTime ||
			len(dashboard.Links[0].Tags) != 1 || dashboard.Links[0].Tags[0] != "chestnut" {
			t.Fatalf("dashboard %q links = %#v, want one chestnut dashboard dropdown preserving variables and time", asset.path, dashboard.Links)
		}
		if len(dashboard.Templating.List) != len(wantVariables) {
			t.Fatalf("dashboard %q variables = %d, want %d", asset.path, len(dashboard.Templating.List), len(wantVariables))
		}
		for index, variable := range dashboard.Templating.List {
			if variable.Datasource.UID != prometheusDatasourceUID {
				t.Fatalf("dashboard %q variable %q datasource UID = %q, want %q",
					asset.path, variable.Name, variable.Datasource.UID, prometheusDatasourceUID)
			}
			want := wantVariables[index]
			if variable.Name != want.name || variable.Query.Query != want.query || variable.Definition != want.query {
				t.Fatalf("dashboard %q variable[%d] = (name=%q, query=%q, definition=%q), want (%q, %q, %q)",
					asset.path, index, variable.Name, variable.Query.Query, variable.Definition, want.name, want.query, want.query)
			}
			if !variable.IncludeAll || !variable.Multi {
				t.Fatalf("dashboard %q variable %q must support All and multi-select", asset.path, variable.Name)
			}
			if variable.AllValue != ".*" {
				t.Fatalf("dashboard %q variable %q allValue = %q, want %q",
					asset.path, variable.Name, variable.AllValue, ".*")
			}
		}

		dashboardPanels := make(map[string]grafanaPanel, len(asset.panelTitles))
		panelIDs := make(map[int]string, len(dashboard.Panels))
		for _, panel := range dashboard.Panels {
			if previousTitle, exists := panelIDs[panel.ID]; exists {
				t.Fatalf("dashboard %q panels %q and %q share id %d", asset.path, previousTitle, panel.Title, panel.ID)
			}
			panelIDs[panel.ID] = panel.Title
			if panel.Type == "row" {
				continue
			}
			if panel.Datasource.UID != prometheusDatasourceUID {
				t.Fatalf("dashboard %q panel %q datasource UID = %q, want %q",
					asset.path, panel.Title, panel.Datasource.UID, prometheusDatasourceUID)
			}
			if strings.TrimSpace(panel.FieldConfig.Defaults.NoValue) == "" {
				t.Fatalf("dashboard %q panel %q must define noValue semantics", asset.path, panel.Title)
			}
			if panel.Type != "table" && (panel.FieldConfig.Defaults.Min == nil || *panel.FieldConfig.Defaults.Min != 0) {
				t.Fatalf("dashboard %q panel %q must pin non-negative metrics to min=0", asset.path, panel.Title)
			}
			if previous, exists := panels[panel.Title]; exists {
				t.Fatalf("panel %q appears in multiple dashboards: %#v and %q", panel.Title, previous, asset.path)
			}
			dashboardPanels[panel.Title] = panel
			panels[panel.Title] = panel
			for _, target := range panel.Targets {
				if target.Expr != "" && !strings.Contains(target.Expr, `job="chestnut-api"`) {
					t.Fatalf("dashboard %q panel %q expression must be scoped to chestnut-api: %q",
						asset.path, panel.Title, target.Expr)
				}
				allExpressions = append(allExpressions, target.Expr)
			}
		}
		if len(dashboardPanels) != len(asset.panelTitles) {
			t.Fatalf("dashboard %q data panels = %d, want %d", asset.path, len(dashboardPanels), len(asset.panelTitles))
		}
		for _, panelTitle := range asset.panelTitles {
			if _, exists := dashboardPanels[panelTitle]; !exists {
				t.Fatalf("dashboard %q panel %q not found", asset.path, panelTitle)
			}
		}
	}
	assertTargetHealthPanels(t, panels)
	assertOverviewSummaryPresentation(t, panels)
	assertDashboardDiagnosticSemantics(t, panels)

	expressions := strings.Join(allExpressions, "\n")
	for _, metricName := range []string{
		"chestnut_http_server_requests_total",
		"chestnut_http_server_request_duration_seconds_bucket",
		"chestnut_http_server_requests_in_flight",
		"chestnut_http_server_websocket_connections",
		"chestnut_http_server_recovered_panics_total",
		"process_cpu_seconds_total",
		"process_resident_memory_bytes",
		"go_goroutines",
		"go_memstats_heap_alloc_bytes",
	} {
		if !strings.Contains(expressions, metricName) {
			t.Fatalf("dashboard expressions do not contain %q", metricName)
		}
	}
	if !strings.Contains(expressions, "$__rate_interval") {
		t.Fatal("dashboard rate expressions must use $__rate_interval")
	}
	if !strings.Contains(expressions, "histogram_quantile(") || !strings.Contains(expressions, "sum by (le") {
		t.Fatal("dashboard latency expressions must aggregate buckets by le before histogram_quantile")
	}
	if strings.Contains(expressions, "avg(histogram_quantile") {
		t.Fatal("dashboard must not average node quantiles")
	}
	for _, forbidden := range []string{
		"de_card_http_", "Response.Code", "nats_",
		`"$instance"`, `"$service"`, `"$job"`, `"$route"`,
		`status=~`, `by (status)`, `{{status}}`,
	} {
		if strings.Contains(allRaw.String(), forbidden) {
			t.Fatalf("dashboards contain forbidden legacy or unrelated value %q", forbidden)
		}
	}

	for _, panelTitle := range []string{"各节点 QPS", "各节点 P95", "各节点 5xx", "各节点并发请求数", "进程 CPU", "进程 RSS", "Go Heap", "Goroutine", "GC 次数", "GC Pause"} {
		panel := panels[panelTitle]
		for _, target := range panel.Targets {
			if !strings.Contains(target.Expr, "node") || target.LegendFormat != "{{node}}" {
				t.Fatalf("panel %q must use node as its only operational identity", panelTitle)
			}
		}
	}
}

func assertDashboardDiagnosticSemantics(t *testing.T, panels map[string]grafanaPanel) {
	t.Helper()

	statusTrend := panels["HTTP 状态类别趋势"]
	wantStatusClasses := []string{"2xx", "3xx", "4xx", "5xx"}
	if len(statusTrend.Targets) != len(wantStatusClasses) {
		t.Fatalf("HTTP 状态类别趋势 targets = %d, want %d", len(statusTrend.Targets), len(wantStatusClasses))
	}
	for index, statusClass := range wantStatusClasses {
		target := statusTrend.Targets[index]
		if !strings.Contains(target.Expr, `status_class="`+statusClass+`"`) || target.LegendFormat != statusClass {
			t.Fatalf("HTTP 状态类别趋势 target[%d] = (expr=%q, legend=%q), want status_class=%q",
				index, target.Expr, target.LegendFormat, statusClass)
		}
	}

	for _, title := range []string{"P95 最慢路由 Top 10", "路由 QPS Top 10"} {
		panel := panels[title]
		if panel.Type != "table" || len(panel.Targets) != 1 || !panel.Targets[0].Instant {
			t.Fatalf("panel %q must be an instant Top-N table", title)
		}
		wantTransformations := "seriesToRows,sortBy,organize"
		gotTransformations := make([]string, 0, len(panel.Transformations))
		for _, transformation := range panel.Transformations {
			gotTransformations = append(gotTransformations, transformation.ID)
		}
		if strings.Join(gotTransformations, ",") != wantTransformations {
			t.Fatalf("panel %q transformations = %v, want %s", title, gotTransformations, wantTransformations)
		}
	}

	selectedRange5xx := panels["所选时段 HTTP 5xx 比例"]
	if len(selectedRange5xx.Targets) != 1 || !selectedRange5xx.Targets[0].Instant {
		t.Fatal("所选时段 HTTP 5xx 比例 must use one instant query")
	}
	if expression := selectedRange5xx.Targets[0].Expr; strings.Count(expression, "increase(") != 2 ||
		strings.Count(expression, "[$__range]") != 2 || strings.Contains(expression, "rate(") ||
		!strings.Contains(expression, "or vector(0)") || !strings.Contains(expression, "> 0)") {
		t.Fatalf("所选时段 HTTP 5xx 比例 must cover the selected range and distinguish zero errors from zero traffic: %q", expression)
	}
	for _, title := range []string{"各路由 5xx 比例", "各节点 5xx"} {
		if expression := panels[title].Targets[0].Expr; !strings.Contains(expression, "or 0 *") {
			t.Fatalf("panel %q must retain healthy zero-error series: %q", title, expression)
		}
	}
	selectedRangePanics := panels["所选时段 Recovery Panic 次数"]
	if len(selectedRangePanics.Targets) != 1 || !selectedRangePanics.Targets[0].Instant {
		t.Fatal("所选时段 Recovery Panic 次数 must use one instant query")
	}
	if expression := selectedRangePanics.Targets[0].Expr; !strings.Contains(expression, "increase(") ||
		strings.Count(expression, "[$__range]") != 1 || strings.Contains(expression, "[5m]") ||
		!strings.Contains(expression, "or vector(0)") || !strings.Contains(expression, "count(up{") {
		t.Fatalf("所选时段 Recovery Panic 次数 must cover the selected range and preserve missing-target semantics: %q", expression)
	}
	if expression := panels["Recovery Panic 趋势"].Targets[0].Expr; !strings.Contains(expression, "or 0 * max by (node) (up{") ||
		!strings.Contains(expression, "== 1)") {
		t.Fatalf("Recovery Panic 趋势 must fill zero only for healthy nodes: %q", expression)
	}
	if panels["GC Pause"].FieldConfig.Defaults.Unit != "percentunit" {
		t.Fatalf("GC Pause unit = %q, want percentunit for seconds-per-second ratio",
			panels["GC Pause"].FieldConfig.Defaults.Unit)
	}
}

func assertOverviewSummaryPresentation(t *testing.T, panels map[string]grafanaPanel) {
	t.Helper()

	available := panels["可用节点数"]
	if available.FieldConfig.Defaults.Color.Mode != "fixed" ||
		available.FieldConfig.Defaults.Color.FixedColor != "semi-dark-blue" {
		t.Fatalf("可用节点数 must use a neutral fixed color")
	}

	availability := panels["节点可用率"]
	steps := availability.FieldConfig.Defaults.Thresholds.Steps
	if availability.FieldConfig.Defaults.Decimals == nil || *availability.FieldConfig.Defaults.Decimals != 0 ||
		len(steps) != 2 || steps[0].Color != "red" || steps[0].Value != nil ||
		steps[1].Color != "green" || steps[1].Value == nil || *steps[1].Value != 100 {
		t.Fatalf("节点可用率 must display whole percentages and turn green only at 100%%")
	}

	for index, title := range []string{"可用节点数", "不可用节点数", "节点可用率", "最近启动节点运行时长"} {
		panel := panels[title]
		if panel.GridPos.Width != 6 || panel.GridPos.X != index*6 || panel.GridPos.Y != 1 {
			t.Fatalf("summary panel %q must occupy one quarter of the health row", title)
		}
	}

	httpInFlight := panels["当前 HTTP 请求数"]
	if httpInFlight.GridPos.Width != 12 || httpInFlight.GridPos.X != 0 || httpInFlight.GridPos.Y != 20 ||
		httpInFlight.Options.GraphMode != "area" || len(httpInFlight.Targets) != 1 || httpInFlight.Targets[0].Instant ||
		!strings.Contains(httpInFlight.Targets[0].Expr, "chestnut_http_server_requests_in_flight") ||
		!strings.Contains(httpInFlight.Targets[0].Expr, "chestnut_http_server_websocket_connections") {
		t.Fatalf("当前 HTTP 请求数 must show the selected-range non-WebSocket in-flight sparkline")
	}

	webSocketConnections := panels["当前 WebSocket 连接数"]
	if webSocketConnections.GridPos.Width != 12 || webSocketConnections.GridPos.X != 12 ||
		webSocketConnections.GridPos.Y != 20 || webSocketConnections.Options.GraphMode != "area" ||
		len(webSocketConnections.Targets) != 1 || webSocketConnections.Targets[0].Instant ||
		!strings.Contains(webSocketConnections.Targets[0].Expr, "chestnut_http_server_websocket_connections") {
		t.Fatalf("当前 WebSocket 连接数 must share the selected-range live-load row")
	}

	uptime := panels["最近启动节点运行时长"]
	uptimeSteps := uptime.FieldConfig.Defaults.Thresholds.Steps
	if uptime.GridPos.Width != 6 || uptime.GridPos.X != 18 ||
		uptime.FieldConfig.Defaults.Unit != "suffix: 小时" ||
		uptime.FieldConfig.Defaults.Decimals == nil || *uptime.FieldConfig.Defaults.Decimals != 2 ||
		len(uptime.Targets) != 1 || !strings.HasSuffix(uptime.Targets[0].Expr, "/ 3600") ||
		len(uptimeSteps) != 3 || uptimeSteps[0].Color != "red" || uptimeSteps[0].Value != nil ||
		uptimeSteps[1].Color != "yellow" || uptimeSteps[1].Value == nil || *uptimeSteps[1].Value != 0.25 ||
		uptimeSteps[2].Color != "green" || uptimeSteps[2].Value == nil || *uptimeSteps[2].Value != 1 {
		t.Fatalf("最近启动节点运行时长 must use Chinese hours and equivalent 15-minute/1-hour thresholds")
	}
}

func TestGrafanaDashboardTargetHealthSemantics(t *testing.T) {
	tests := []struct {
		name             string
		upValues         []float64
		wantAvailable    int
		wantUnavailable  int
		wantAvailability float64
		wantHasTargets   bool
	}{
		{
			name:             "all targets up",
			upValues:         []float64{1, 1, 1},
			wantAvailable:    3,
			wantUnavailable:  0,
			wantAvailability: 100,
			wantHasTargets:   true,
		},
		{
			name:             "one target down",
			upValues:         []float64{1, 1, 0},
			wantAvailable:    2,
			wantUnavailable:  1,
			wantAvailability: 200.0 / 3.0,
			wantHasTargets:   true,
		},
		{
			name:             "all targets down",
			upValues:         []float64{0, 0, 0},
			wantAvailable:    0,
			wantUnavailable:  3,
			wantAvailability: 0,
			wantHasTargets:   true,
		},
		{
			name:           "no target series",
			wantHasTargets: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotAvailable, gotUnavailable, gotAvailability, gotHasTargets := summarizeTargetHealth(test.upValues)
			if gotAvailable != test.wantAvailable || gotUnavailable != test.wantUnavailable ||
				gotAvailability != test.wantAvailability || gotHasTargets != test.wantHasTargets {
				t.Fatalf("summarizeTargetHealth(%v) = (%d, %d, %v, %v), want (%d, %d, %v, %v)",
					test.upValues, gotAvailable, gotUnavailable, gotAvailability, gotHasTargets,
					test.wantAvailable, test.wantUnavailable, test.wantAvailability, test.wantHasTargets)
			}
		})
	}

	raw, err := os.ReadFile(grafanaOverviewDashboardAssetPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", grafanaOverviewDashboardAssetPath, err)
	}
	var dashboard grafanaDashboardAsset
	if err := json.Unmarshal(raw, &dashboard); err != nil {
		t.Fatalf("json.Unmarshal dashboard: %v", err)
	}
	panels := make(map[string]grafanaPanel, len(dashboard.Panels))
	for _, panel := range dashboard.Panels {
		panels[panel.Title] = panel
	}
	assertTargetHealthPanels(t, panels)
}

func assertTargetHealthPanels(t *testing.T, panels map[string]grafanaPanel) {
	t.Helper()

	const targetFilters = `job="chestnut-api",node=~"$node"`
	wantExpressions := map[string]string{
		"可用节点数":  "sum(up{" + targetFilters + "})",
		"不可用节点数": "count(up{" + targetFilters + "}) - sum(up{" + targetFilters + "})",
		"节点可用率":  "100 * sum(up{" + targetFilters + "}) / count(up{" + targetFilters + "})",
		"节点采集明细": "max by (job, node, instance) (up{" + targetFilters + "})",
	}
	for title, wantExpression := range wantExpressions {
		panel, exists := panels[title]
		if !exists {
			t.Fatalf("dashboard panel %q not found", title)
		}
		if panel.FieldConfig.Defaults.NoValue != "无节点采集数据" {
			t.Fatalf("panel %q noValue = %q, want 无节点采集数据", title, panel.FieldConfig.Defaults.NoValue)
		}
		if len(panel.Targets) != 1 || panel.Targets[0].Expr != wantExpression {
			t.Fatalf("panel %q expression = %v, want %q", title, panel.Targets, wantExpression)
		}
		if strings.Contains(panel.Targets[0].Expr, "$service") || strings.Contains(panel.Targets[0].Expr, "$route") {
			t.Fatalf("target health panel %q must not use service or route filters", title)
		}
	}

	detailPanel := panels["节点采集明细"]
	if detailPanel.Type != "table" || !detailPanel.Targets[0].Instant || detailPanel.Targets[0].Format != "table" {
		t.Fatalf("节点采集明细 = (type=%q, instant=%v, format=%q), want table/true/table",
			detailPanel.Type, detailPanel.Targets[0].Instant, detailPanel.Targets[0].Format)
	}
	if len(detailPanel.FieldConfig.Defaults.Mappings) != 1 || detailPanel.FieldConfig.Defaults.Mappings[0].Type != "value" {
		t.Fatalf("节点采集明细 mappings = %v, want one value mapping", detailPanel.FieldConfig.Defaults.Mappings)
	}
	mappingOptions := detailPanel.FieldConfig.Defaults.Mappings[0].Options
	if mappingOptions["1"].Text != "正常" || mappingOptions["1"].Color != "green" ||
		mappingOptions["0"].Text != "异常" || mappingOptions["0"].Color != "red" {
		t.Fatalf("节点采集明细 mapping options = %v, want 1=green 正常 and 0=red 异常", mappingOptions)
	}
	if detailPanel.GridPos.Height != 6 || detailPanel.GridPos.Y != 5 ||
		panels["集群总 QPS"].GridPos.Y != 12 || panels["当前 HTTP 请求数"].GridPos.Y != 20 ||
		panels["所选时段 Recovery Panic 次数"].GridPos.Y != 27 ||
		panels["HTTP 请求延迟分位数"].GridPos.Y != 35 {
		t.Fatal("节点采集明细及后续区域 must use the compact vertical layout")
	}
	if len(detailPanel.Transformations) != 3 ||
		detailPanel.Transformations[0].ID != "labelsToFields" ||
		detailPanel.Transformations[1].ID != "sortBy" ||
		detailPanel.Transformations[2].ID != "organize" {
		t.Fatalf("节点采集明细 transformations = %v, want labelsToFields/sortBy/organize",
			detailPanel.Transformations)
	}
	organize := detailPanel.Transformations[2].Options
	if len(organize.ExcludeByName) != 2 || !organize.ExcludeByName["job"] || !organize.ExcludeByName["instance"] ||
		len(organize.IndexByName) != 3 || organize.IndexByName["Time"] != 0 ||
		organize.IndexByName["node"] != 1 || organize.IndexByName["Value"] != 2 ||
		organize.RenameByName["Time"] != "采集时间" || organize.RenameByName["node"] != "节点" ||
		organize.RenameByName["Value"] != "采集状态" {
		t.Fatalf("节点采集明细 organize options = %#v, want only 采集时间/节点/采集状态", organize)
	}
}

func summarizeTargetHealth(upValues []float64) (available int, unavailable int, availability float64, hasTargets bool) {
	if len(upValues) == 0 {
		return 0, 0, 0, false
	}
	for _, value := range upValues {
		if value == 1 {
			available++
		}
	}
	unavailable = len(upValues) - available
	availability = 100 * float64(available) / float64(len(upValues))
	return available, unavailable, availability, true
}
