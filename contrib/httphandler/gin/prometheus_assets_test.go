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
	grafanaRouteDetailDashboardAssetPath     = "conf.example/grafana/dashboards/chestnut-api/chestnut-api-route-detail-dashboard.json"
	grafanaInstanceRuntimeDashboardAssetPath = "conf.example/grafana/dashboards/chestnut-api/chestnut-api-instance-runtime-dashboard.json"
	prometheusDatasourceUID                  = "prometheus"
	grafanaRoutePlaceholderValue             = "请选择 Route"
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
	Name        string                  `json:"name"`
	Label       string                  `json:"label"`
	Type        string                  `json:"type"`
	Definition  string                  `json:"definition"`
	AllValue    string                  `json:"allValue"`
	IncludeAll  bool                    `json:"includeAll"`
	Multi       bool                    `json:"multi"`
	Current     grafanaVariableOption   `json:"current"`
	Options     []grafanaVariableOption `json:"options"`
	Regex       string                  `json:"regex"`
	SkipURLSync bool                    `json:"skipUrlSync"`
	Datasource  grafanaDatasource       `json:"datasource"`
	Query       struct {
		Query string `json:"query"`
	} `json:"query"`
}

type grafanaVariableOption struct {
	Selected bool   `json:"selected"`
	Text     string `json:"text"`
	Value    string `json:"value"`
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
			Custom struct {
				DrawStyle   string `json:"drawStyle"`
				FillOpacity int    `json:"fillOpacity"`
				LineWidth   int    `json:"lineWidth"`
				PointSize   int    `json:"pointSize"`
				ShowPoints  string `json:"showPoints"`
			} `json:"custom"`
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
		Overrides []struct {
			Matcher struct {
				Options string `json:"options"`
			} `json:"matcher"`
			Properties []struct {
				ID    string          `json:"id"`
				Value json.RawMessage `json:"value"`
			} `json:"properties"`
		} `json:"overrides"`
	} `json:"fieldConfig"`
	Options struct {
		GraphMode string `json:"graphMode"`
		Color     struct {
			Mode    string `json:"mode"`
			Reverse bool   `json:"reverse"`
			Scheme  string `json:"scheme"`
			Steps   int    `json:"steps"`
		} `json:"color"`
		Legend struct {
			DisplayMode string `json:"displayMode"`
			Placement   string `json:"placement"`
			SortBy      string `json:"sortBy"`
			SortDesc    *bool  `json:"sortDesc"`
		} `json:"legend"`
		Tooltip struct {
			Mode string `json:"mode"`
		} `json:"tooltip"`
	} `json:"options"`
	Targets []struct {
		Expr         string `json:"expr"`
		Format       string `json:"format"`
		Instant      bool   `json:"instant"`
		LegendFormat string `json:"legendFormat"`
		RefID        string `json:"refId"`
	} `json:"targets"`
	Transformations []struct {
		ID      string `json:"id"`
		Options struct {
			ByField       string            `json:"byField"`
			ExcludeByName map[string]bool   `json:"excludeByName"`
			IndexByName   map[string]int    `json:"indexByName"`
			Mode          string            `json:"mode"`
			RenameByName  map[string]string `json:"renameByName"`
			Sort          []struct {
				Descending bool   `json:"desc"`
				Field      string `json:"field"`
			} `json:"sort"`
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

func grafanaFieldOverrideNumber(t *testing.T, panel grafanaPanel, fieldName string, propertyID string) (float64, bool) {
	t.Helper()
	for _, override := range panel.FieldConfig.Overrides {
		if override.Matcher.Options != fieldName {
			continue
		}
		for _, property := range override.Properties {
			if property.ID != propertyID {
				continue
			}
			var value float64
			if err := json.Unmarshal(property.Value, &value); err != nil {
				t.Fatalf("decode field %q property %q: %v", fieldName, propertyID, err)
			}
			return value, true
		}
	}
	return 0, false
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
				"HTTP 并发请求趋势", "WebSocket 连接数趋势",
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
				"所选时段 P95 最慢路由 Top 10", "所选时段请求数量最高的路由 Top 10", "整体请求延迟分布热力图",
				"各路由 5xx 比例最高的 Top10", "各路由 Recovery Panic 次数最高的 Top10",
			},
		},
		{
			path:    grafanaInstanceRuntimeDashboardAssetPath,
			uid:     "chestnut-api-instance-runtime",
			title:   "Chestnut API 节点与运行时",
			refresh: "1m",
			panelTitles: []string{
				"各节点 QPS", "各节点 P95", "各节点 5xx QPS", "各节点 Recovery Panic 趋势",
				"各节点 HTTP 并发请求趋势", "各节点 WebSocket 连接数趋势",
				"进程 CPU", "进程 CPU 占整机比例", "进程常驻内存",
				"Go 堆当前已分配内存", "Go 内存分配速率", "各节点 Go 协程数量趋势", "各节点 GC 频率趋势", "各节点 GC 暂停时间占比趋势",
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
		"go_sched_gomaxprocs_threads",
		"process_resident_memory_bytes",
		"go_goroutines",
		"go_gc_duration_seconds_count",
		"go_gc_duration_seconds_sum",
		"go_memstats_alloc_bytes_total",
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

	for _, panelTitle := range []string{
		"各节点 QPS", "各节点 P95", "各节点 5xx QPS", "各节点 Recovery Panic 趋势",
		"各节点 HTTP 并发请求趋势", "各节点 WebSocket 连接数趋势",
		"进程 CPU", "进程 CPU 占整机比例", "进程常驻内存", "Go 堆当前已分配内存", "Go 内存分配速率",
		"各节点 Go 协程数量趋势", "各节点 GC 频率趋势", "各节点 GC 暂停时间占比趋势",
	} {
		panel := panels[panelTitle]
		for _, target := range panel.Targets {
			if !strings.Contains(target.Expr, "node") || target.LegendFormat != "{{node}}" {
				t.Fatalf("panel %q must use node as its only operational identity", panelTitle)
			}
		}
		if panel.Options.Legend.SortBy != "Name" ||
			panel.Options.Legend.SortDesc == nil || *panel.Options.Legend.SortDesc {
			t.Fatalf("panel %q legend must sort by Name ascending", panelTitle)
		}
		if panel.Options.Legend.DisplayMode != "list" || panel.Options.Legend.Placement != "bottom" {
			t.Fatalf("panel %q legend must use a horizontal list below the chart", panelTitle)
		}
	}
	for _, panelTitle := range []string{
		"进程 CPU", "进程常驻内存", "Go 堆当前已分配内存", "Go 内存分配速率",
		"各节点 Go 协程数量趋势", "各节点 GC 频率趋势", "各节点 GC 暂停时间占比趋势",
	} {
		if expression := panels[panelTitle].Targets[0].Expr; !strings.Contains(expression, "sum by (node) (") {
			t.Fatalf("panel %q must aggregate by node so list legend order follows the node label: %q",
				panelTitle, expression)
		}
	}
}

func TestGrafanaRouteDetailDashboardRequiresSingleRoute(t *testing.T) {
	raw, err := os.ReadFile(grafanaRouteDetailDashboardAssetPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", grafanaRouteDetailDashboardAssetPath, err)
	}

	var dashboard grafanaDashboardAsset
	if err := json.Unmarshal(raw, &dashboard); err != nil {
		t.Fatalf("json.Unmarshal dashboard: %v", err)
	}
	if dashboard.UID != "chestnut-api-route-detail" ||
		dashboard.Title != "Chestnut API 路由详情" ||
		dashboard.Editable {
		t.Fatalf("route detail dashboard identity = (uid=%q, title=%q, editable=%v)",
			dashboard.UID, dashboard.Title, dashboard.Editable)
	}
	if dashboard.Refresh != "30s" || dashboard.Time.From != "now-1h" || dashboard.Time.To != "now" ||
		dashboard.SchemaVersion != 42 || dashboard.Timezone != "Asia/Shanghai" {
		t.Fatalf("route detail dashboard has inconsistent time or schema settings")
	}

	if len(dashboard.Templating.List) != 1 {
		t.Fatalf("route detail dashboard variables = %d, want only Route", len(dashboard.Templating.List))
	}
	routeVariable := dashboard.Templating.List[0]
	const routeVariableQuery = `query_result(count by (route) (chestnut_http_server_recovered_panics_total{job="chestnut-api"}) or label_replace(vector(0), "route", "请选择 Route", "", ""))`
	if routeVariable.Name != "route" || routeVariable.Label != "Route" || routeVariable.Type != "query" ||
		routeVariable.Datasource.UID != prometheusDatasourceUID ||
		routeVariable.Definition != routeVariableQuery || routeVariable.Query.Query != routeVariableQuery {
		t.Fatalf("route detail dashboard variable is invalid: %#v", routeVariable)
	}
	if routeVariable.Multi || routeVariable.IncludeAll || routeVariable.AllValue != "" {
		t.Fatal("Route variable must be single-select without an All option")
	}
	if routeVariable.SkipURLSync {
		t.Fatal("Route variable must accept Grafana var-route URL parameters")
	}
	if routeVariable.Regex != `/.*route="([^"]+)".*/` {
		t.Fatalf("Route variable regex = %q, want route label extraction", routeVariable.Regex)
	}
	if !routeVariable.Current.Selected ||
		routeVariable.Current.Text != grafanaRoutePlaceholderValue ||
		routeVariable.Current.Value != grafanaRoutePlaceholderValue ||
		len(routeVariable.Options) != 1 ||
		!routeVariable.Options[0].Selected ||
		routeVariable.Options[0].Value != grafanaRoutePlaceholderValue {
		t.Fatalf("Route variable must default to the non-matching placeholder %q", grafanaRoutePlaceholderValue)
	}

	type gridPosition struct {
		x      int
		y      int
		width  int
		height int
	}
	wantPanels := map[string]gridPosition{
		"集群总 QPS":           {x: 0, y: 1, width: 12, height: 8},
		"HTTP 状态类别趋势":       {x: 12, y: 1, width: 12, height: 8},
		"HTTP 请求延迟分位数":      {x: 0, y: 10, width: 12, height: 9},
		"HTTP 平均响应耗时":       {x: 12, y: 10, width: 12, height: 9},
		"Recovery Panic 趋势": {x: 0, y: 20, width: 24, height: 8},
	}
	wantRows := map[string]int{
		"Route 流量": 0,
		"Route 延迟": 9,
		"Route 异常": 19,
	}
	panels := make(map[string]grafanaPanel, len(wantPanels))
	seenRows := make(map[string]struct{}, len(wantRows))
	for _, panel := range dashboard.Panels {
		if panel.Type == "row" {
			wantY, exists := wantRows[panel.Title]
			if !exists {
				t.Fatalf("route detail dashboard has unexpected row %q", panel.Title)
			}
			if _, duplicate := seenRows[panel.Title]; duplicate {
				t.Fatalf("route detail dashboard repeats row %q", panel.Title)
			}
			seenRows[panel.Title] = struct{}{}
			if panel.GridPos.X != 0 || panel.GridPos.Y != wantY ||
				panel.GridPos.Width != 24 || panel.GridPos.Height != 1 {
				t.Fatalf("route detail dashboard row layout is invalid: %#v", panel)
			}
			continue
		}

		wantGrid, exists := wantPanels[panel.Title]
		if !exists {
			t.Fatalf("route detail dashboard has unexpected panel %q", panel.Title)
		}
		if _, duplicate := panels[panel.Title]; duplicate {
			t.Fatalf("route detail dashboard repeats panel %q", panel.Title)
		}
		panels[panel.Title] = panel
		if panel.Type != "timeseries" || panel.Datasource.UID != prometheusDatasourceUID ||
			panel.GridPos.X != wantGrid.x || panel.GridPos.Y != wantGrid.y ||
			panel.GridPos.Width != wantGrid.width || panel.GridPos.Height != wantGrid.height {
			t.Fatalf("route detail dashboard panel %q layout or datasource is invalid", panel.Title)
		}
		if panel.FieldConfig.Defaults.NoValue != grafanaRoutePlaceholderValue ||
			panel.FieldConfig.Defaults.Min == nil || *panel.FieldConfig.Defaults.Min != 0 {
			t.Fatalf("route detail dashboard panel %q must expose empty-selection semantics", panel.Title)
		}
		for _, target := range panel.Targets {
			routeMatcherCount := strings.Count(target.Expr, `route="$route"`)
			jobMatcherCount := strings.Count(target.Expr, `job="chestnut-api"`)
			if routeMatcherCount == 0 || routeMatcherCount != jobMatcherCount ||
				strings.Contains(target.Expr, "$node") ||
				strings.Contains(target.Expr, `route=~`) ||
				strings.Contains(target.Expr, "vector(0)") {
				t.Fatalf("route detail dashboard panel %q target must filter every series by exactly one Route and no Node variable: %q",
					panel.Title, target.Expr)
			}
		}
	}
	if len(seenRows) != len(wantRows) || len(panels) != len(wantPanels) {
		t.Fatalf("route detail dashboard has %d rows and %d data panels, want %d and %d",
			len(seenRows), len(panels), len(wantRows), len(wantPanels))
	}

	qps := panels["集群总 QPS"]
	if len(qps.Targets) != 1 ||
		!strings.Contains(qps.Targets[0].Expr, "rate(chestnut_http_server_requests_total") ||
		!strings.Contains(qps.Targets[0].Expr, "0 * sum(chestnut_http_server_recovered_panics_total") {
		t.Fatalf("route detail QPS must retain a selected-route zero baseline: %q", qps.Targets[0].Expr)
	}
	statusTrend := panels["HTTP 状态类别趋势"]
	wantStatusClasses := []string{"2xx", "3xx", "4xx", "5xx"}
	if len(statusTrend.Targets) != len(wantStatusClasses) {
		t.Fatalf("route detail HTTP status targets = %d, want %d", len(statusTrend.Targets), len(wantStatusClasses))
	}
	for index, statusClass := range wantStatusClasses {
		target := statusTrend.Targets[index]
		if !strings.Contains(target.Expr, `status_class="`+statusClass+`"`) ||
			!strings.Contains(target.Expr, "0 * sum(chestnut_http_server_recovered_panics_total") ||
			target.LegendFormat != statusClass {
			t.Fatalf("route detail HTTP status target[%d] is invalid: %#v", index, target)
		}
	}

	latencyQuantiles := panels["HTTP 请求延迟分位数"]
	wantQuantiles := []struct {
		quantile string
		legend   string
	}{
		{quantile: "0.50", legend: "P50"},
		{quantile: "0.90", legend: "P90"},
		{quantile: "0.95", legend: "P95"},
		{quantile: "0.99", legend: "P99"},
	}
	if latencyQuantiles.FieldConfig.Defaults.Unit != "s" ||
		latencyQuantiles.FieldConfig.Defaults.Custom.ShowPoints != "never" ||
		len(latencyQuantiles.Targets) != len(wantQuantiles) {
		t.Fatal("route detail HTTP latency quantiles must display four point-free series in seconds")
	}
	for index, want := range wantQuantiles {
		target := latencyQuantiles.Targets[index]
		if !strings.Contains(target.Expr, "histogram_quantile("+want.quantile) ||
			!strings.Contains(target.Expr, "sum by (le) (rate(chestnut_http_server_request_duration_seconds_bucket") ||
			!strings.Contains(target.Expr, "and on() (sum(rate(chestnut_http_server_request_duration_seconds_count") ||
			!strings.Contains(target.Expr, "> 0)") ||
			!strings.Contains(target.Expr, "or 0 * sum(chestnut_http_server_recovered_panics_total") ||
			!strings.Contains(target.Expr, "[$__rate_interval]") ||
			target.LegendFormat != want.legend {
			t.Fatalf("route detail HTTP latency quantile target[%d] is invalid: %#v", index, target)
		}
	}

	averageLatency := panels["HTTP 平均响应耗时"]
	if averageLatency.FieldConfig.Defaults.Unit != "s" ||
		averageLatency.FieldConfig.Defaults.Custom.ShowPoints != "never" ||
		len(averageLatency.Targets) != 1 ||
		!strings.Contains(averageLatency.Targets[0].Expr, "chestnut_http_server_request_duration_seconds_sum") ||
		!strings.Contains(averageLatency.Targets[0].Expr, "chestnut_http_server_request_duration_seconds_count") ||
		!strings.Contains(averageLatency.Targets[0].Expr, "[$__rate_interval]") ||
		!strings.Contains(averageLatency.Targets[0].Expr, "> 0)") ||
		!strings.Contains(averageLatency.Targets[0].Expr, "or 0 * sum(chestnut_http_server_recovered_panics_total") {
		t.Fatalf("route detail HTTP average latency query is invalid: %#v", averageLatency.Targets)
	}

	recoveryPanics := panels["Recovery Panic 趋势"]
	if recoveryPanics.FieldConfig.Defaults.Unit != "ops" ||
		recoveryPanics.FieldConfig.Defaults.Custom.ShowPoints != "never" ||
		len(recoveryPanics.Targets) != 1 ||
		!strings.Contains(recoveryPanics.Targets[0].Expr, "sum by (node) (rate(chestnut_http_server_recovered_panics_total") ||
		!strings.Contains(recoveryPanics.Targets[0].Expr, "[$__rate_interval]") ||
		recoveryPanics.Targets[0].LegendFormat != "{{node}}" ||
		recoveryPanics.Options.Legend.DisplayMode != "list" ||
		recoveryPanics.Options.Legend.Placement != "bottom" ||
		recoveryPanics.Options.Legend.SortBy != "Name" ||
		recoveryPanics.Options.Legend.SortDesc == nil ||
		*recoveryPanics.Options.Legend.SortDesc {
		t.Fatalf("route detail Recovery Panic trend is invalid: %#v", recoveryPanics)
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

	selectedRangeP95 := panels["所选时段 P95 最慢路由 Top 10"]
	if selectedRangeP95.Type != "table" || len(selectedRangeP95.Targets) != 4 {
		t.Fatal("所选时段 P95 最慢路由 Top 10 must combine P95, sample count, slow-request count, and slow-request ratio")
	}
	for index, target := range selectedRangeP95.Targets {
		if !target.Instant || target.Format != "table" {
			t.Fatalf("所选时段 P95 最慢路由 Top 10 target[%d] must be an instant table query", index)
		}
		if strings.Count(target.Expr, `route!="/api/*path"`) != strings.Count(target.Expr, "increase(") {
			t.Fatalf("所选时段 P95 最慢路由 Top 10 target[%d] must exclude /api/*path from every range aggregation: %q",
				index, target.Expr)
		}
	}
	gotP95Transformations := make([]string, 0, len(selectedRangeP95.Transformations))
	for _, transformation := range selectedRangeP95.Transformations {
		gotP95Transformations = append(gotP95Transformations, transformation.ID)
	}
	if strings.Join(gotP95Transformations, ",") != "joinByField,sortBy,organize" ||
		selectedRangeP95.Transformations[0].Options.ByField != "route" ||
		selectedRangeP95.Transformations[0].Options.Mode != "inner" {
		t.Fatalf("所选时段 P95 最慢路由 Top 10 must inner-join table queries by route: %v", gotP95Transformations)
	}
	if sortOptions := selectedRangeP95.Transformations[1].Options.Sort; len(sortOptions) != 1 ||
		!sortOptions[0].Descending || sortOptions[0].Field != "Value #A" {
		t.Fatal("所选时段 P95 最慢路由 Top 10 must sort the joined table by P95 descending")
	}
	renamedFields := selectedRangeP95.Transformations[2].Options.RenameByName
	if renamedFields["route"] != "Route" || renamedFields["Value #A"] != "P95" ||
		renamedFields["Value #B"] != "请求数" || renamedFields["Value #C"] != ">800ms 比例" ||
		renamedFields["Value #D"] != ">800ms 请求数" {
		t.Fatalf("所选时段 P95 最慢路由 Top 10 renamed fields = %v", renamedFields)
	}
	indexedP95Fields := selectedRangeP95.Transformations[2].Options.IndexByName
	if indexedP95Fields["route"] != 0 || indexedP95Fields["Value #B"] != 1 ||
		indexedP95Fields["Value #A"] != 2 || indexedP95Fields["Value #D"] != 3 ||
		indexedP95Fields["Value #C"] != 4 {
		t.Fatalf("所选时段 P95 最慢路由 Top 10 field order = %v", indexedP95Fields)
	}
	if expression := selectedRangeP95.Targets[0].Expr; !strings.Contains(expression, "increase(") ||
		!strings.Contains(expression, "[$__range]") || strings.Contains(expression, "rate(") {
		t.Fatalf("所选时段 P95 最慢路由 Top 10 must aggregate the full selected range: %q", expression)
	}
	sampleCountTarget := selectedRangeP95.Targets[1]
	if sampleCountTarget.RefID != "B" ||
		!strings.Contains(sampleCountTarget.Expr, "sum by (route)") ||
		!strings.Contains(sampleCountTarget.Expr, "chestnut_http_server_request_duration_seconds_count") ||
		strings.Count(sampleCountTarget.Expr, "increase(") != 1 ||
		strings.Count(sampleCountTarget.Expr, "[$__range]") != 1 ||
		strings.Contains(sampleCountTarget.Expr, "requests_total") {
		t.Fatalf("所选时段 P95 最慢路由 Top 10 sample-count query is invalid: %q", sampleCountTarget.Expr)
	}
	slowRequestTarget := selectedRangeP95.Targets[2]
	if slowRequestTarget.RefID != "C" ||
		!strings.Contains(slowRequestTarget.Expr, `le="0.8"`) ||
		!strings.Contains(slowRequestTarget.Expr, "chestnut_http_server_request_duration_seconds_bucket") ||
		!strings.Contains(slowRequestTarget.Expr, "chestnut_http_server_request_duration_seconds_count") ||
		strings.Count(slowRequestTarget.Expr, "increase(") != 2 ||
		strings.Count(slowRequestTarget.Expr, "[$__range]") != 2 ||
		strings.Contains(slowRequestTarget.Expr, "requests_total") {
		t.Fatalf("所选时段 P95 最慢路由 Top 10 slow-request ratio query is invalid: %q", slowRequestTarget.Expr)
	}
	slowRequestCountTarget := selectedRangeP95.Targets[3]
	if slowRequestCountTarget.RefID != "D" ||
		!strings.Contains(slowRequestCountTarget.Expr, "clamp_min(") ||
		!strings.Contains(slowRequestCountTarget.Expr, `le="0.8"`) ||
		!strings.Contains(slowRequestCountTarget.Expr, "chestnut_http_server_request_duration_seconds_bucket") ||
		!strings.Contains(slowRequestCountTarget.Expr, "chestnut_http_server_request_duration_seconds_count") ||
		strings.Count(slowRequestCountTarget.Expr, "increase(") != 2 ||
		strings.Count(slowRequestCountTarget.Expr, "[$__range]") != 2 ||
		strings.Contains(slowRequestCountTarget.Expr, "requests_total") {
		t.Fatalf("所选时段 P95 最慢路由 Top 10 slow-request count query is invalid: %q", slowRequestCountTarget.Expr)
	}
	p95Steps := selectedRangeP95.FieldConfig.Defaults.Thresholds.Steps
	if selectedRangeP95.FieldConfig.Defaults.Color.Mode != "thresholds" ||
		len(p95Steps) != 3 || p95Steps[0].Color != "green" || p95Steps[0].Value != nil ||
		p95Steps[1].Color != "orange" || p95Steps[1].Value == nil || *p95Steps[1].Value != 0.5 ||
		p95Steps[2].Color != "red" || p95Steps[2].Value == nil || *p95Steps[2].Value != 0.8 {
		t.Fatal("所选时段 P95 最慢路由 Top 10 must use fixed green/orange/red thresholds at 500ms and 800ms")
	}
	if p95Max, exists := grafanaFieldOverrideNumber(t, selectedRangeP95, "P95", "max"); !exists || p95Max != 1 {
		t.Fatalf("所选时段 P95 最慢路由 Top 10 P95 max = %v, want 1 second", p95Max)
	}

	selectedRangeTraffic := panels["所选时段请求数量最高的路由 Top 10"]
	if selectedRangeTraffic.Type != "table" || len(selectedRangeTraffic.Targets) != 4 {
		t.Fatal("所选时段请求数量最高的路由 Top 10 must combine sample count, P95, slow-request count, and slow-request ratio")
	}
	for index, target := range selectedRangeTraffic.Targets {
		if !target.Instant || target.Format != "table" {
			t.Fatalf("所选时段请求数量最高的路由 Top 10 target[%d] must be an instant table query", index)
		}
		if strings.Count(target.Expr, `route!="/api/*path"`) != strings.Count(target.Expr, "increase(") {
			t.Fatalf("所选时段请求数量最高的路由 Top 10 target[%d] must exclude /api/*path from every range aggregation: %q",
				index, target.Expr)
		}
	}
	gotTrafficTransformations := make([]string, 0, len(selectedRangeTraffic.Transformations))
	for _, transformation := range selectedRangeTraffic.Transformations {
		gotTrafficTransformations = append(gotTrafficTransformations, transformation.ID)
	}
	if strings.Join(gotTrafficTransformations, ",") != "joinByField,sortBy,organize" ||
		selectedRangeTraffic.Transformations[0].Options.ByField != "route" ||
		selectedRangeTraffic.Transformations[0].Options.Mode != "inner" {
		t.Fatalf("所选时段请求数量最高的路由 Top 10 must inner-join table queries by route: %v", gotTrafficTransformations)
	}
	if sortOptions := selectedRangeTraffic.Transformations[1].Options.Sort; len(sortOptions) != 1 ||
		!sortOptions[0].Descending || sortOptions[0].Field != "Value #A" {
		t.Fatal("所选时段请求数量最高的路由 Top 10 must sort the joined table by request count descending")
	}
	renamedTrafficFields := selectedRangeTraffic.Transformations[2].Options.RenameByName
	if renamedTrafficFields["route"] != "Route" || renamedTrafficFields["Value #A"] != "请求数" ||
		renamedTrafficFields["Value #B"] != "P95" || renamedTrafficFields["Value #C"] != ">800ms 比例" ||
		renamedTrafficFields["Value #D"] != ">800ms 请求数" {
		t.Fatalf("所选时段请求数量最高的路由 Top 10 renamed fields = %v", renamedTrafficFields)
	}
	indexedTrafficFields := selectedRangeTraffic.Transformations[2].Options.IndexByName
	if indexedTrafficFields["route"] != 0 || indexedTrafficFields["Value #A"] != 1 ||
		indexedTrafficFields["Value #B"] != 2 || indexedTrafficFields["Value #D"] != 3 ||
		indexedTrafficFields["Value #C"] != 4 {
		t.Fatalf("所选时段请求数量最高的路由 Top 10 field order = %v", indexedTrafficFields)
	}
	trafficCountTarget := selectedRangeTraffic.Targets[0]
	if trafficCountTarget.RefID != "A" ||
		!strings.Contains(trafficCountTarget.Expr, "topk(10") ||
		!strings.Contains(trafficCountTarget.Expr, "sum by (route)") ||
		!strings.Contains(trafficCountTarget.Expr, "chestnut_http_server_request_duration_seconds_count") ||
		strings.Count(trafficCountTarget.Expr, "increase(") != 1 ||
		strings.Count(trafficCountTarget.Expr, "[$__range]") != 1 ||
		strings.Contains(trafficCountTarget.Expr, "requests_total") {
		t.Fatalf("所选时段请求数量最高的路由 Top 10 request-count query is invalid: %q", trafficCountTarget.Expr)
	}
	trafficP95Target := selectedRangeTraffic.Targets[1]
	if trafficP95Target.RefID != "B" ||
		!strings.Contains(trafficP95Target.Expr, "histogram_quantile(0.95") ||
		!strings.Contains(trafficP95Target.Expr, "sum by (le, route)") ||
		!strings.Contains(trafficP95Target.Expr, "chestnut_http_server_request_duration_seconds_bucket") ||
		strings.Count(trafficP95Target.Expr, "increase(") != 1 ||
		strings.Count(trafficP95Target.Expr, "[$__range]") != 1 {
		t.Fatalf("所选时段请求数量最高的路由 Top 10 P95 query is invalid: %q", trafficP95Target.Expr)
	}
	trafficSlowRequestTarget := selectedRangeTraffic.Targets[2]
	if trafficSlowRequestTarget.RefID != "C" ||
		!strings.Contains(trafficSlowRequestTarget.Expr, `le="0.8"`) ||
		!strings.Contains(trafficSlowRequestTarget.Expr, "chestnut_http_server_request_duration_seconds_bucket") ||
		!strings.Contains(trafficSlowRequestTarget.Expr, "chestnut_http_server_request_duration_seconds_count") ||
		strings.Count(trafficSlowRequestTarget.Expr, "increase(") != 2 ||
		strings.Count(trafficSlowRequestTarget.Expr, "[$__range]") != 2 ||
		strings.Contains(trafficSlowRequestTarget.Expr, "requests_total") {
		t.Fatalf("所选时段请求数量最高的路由 Top 10 slow-request ratio query is invalid: %q", trafficSlowRequestTarget.Expr)
	}
	trafficSlowRequestCountTarget := selectedRangeTraffic.Targets[3]
	if trafficSlowRequestCountTarget.RefID != "D" ||
		!strings.Contains(trafficSlowRequestCountTarget.Expr, "clamp_min(") ||
		!strings.Contains(trafficSlowRequestCountTarget.Expr, `le="0.8"`) ||
		!strings.Contains(trafficSlowRequestCountTarget.Expr, "chestnut_http_server_request_duration_seconds_bucket") ||
		!strings.Contains(trafficSlowRequestCountTarget.Expr, "chestnut_http_server_request_duration_seconds_count") ||
		strings.Count(trafficSlowRequestCountTarget.Expr, "increase(") != 2 ||
		strings.Count(trafficSlowRequestCountTarget.Expr, "[$__range]") != 2 ||
		strings.Contains(trafficSlowRequestCountTarget.Expr, "requests_total") {
		t.Fatalf("所选时段请求数量最高的路由 Top 10 slow-request count query is invalid: %q", trafficSlowRequestCountTarget.Expr)
	}
	trafficP95Steps := selectedRangeTraffic.FieldConfig.Defaults.Thresholds.Steps
	if selectedRangeTraffic.FieldConfig.Defaults.Color.Mode != "thresholds" ||
		len(trafficP95Steps) != 3 || trafficP95Steps[0].Color != "green" || trafficP95Steps[0].Value != nil ||
		trafficP95Steps[1].Color != "orange" || trafficP95Steps[1].Value == nil || *trafficP95Steps[1].Value != 0.5 ||
		trafficP95Steps[2].Color != "red" || trafficP95Steps[2].Value == nil || *trafficP95Steps[2].Value != 0.8 {
		t.Fatal("所选时段请求数量最高的路由 Top 10 must use fixed green/orange/red P95 thresholds at 500ms and 800ms")
	}
	if p95Max, exists := grafanaFieldOverrideNumber(t, selectedRangeTraffic, "P95", "max"); !exists || p95Max != 1 {
		t.Fatalf("所选时段请求数量最高的路由 Top 10 P95 max = %v, want 1 second", p95Max)
	}
	wantLatencyTableWidths := map[string]float64{
		"Route":      380,
		"请求数":        90,
		"P95":        160,
		">800ms 请求数": 140,
		">800ms 比例":  130,
	}
	for _, panel := range []grafanaPanel{selectedRangeP95, selectedRangeTraffic} {
		for field, wantWidth := range wantLatencyTableWidths {
			width, exists := grafanaFieldOverrideNumber(t, panel, field, "custom.width")
			if !exists || width != wantWidth {
				t.Fatalf("%s field %q width = %v, want %v", panel.Title, field, width, wantWidth)
			}
		}
	}

	if selectedRangeP95.GridPos.X != 0 || selectedRangeP95.GridPos.Y != 11 ||
		selectedRangeP95.GridPos.Height != 13 || selectedRangeP95.GridPos.Width != 12 ||
		selectedRangeTraffic.GridPos.X != 12 || selectedRangeTraffic.GridPos.Y != 11 ||
		selectedRangeTraffic.GridPos.Height != 13 || selectedRangeTraffic.GridPos.Width != 12 {
		t.Fatal("route dashboard must place both selected-range Top 10 tables side by side below the heatmap")
	}

	route5xxTop10 := panels["各路由 5xx 比例最高的 Top10"]
	if route5xxTop10.Type != "table" || len(route5xxTop10.Targets) != 3 {
		t.Fatal("各路由 5xx 比例最高的 Top10 must combine the 5xx ratio, 5xx count, and request count")
	}
	for index, target := range route5xxTop10.Targets {
		if !target.Instant || target.Format != "table" {
			t.Fatalf("各路由 5xx 比例最高的 Top10 target[%d] must be an instant table query", index)
		}
		if strings.Count(target.Expr, `route!="/api/*path"`) != strings.Count(target.Expr, "increase(") {
			t.Fatalf("各路由 5xx 比例最高的 Top10 target[%d] must exclude /api/*path from every range aggregation: %q",
				index, target.Expr)
		}
	}
	gotRoute5xxTransformations := make([]string, 0, len(route5xxTop10.Transformations))
	for _, transformation := range route5xxTop10.Transformations {
		gotRoute5xxTransformations = append(gotRoute5xxTransformations, transformation.ID)
	}
	if strings.Join(gotRoute5xxTransformations, ",") != "joinByField,sortBy,organize" ||
		route5xxTop10.Transformations[0].Options.ByField != "route" ||
		route5xxTop10.Transformations[0].Options.Mode != "inner" {
		t.Fatalf("各路由 5xx 比例最高的 Top10 must inner-join table queries by route: %v", gotRoute5xxTransformations)
	}
	if sortOptions := route5xxTop10.Transformations[1].Options.Sort; len(sortOptions) != 1 ||
		!sortOptions[0].Descending || sortOptions[0].Field != "Value #A" {
		t.Fatal("各路由 5xx 比例最高的 Top10 must sort the joined table by 5xx ratio descending")
	}
	renamedRoute5xxFields := route5xxTop10.Transformations[2].Options.RenameByName
	if renamedRoute5xxFields["route"] != "Route" || renamedRoute5xxFields["Value #A"] != "5xx 比例" ||
		renamedRoute5xxFields["Value #B"] != "请求数" || renamedRoute5xxFields["Value #C"] != "5xx 次数" {
		t.Fatalf("各路由 5xx 比例最高的 Top10 renamed fields = %v", renamedRoute5xxFields)
	}
	indexedRoute5xxFields := route5xxTop10.Transformations[2].Options.IndexByName
	if indexedRoute5xxFields["route"] != 0 || indexedRoute5xxFields["Value #B"] != 1 ||
		indexedRoute5xxFields["Value #C"] != 2 || indexedRoute5xxFields["Value #A"] != 3 {
		t.Fatalf("各路由 5xx 比例最高的 Top10 field order = %v", indexedRoute5xxFields)
	}
	route5xxRatioTarget := route5xxTop10.Targets[0]
	if route5xxRatioTarget.RefID != "A" ||
		!strings.Contains(route5xxRatioTarget.Expr, "topk(10") ||
		!strings.Contains(route5xxRatioTarget.Expr, `status_class="5xx"`) ||
		!strings.Contains(route5xxRatioTarget.Expr, "> 0)") ||
		!strings.Contains(route5xxRatioTarget.Expr, "chestnut_http_server_requests_total") ||
		strings.Contains(route5xxRatioTarget.Expr, "or 0 *") ||
		strings.Count(route5xxRatioTarget.Expr, "increase(") != 2 ||
		strings.Count(route5xxRatioTarget.Expr, "[$__range]") != 2 ||
		strings.Contains(route5xxRatioTarget.Expr, "rate(") {
		t.Fatalf("各路由 5xx 比例最高的 Top10 ratio query is invalid: %q", route5xxRatioTarget.Expr)
	}
	routeRequestCountTarget := route5xxTop10.Targets[1]
	if routeRequestCountTarget.RefID != "B" ||
		!strings.Contains(routeRequestCountTarget.Expr, "sum by (route)") ||
		!strings.Contains(routeRequestCountTarget.Expr, "chestnut_http_server_requests_total") ||
		!strings.Contains(routeRequestCountTarget.Expr, `status_class="5xx"`) ||
		!strings.Contains(routeRequestCountTarget.Expr, "and on (route)") ||
		!strings.Contains(routeRequestCountTarget.Expr, "topk(10") ||
		!strings.Contains(routeRequestCountTarget.Expr, "> 0)") ||
		strings.Count(routeRequestCountTarget.Expr, "increase(") != 3 ||
		strings.Count(routeRequestCountTarget.Expr, "[$__range]") != 3 ||
		strings.Contains(routeRequestCountTarget.Expr, "rate(") {
		t.Fatalf("各路由 5xx 比例最高的 Top10 request-count query is invalid: %q", routeRequestCountTarget.Expr)
	}
	route5xxCountTarget := route5xxTop10.Targets[2]
	if route5xxCountTarget.RefID != "C" ||
		!strings.Contains(route5xxCountTarget.Expr, `status_class="5xx"`) ||
		!strings.Contains(route5xxCountTarget.Expr, "and on (route)") ||
		!strings.Contains(route5xxCountTarget.Expr, "topk(10") ||
		!strings.Contains(route5xxCountTarget.Expr, "> 0)") ||
		strings.Count(route5xxCountTarget.Expr, "increase(") != 3 ||
		strings.Count(route5xxCountTarget.Expr, "[$__range]") != 3 ||
		strings.Contains(route5xxCountTarget.Expr, "rate(") {
		t.Fatalf("各路由 5xx 比例最高的 Top10 5xx-count query is invalid: %q", route5xxCountTarget.Expr)
	}
	if route5xxTop10.FieldConfig.Defaults.NoValue != "所选时段无 HTTP 5xx" {
		t.Fatalf("各路由 5xx 比例最高的 Top10 noValue = %q, want selected-range no-5xx semantics",
			route5xxTop10.FieldConfig.Defaults.NoValue)
	}
	route5xxSteps := route5xxTop10.FieldConfig.Defaults.Thresholds.Steps
	if route5xxTop10.FieldConfig.Defaults.Color.Mode != "thresholds" ||
		route5xxTop10.FieldConfig.Defaults.Unit != "percent" ||
		len(route5xxSteps) != 3 || route5xxSteps[0].Color != "green" || route5xxSteps[0].Value != nil ||
		route5xxSteps[1].Color != "yellow" || route5xxSteps[1].Value == nil || *route5xxSteps[1].Value != 1 ||
		route5xxSteps[2].Color != "red" || route5xxSteps[2].Value == nil || *route5xxSteps[2].Value != 5 {
		t.Fatal("各路由 5xx 比例最高的 Top10 must use percentage thresholds at 1% and 5%")
	}
	if route5xxTop10.GridPos.X != 0 || route5xxTop10.GridPos.Y != 25 ||
		route5xxTop10.GridPos.Width != 12 || route5xxTop10.GridPos.Height != selectedRangeP95.GridPos.Height {
		t.Fatal("各路由 5xx 比例最高的 Top10 must occupy the left half of the route-analysis row")
	}

	routePanicTop10 := panels["各路由 Recovery Panic 次数最高的 Top10"]
	if routePanicTop10.Type != "table" || len(routePanicTop10.Targets) != 2 {
		t.Fatal("各路由 Recovery Panic 次数最高的 Top10 must combine the panic count and request count")
	}
	for index, target := range routePanicTop10.Targets {
		if !target.Instant || target.Format != "table" {
			t.Fatalf("各路由 Recovery Panic 次数最高的 Top10 target[%d] must be an instant table query", index)
		}
		if strings.Count(target.Expr, `route!="/api/*path"`) != strings.Count(target.Expr, "increase(") {
			t.Fatalf("各路由 Recovery Panic 次数最高的 Top10 target[%d] must exclude /api/*path from every range aggregation: %q",
				index, target.Expr)
		}
	}
	gotRoutePanicTransformations := make([]string, 0, len(routePanicTop10.Transformations))
	for _, transformation := range routePanicTop10.Transformations {
		gotRoutePanicTransformations = append(gotRoutePanicTransformations, transformation.ID)
	}
	if strings.Join(gotRoutePanicTransformations, ",") != "joinByField,sortBy,organize" ||
		routePanicTop10.Transformations[0].Options.ByField != "route" ||
		routePanicTop10.Transformations[0].Options.Mode != "inner" {
		t.Fatalf("各路由 Recovery Panic 次数最高的 Top10 must inner-join table queries by route: %v", gotRoutePanicTransformations)
	}
	if sortOptions := routePanicTop10.Transformations[1].Options.Sort; len(sortOptions) != 1 ||
		!sortOptions[0].Descending || sortOptions[0].Field != "Value #A" {
		t.Fatal("各路由 Recovery Panic 次数最高的 Top10 must sort by panic count descending")
	}
	renamedRoutePanicFields := routePanicTop10.Transformations[2].Options.RenameByName
	if renamedRoutePanicFields["route"] != "Route" ||
		renamedRoutePanicFields["Value #A"] != "Recovery Panic 次数" ||
		renamedRoutePanicFields["Value #B"] != "请求数" {
		t.Fatalf("各路由 Recovery Panic 次数最高的 Top10 renamed fields = %v", renamedRoutePanicFields)
	}
	indexedRoutePanicFields := routePanicTop10.Transformations[2].Options.IndexByName
	if indexedRoutePanicFields["route"] != 0 || indexedRoutePanicFields["Value #B"] != 1 ||
		indexedRoutePanicFields["Value #A"] != 2 {
		t.Fatalf("各路由 Recovery Panic 次数最高的 Top10 field order = %v", indexedRoutePanicFields)
	}
	routePanicCountTarget := routePanicTop10.Targets[0]
	if routePanicCountTarget.RefID != "A" ||
		!strings.Contains(routePanicCountTarget.Expr, "topk(10") ||
		!strings.Contains(routePanicCountTarget.Expr, "chestnut_http_server_recovered_panics_total") ||
		!strings.Contains(routePanicCountTarget.Expr, "> 0)") ||
		strings.Count(routePanicCountTarget.Expr, "increase(") != 1 ||
		strings.Count(routePanicCountTarget.Expr, "[$__range]") != 1 ||
		strings.Contains(routePanicCountTarget.Expr, "rate(") {
		t.Fatalf("各路由 Recovery Panic 次数最高的 Top10 panic-count query is invalid: %q", routePanicCountTarget.Expr)
	}
	panicRouteRequestCountTarget := routePanicTop10.Targets[1]
	if panicRouteRequestCountTarget.RefID != "B" ||
		!strings.Contains(panicRouteRequestCountTarget.Expr, "sum by (route)") ||
		!strings.Contains(panicRouteRequestCountTarget.Expr, "chestnut_http_server_requests_total") ||
		!strings.Contains(panicRouteRequestCountTarget.Expr, "chestnut_http_server_recovered_panics_total") ||
		!strings.Contains(panicRouteRequestCountTarget.Expr, "and on (route)") ||
		!strings.Contains(panicRouteRequestCountTarget.Expr, "topk(10") ||
		!strings.Contains(panicRouteRequestCountTarget.Expr, "> 0)") ||
		strings.Count(panicRouteRequestCountTarget.Expr, "increase(") != 2 ||
		strings.Count(panicRouteRequestCountTarget.Expr, "[$__range]") != 2 ||
		strings.Contains(panicRouteRequestCountTarget.Expr, "rate(") ||
		strings.Contains(panicRouteRequestCountTarget.Expr, `status_class=`) {
		t.Fatalf("各路由 Recovery Panic 次数最高的 Top10 request-count query is invalid: %q", panicRouteRequestCountTarget.Expr)
	}
	if routePanicTop10.FieldConfig.Defaults.NoValue != "所选时段无 Recovery Panic" {
		t.Fatalf("各路由 Recovery Panic 次数最高的 Top10 noValue = %q, want selected-range no-panic semantics",
			routePanicTop10.FieldConfig.Defaults.NoValue)
	}
	routePanicSteps := routePanicTop10.FieldConfig.Defaults.Thresholds.Steps
	if routePanicTop10.FieldConfig.Defaults.Color.Mode != "thresholds" ||
		routePanicTop10.FieldConfig.Defaults.Unit != "short" ||
		len(routePanicSteps) != 2 || routePanicSteps[0].Color != "green" || routePanicSteps[0].Value != nil ||
		routePanicSteps[1].Color != "red" || routePanicSteps[1].Value == nil || *routePanicSteps[1].Value != 1 {
		t.Fatal("各路由 Recovery Panic 次数最高的 Top10 must turn red at one panic")
	}
	if routePanicTop10.GridPos.X != 12 || routePanicTop10.GridPos.Y != 25 ||
		routePanicTop10.GridPos.Width != 12 || routePanicTop10.GridPos.Height != selectedRangeP95.GridPos.Height {
		t.Fatal("各路由 Recovery Panic 次数最高的 Top10 must occupy the right half of the route-analysis row")
	}
	for _, removedTitle := range []string{"路由 QPS Top 10", "各路由 5xx 比例", "请求明细"} {
		if _, exists := panels[removedTitle]; exists {
			t.Fatalf("removed route-analysis panel %q must not remain in the dashboards", removedTitle)
		}
	}

	latencyHeatmap := panels["整体请求延迟分布热力图"]
	if latencyHeatmap.Type != "heatmap" || len(latencyHeatmap.Targets) != 1 ||
		latencyHeatmap.Targets[0].Format != "heatmap" || latencyHeatmap.Targets[0].Instant {
		t.Fatal("整体请求延迟分布热力图 must use one range heatmap query")
	}
	if expression := latencyHeatmap.Targets[0].Expr; !strings.Contains(expression, "sum by (le)") ||
		!strings.Contains(expression, "chestnut_http_server_request_duration_seconds_bucket") ||
		!strings.Contains(expression, "[$__rate_interval]") ||
		strings.Contains(expression, "le, route") || strings.Contains(expression, `route=~`) {
		t.Fatalf("整体请求延迟分布热力图 must aggregate all routes only by latency bucket: %q", expression)
	}
	if latencyHeatmap.GridPos.X != 0 || latencyHeatmap.GridPos.Y != 1 ||
		latencyHeatmap.GridPos.Width != 24 || latencyHeatmap.GridPos.Height != 10 {
		t.Fatal("整体请求延迟分布热力图 must occupy the first full-width row in the latency section")
	}
	heatmapColor := latencyHeatmap.Options.Color
	if heatmapColor.Mode != "scheme" || heatmapColor.Scheme != "YlOrRd" ||
		heatmapColor.Reverse || heatmapColor.Steps != 64 {
		t.Fatal("整体请求延迟分布热力图 must use a 64-step low-to-high YlOrRd color scheme")
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
	node5xxQPS := panels["各节点 5xx QPS"]
	if expression := node5xxQPS.Targets[0].Expr; node5xxQPS.FieldConfig.Defaults.Unit != "reqps" ||
		!strings.Contains(expression, `status_class="5xx"`) ||
		!strings.Contains(expression, "rate(") {
		t.Fatalf("各节点 5xx QPS must retain its HTTP 5xx request-rate semantics: %q", expression)
	}
	nodeRecoveryPanics := panels["各节点 Recovery Panic 趋势"]
	nodeHTTPInFlight := panels["各节点 HTTP 并发请求趋势"]
	nodeWebSocketConnections := panels["各节点 WebSocket 连接数趋势"]
	if node5xxQPS.GridPos.X != 0 || node5xxQPS.GridPos.Y != 10 ||
		node5xxQPS.GridPos.Width != 12 || node5xxQPS.GridPos.Height != 8 ||
		nodeRecoveryPanics.GridPos.X != 12 || nodeRecoveryPanics.GridPos.Y != 10 ||
		nodeRecoveryPanics.GridPos.Width != 12 || nodeRecoveryPanics.GridPos.Height != 8 ||
		nodeHTTPInFlight.GridPos.X != 0 || nodeHTTPInFlight.GridPos.Y != 18 ||
		nodeHTTPInFlight.GridPos.Width != 12 || nodeHTTPInFlight.GridPos.Height != 8 ||
		nodeWebSocketConnections.GridPos.X != 12 || nodeWebSocketConnections.GridPos.Y != 18 ||
		nodeWebSocketConnections.GridPos.Width != 12 || nodeWebSocketConnections.GridPos.Height != 8 ||
		panels["进程 CPU"].GridPos.Y != 27 || panels["各节点 Go 协程数量趋势"].GridPos.Y != 43 {
		t.Fatal("instance-runtime dashboard must use paired node-level HTTP and WebSocket concurrency panels")
	}
	if expression := nodeHTTPInFlight.Targets[0].Expr; nodeHTTPInFlight.FieldConfig.Defaults.Custom.ShowPoints != "never" ||
		!strings.Contains(expression, "clamp_min(") ||
		!strings.Contains(expression, "chestnut_http_server_requests_in_flight") ||
		!strings.Contains(expression, "chestnut_http_server_websocket_connections") ||
		strings.Count(expression, "or 0 * max by (node) (up{") != 2 {
		t.Fatalf("各节点 HTTP 并发请求趋势 must subtract WebSocket connections and hide points: %q", expression)
	}
	if expression := nodeWebSocketConnections.Targets[0].Expr; nodeWebSocketConnections.FieldConfig.Defaults.Custom.ShowPoints != "never" ||
		!strings.Contains(expression, "sum by (node) (chestnut_http_server_websocket_connections") ||
		!strings.Contains(expression, "or 0 * max by (node) (up{") {
		t.Fatalf("各节点 WebSocket 连接数趋势 must aggregate by node, fill healthy nodes with zero, and hide points: %q", expression)
	}
	for _, panelTitle := range []string{"各节点 QPS", "各节点 P95", "各节点 5xx QPS"} {
		panel := panels[panelTitle]
		expression := panel.Targets[0].Expr
		if panel.FieldConfig.Defaults.Custom.ShowPoints != "never" ||
			!strings.Contains(expression, "or 0 * max by (node) (up{") ||
			!strings.Contains(expression, "== 1)") {
			t.Fatalf("panel %q must hide points and fill zero only for healthy nodes: %q", panelTitle, expression)
		}
	}
	if expression := panels["各节点 P95"].Targets[0].Expr; !strings.Contains(expression, "and on (node)") ||
		!strings.Contains(expression, "chestnut_http_server_request_duration_seconds_count") ||
		!strings.Contains(expression, "> 0)") {
		t.Fatalf("各节点 P95 must discard zero-sample NaN values before filling healthy nodes with zero: %q", expression)
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
	for _, panelTitle := range []string{"Recovery Panic 趋势", "各节点 Recovery Panic 趋势"} {
		panel := panels[panelTitle]
		expression := panel.Targets[0].Expr
		if panel.FieldConfig.Defaults.Custom.ShowPoints != "never" ||
			!strings.Contains(expression, "or 0 * max by (node) (up{") ||
			!strings.Contains(expression, "== 1)") {
			t.Fatalf("%s must hide points and fill zero only for healthy nodes: %q", panelTitle, expression)
		}
	}
	goGCPauseShare := panels["各节点 GC 暂停时间占比趋势"]
	if expression := goGCPauseShare.Targets[0].Expr; goGCPauseShare.FieldConfig.Defaults.Unit != "percentunit" ||
		goGCPauseShare.FieldConfig.Defaults.Custom.ShowPoints != "never" ||
		!strings.Contains(expression, "rate(go_gc_duration_seconds_sum") ||
		!strings.Contains(expression, "[$__rate_interval]") ||
		!strings.Contains(expression, "or 0 * max by (node) (up{") ||
		!strings.Contains(expression, "== 1)") {
		t.Fatalf("各节点 GC 暂停时间占比趋势 must use GC pause seconds per second, fill healthy nodes with zero, and hide points: %q",
			expression)
	}
	goGCFrequency := panels["各节点 GC 频率趋势"]
	if expression := goGCFrequency.Targets[0].Expr; goGCFrequency.FieldConfig.Defaults.Unit != "ops" ||
		goGCFrequency.FieldConfig.Defaults.Custom.ShowPoints != "never" ||
		!strings.Contains(expression, "rate(go_gc_duration_seconds_count") ||
		!strings.Contains(expression, "[$__rate_interval]") ||
		!strings.Contains(expression, "or 0 * max by (node) (up{") ||
		!strings.Contains(expression, "== 1)") {
		t.Fatalf("各节点 GC 频率趋势 must use the default Go GC counter, fill healthy nodes with zero, and hide points: %q",
			expression)
	}
	processCPU := panels["进程 CPU"]
	processCPUShare := panels["进程 CPU 占整机比例"]
	processRSS := panels["进程常驻内存"]
	goHeap := panels["Go 堆当前已分配内存"]
	goAllocationRate := panels["Go 内存分配速率"]
	for _, panelTitle := range []string{
		"进程 CPU", "进程 CPU 占整机比例", "进程常驻内存", "Go 堆当前已分配内存",
		"Go 内存分配速率", "各节点 Go 协程数量趋势",
	} {
		if panels[panelTitle].FieldConfig.Defaults.Custom.ShowPoints != "never" {
			t.Fatalf("panel %q must hide data points", panelTitle)
		}
	}
	if processCPU.GridPos.X != 0 || processCPU.GridPos.Y != 27 ||
		processCPU.GridPos.Width != 12 || processCPU.GridPos.Height != 8 ||
		processCPUShare.GridPos.X != 12 || processCPUShare.GridPos.Y != 27 ||
		processCPUShare.GridPos.Width != 12 || processCPUShare.GridPos.Height != 8 ||
		processRSS.GridPos.X != 0 || processRSS.GridPos.Y != 35 || processRSS.GridPos.Width != 8 ||
		goHeap.GridPos.X != 8 || goHeap.GridPos.Y != 35 || goHeap.GridPos.Width != 8 ||
		goAllocationRate.GridPos.X != 16 || goAllocationRate.GridPos.Y != 35 ||
		goAllocationRate.GridPos.Width != 8 ||
		panels["各节点 Go 协程数量趋势"].GridPos.Y != 43 || panels["各节点 GC 频率趋势"].GridPos.Y != 43 ||
		panels["各节点 GC 暂停时间占比趋势"].GridPos.Y != 43 {
		t.Fatal("instance-runtime dashboard must pair CPU panels, place three memory panels, then place Go counters below")
	}
	if expression := goAllocationRate.Targets[0].Expr; goAllocationRate.FieldConfig.Defaults.Unit != "Bps" ||
		!strings.Contains(expression, "rate(go_memstats_alloc_bytes_total") ||
		!strings.Contains(expression, "[$__rate_interval]") {
		t.Fatalf("Go 内存分配速率 must display the per-second allocation rate using the dashboard rate interval: %q",
			expression)
	}
	if expression := processCPUShare.Targets[0].Expr; processCPUShare.FieldConfig.Defaults.Unit != "percent" ||
		processCPUShare.FieldConfig.Defaults.Custom.ShowPoints != "never" ||
		!strings.HasPrefix(expression, "100 * ") ||
		!strings.Contains(expression, "sum by (node) (rate(process_cpu_seconds_total") ||
		!strings.Contains(expression, "max by (node) (go_sched_gomaxprocs_threads") {
		t.Fatalf("进程 CPU 占整机比例 must divide process CPU cores by node GOMAXPROCS and display percent without points: %q",
			expression)
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

	httpInFlight := panels["HTTP 并发请求趋势"]
	if httpInFlight.GridPos.Width != 12 || httpInFlight.GridPos.X != 0 || httpInFlight.GridPos.Y != 20 ||
		httpInFlight.Type != "timeseries" || len(httpInFlight.Targets) != 1 || httpInFlight.Targets[0].Instant ||
		httpInFlight.FieldConfig.Defaults.Custom.ShowPoints != "never" ||
		httpInFlight.Options.Tooltip.Mode != "multi" ||
		!strings.Contains(httpInFlight.Targets[0].Expr, "chestnut_http_server_requests_in_flight") ||
		!strings.Contains(httpInFlight.Targets[0].Expr, "chestnut_http_server_websocket_connections") ||
		strings.Count(httpInFlight.Targets[0].Expr, "vector(0)") != 2 {
		t.Fatalf("HTTP 并发请求趋势 must hide points and fill missing data with zero")
	}

	webSocketConnections := panels["WebSocket 连接数趋势"]
	if webSocketConnections.GridPos.Width != 12 || webSocketConnections.GridPos.X != 12 ||
		webSocketConnections.GridPos.Y != 20 || webSocketConnections.Type != "timeseries" ||
		len(webSocketConnections.Targets) != 1 || webSocketConnections.Targets[0].Instant ||
		webSocketConnections.FieldConfig.Defaults.Custom.ShowPoints != "never" ||
		webSocketConnections.Options.Tooltip.Mode != "multi" ||
		!strings.Contains(webSocketConnections.Targets[0].Expr, "chestnut_http_server_websocket_connections") ||
		!strings.Contains(webSocketConnections.Targets[0].Expr, "or vector(0)") {
		t.Fatalf("WebSocket 连接数趋势 must hide points and fill missing data with zero")
	}

	for _, panelTitle := range []string{"HTTP 请求延迟分位数", "HTTP 平均响应耗时"} {
		if panels[panelTitle].FieldConfig.Defaults.Custom.ShowPoints != "never" {
			t.Fatalf("%s must hide points", panelTitle)
		}
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
		panels["集群总 QPS"].GridPos.Y != 12 || panels["HTTP 并发请求趋势"].GridPos.Y != 20 ||
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
