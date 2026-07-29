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
	prometheusScrapeAssetPath = "conf.example/prometheus/scrape.d/chestnut-api.yaml"
	grafanaDashboardAssetPath = "conf.example/grafana/dashboards/chestnut-api/chestnut-api-overview-dashboard.json"
	prometheusDatasourceUID   = "prometheus"
	minimumScrapeTargets      = 3
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
	Title         string `json:"title"`
	UID           string `json:"uid"`
	Editable      bool   `json:"editable"`
	Refresh       string `json:"refresh"`
	SchemaVersion int    `json:"schemaVersion"`
	Time          struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"time"`
	Templating struct {
		List []grafanaVariable `json:"list"`
	} `json:"templating"`
	Panels []grafanaPanel `json:"panels"`
}

type grafanaVariable struct {
	Name       string            `json:"name"`
	Definition string            `json:"definition"`
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
			NoValue  string                `json:"noValue"`
			Mappings []grafanaValueMapping `json:"mappings"`
		} `json:"defaults"`
	} `json:"fieldConfig"`
	Targets []struct {
		Expr         string `json:"expr"`
		Format       string `json:"format"`
		Instant      bool   `json:"instant"`
		LegendFormat string `json:"legendFormat"`
	} `json:"targets"`
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
	for _, staticConfig := range job.StaticConfigs {
		targetCount += len(staticConfig.Targets)
		if strings.TrimSpace(staticConfig.Labels["node"]) == "" {
			t.Fatalf("static config %v has empty node label", staticConfig.Targets)
		}
		if _, exists := staticConfig.Labels["service"]; exists {
			t.Fatalf("static config %v must not define conflicting service target label", staticConfig.Targets)
		}
		nodes[staticConfig.Labels["node"]] = struct{}{}
		for _, target := range staticConfig.Targets {
			if !strings.HasPrefix(target, "192.0.2.") {
				t.Fatalf("target %q must use RFC 5737 documentation address space", target)
			}
		}
	}
	if targetCount < minimumScrapeTargets || len(nodes) < minimumScrapeTargets {
		t.Fatalf("targets = %d, unique nodes = %d, want at least %d each", targetCount, len(nodes), minimumScrapeTargets)
	}
}

func TestGrafanaDashboardAssetCoversAPIClusterOperations(t *testing.T) {
	raw, err := os.ReadFile(grafanaDashboardAssetPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", grafanaDashboardAssetPath, err)
	}
	var dashboard grafanaDashboardAsset
	if err := json.Unmarshal(raw, &dashboard); err != nil {
		t.Fatalf("json.Unmarshal dashboard: %v", err)
	}

	if dashboard.UID != "chestnut-api-overview" || dashboard.Editable {
		t.Fatalf("dashboard identity = (uid=%q, editable=%v), want stable read-only identity", dashboard.UID, dashboard.Editable)
	}
	if dashboard.Refresh != "30s" || dashboard.Time.From != "now-1h" || dashboard.Time.To != "now" {
		t.Fatalf("dashboard time = (refresh=%q, from=%q, to=%q), want (30s, now-1h, now)", dashboard.Refresh, dashboard.Time.From, dashboard.Time.To)
	}
	if dashboard.SchemaVersion == 0 {
		t.Fatal("dashboard schemaVersion must be set")
	}

	wantVariables := []struct {
		name  string
		query string
	}{
		{name: "job", query: "label_values(up, job)"},
		{name: "node", query: `label_values(up{job=~"$job"}, node)`},
		{name: "instance", query: `label_values(up{job=~"$job",node=~"$node"}, instance)`},
		{name: "service", query: `label_values(chestnut_http_server_requests_in_flight{job=~"$job",node=~"$node",instance=~"$instance"}, service)`},
		{name: "route", query: `label_values(chestnut_http_server_requests_total{job=~"$job",service=~"$service",node=~"$node",instance=~"$instance"}, route)`},
	}
	variables := make(map[string]grafanaVariable, len(dashboard.Templating.List))
	if len(dashboard.Templating.List) != len(wantVariables) {
		t.Fatalf("len(dashboard variables) = %d, want %d", len(dashboard.Templating.List), len(wantVariables))
	}
	for index, variable := range dashboard.Templating.List {
		variables[variable.Name] = variable
		if variable.Datasource.UID != prometheusDatasourceUID {
			t.Fatalf("variable %q datasource UID = %q, want %q", variable.Name, variable.Datasource.UID, prometheusDatasourceUID)
		}
		if strings.TrimSpace(variable.Query.Query) == "" {
			t.Fatalf("variable %q has empty query", variable.Name)
		}
		want := wantVariables[index]
		if variable.Name != want.name || variable.Query.Query != want.query || variable.Definition != want.query {
			t.Fatalf("dashboard variable[%d] = (name=%q, query=%q, definition=%q), want (%q, %q, %q)",
				index, variable.Name, variable.Query.Query, variable.Definition, want.name, want.query, want.query)
		}
	}

	wantPanels := []string{
		"Target 可用数", "Target 不可用数", "Target 可用率", "Target 健康明细", "当前 In-flight", "进程 Uptime",
		"集群总 QPS", "HTTP 状态趋势", "HTTP 5xx 比例",
		"Recovery Panic 增量", "Recovery Panic 趋势", "集群延迟分位数", "平均耗时", "P95 最慢路由 Top 10",
		"路由 QPS Top 10", "各路由 P95", "各路由 5xx 比例", "请求明细", "各实例 QPS",
		"各实例 P95", "各实例 5xx", "各实例 In-flight", "进程 CPU", "进程 RSS", "Go Heap",
		"Goroutine", "GC 次数", "GC Pause",
	}
	panels := make(map[string]grafanaPanel, len(dashboard.Panels))
	panelIDs := make(map[int]string, len(dashboard.Panels))
	allExpressions := make([]string, 0, len(dashboard.Panels))
	for _, panel := range dashboard.Panels {
		panels[panel.Title] = panel
		if previousTitle, exists := panelIDs[panel.ID]; exists {
			t.Fatalf("dashboard panels %q and %q share id %d", previousTitle, panel.Title, panel.ID)
		}
		panelIDs[panel.ID] = panel.Title
		if panel.Type != "row" {
			if panel.Datasource.UID != prometheusDatasourceUID {
				t.Fatalf("panel %q datasource UID = %q, want %q", panel.Title, panel.Datasource.UID, prometheusDatasourceUID)
			}
			if strings.TrimSpace(panel.FieldConfig.Defaults.NoValue) == "" {
				t.Fatalf("panel %q must define noValue semantics", panel.Title)
			}
		}
		for _, target := range panel.Targets {
			allExpressions = append(allExpressions, target.Expr)
		}
	}
	for _, panelTitle := range wantPanels {
		if _, exists := panels[panelTitle]; !exists {
			t.Fatalf("dashboard panel %q not found", panelTitle)
		}
	}
	assertTargetHealthPanels(t, panels)

	expressions := strings.Join(allExpressions, "\n")
	for _, metricName := range []string{
		"chestnut_http_server_requests_total",
		"chestnut_http_server_request_duration_seconds_bucket",
		"chestnut_http_server_requests_in_flight",
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
		t.Fatal("dashboard must not average instance quantiles")
	}
	for _, forbidden := range []string{"de_card_http_", "Response.Code", "nats_"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("dashboard contains forbidden legacy or unrelated value %q", forbidden)
		}
	}

	for _, panelTitle := range []string{"各实例 QPS", "各实例 P95", "各实例 5xx", "各实例 In-flight", "进程 CPU", "进程 RSS", "Go Heap", "Goroutine", "GC 次数", "GC Pause"} {
		panel := panels[panelTitle]
		for _, target := range panel.Targets {
			if !strings.Contains(target.Expr, "instance") || !strings.Contains(target.LegendFormat, "instance") || !strings.Contains(target.LegendFormat, "node") {
				t.Fatalf("panel %q must retain instance and show node/instance legend", panelTitle)
			}
		}
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

	raw, err := os.ReadFile(grafanaDashboardAssetPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", grafanaDashboardAssetPath, err)
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

	const targetFilters = `job=~"$job",node=~"$node",instance=~"$instance"`
	wantExpressions := map[string]string{
		"Target 可用数":  "sum(up{" + targetFilters + "})",
		"Target 不可用数": "count(up{" + targetFilters + "}) - sum(up{" + targetFilters + "})",
		"Target 可用率":  "100 * sum(up{" + targetFilters + "}) / count(up{" + targetFilters + "})",
		"Target 健康明细": "max by (job, node, instance) (up{" + targetFilters + "})",
	}
	for title, wantExpression := range wantExpressions {
		panel, exists := panels[title]
		if !exists {
			t.Fatalf("dashboard panel %q not found", title)
		}
		if panel.FieldConfig.Defaults.NoValue != "无 target 数据" {
			t.Fatalf("panel %q noValue = %q, want 无 target 数据", title, panel.FieldConfig.Defaults.NoValue)
		}
		if len(panel.Targets) != 1 || panel.Targets[0].Expr != wantExpression {
			t.Fatalf("panel %q expression = %v, want %q", title, panel.Targets, wantExpression)
		}
		if strings.Contains(panel.Targets[0].Expr, "$service") || strings.Contains(panel.Targets[0].Expr, "$route") {
			t.Fatalf("target health panel %q must not use service or route filters", title)
		}
	}

	detailPanel := panels["Target 健康明细"]
	if detailPanel.Type != "table" || !detailPanel.Targets[0].Instant || detailPanel.Targets[0].Format != "table" {
		t.Fatalf("Target 健康明细 = (type=%q, instant=%v, format=%q), want table/true/table",
			detailPanel.Type, detailPanel.Targets[0].Instant, detailPanel.Targets[0].Format)
	}
	if len(detailPanel.FieldConfig.Defaults.Mappings) != 1 || detailPanel.FieldConfig.Defaults.Mappings[0].Type != "value" {
		t.Fatalf("Target 健康明细 mappings = %v, want one value mapping", detailPanel.FieldConfig.Defaults.Mappings)
	}
	mappingOptions := detailPanel.FieldConfig.Defaults.Mappings[0].Options
	if mappingOptions["1"].Text != "UP" || mappingOptions["1"].Color != "green" ||
		mappingOptions["0"].Text != "DOWN" || mappingOptions["0"].Color != "red" {
		t.Fatalf("Target 健康明细 mapping options = %v, want 1=green UP and 0=red DOWN", mappingOptions)
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
