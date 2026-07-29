package appconfig

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bpcoder16/Chestnut/v4/appconfig/env"
)

const (
	testPrometheusServiceName = "test-api"
	invalidHTTPStatusCode     = 600
)

func TestAppConfigCheckAllowsPrometheusDisabled(t *testing.T) {
	config := AppConfig{
		Env: env.Option{
			AppName: "test",
			RunMode: env.RunModeTest,
		},
	}

	if err := config.Check(); err != nil {
		t.Fatalf("AppConfig.Check() error = %v, want nil", err)
	}
}

func TestParseConfigLoadsPrometheus(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "app-server.yaml")
	content := []byte(`env:
  appName: "test"
  runMode: "test"
prometheus:
  enabled: true
  serviceName: "test-api"
  excludedPaths:
    - "/metrics"
    - "/health"
  excludedStatusCodes:
    - 404
`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	var config AppConfig
	if err := ParseConfig(configPath, &config); err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if !config.Prometheus.Enabled {
		t.Fatal("Prometheus.Enabled = false, want true")
	}
	if config.Prometheus.ServiceName != testPrometheusServiceName {
		t.Fatalf("Prometheus.ServiceName = %q, want %q", config.Prometheus.ServiceName, testPrometheusServiceName)
	}
	if got := strings.Join(config.Prometheus.ExcludedPaths, ","); got != "/metrics,/health" {
		t.Fatalf("Prometheus.ExcludedPaths = %q, want %q", got, "/metrics,/health")
	}
	if len(config.Prometheus.ExcludedStatusCodes) != 1 || config.Prometheus.ExcludedStatusCodes[0] != http.StatusNotFound {
		t.Fatalf("Prometheus.ExcludedStatusCodes = %v, want [%d]", config.Prometheus.ExcludedStatusCodes, http.StatusNotFound)
	}
}

func TestAppConfigCheckRejectsInvalidPrometheusConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*Prometheus)
		wantMessage string
	}{
		{
			name: "service name",
			mutate: func(config *Prometheus) {
				config.ServiceName = " "
			},
			wantMessage: "serviceName",
		},
		{
			name: "relative excluded path",
			mutate: func(config *Prometheus) {
				config.ExcludedPaths = []string{"health"}
			},
			wantMessage: "excludedPaths",
		},
		{
			name: "excluded path query",
			mutate: func(config *Prometheus) {
				config.ExcludedPaths = []string{"/health?full=1"}
			},
			wantMessage: "excludedPaths",
		},
		{
			name: "excluded path fragment",
			mutate: func(config *Prometheus) {
				config.ExcludedPaths = []string{"/health#full"}
			},
			wantMessage: "excludedPaths",
		},
		{
			name: "status below range",
			mutate: func(config *Prometheus) {
				config.ExcludedStatusCodes = []int{http.StatusContinue - 1}
			},
			wantMessage: "excludedStatusCodes",
		},
		{
			name: "status above range",
			mutate: func(config *Prometheus) {
				config.ExcludedStatusCodes = []int{invalidHTTPStatusCode}
			},
			wantMessage: "excludedStatusCodes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validPrometheusAppConfig()
			tt.mutate(&config.Prometheus)

			err := config.Check()
			if err == nil {
				t.Fatal("AppConfig.Check() error = nil, want error")
			}
			if !strings.Contains(err.Error(), "prometheus.") || !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("AppConfig.Check() error = %q, want prometheus.%s", err, tt.wantMessage)
			}
		})
	}
}

func validPrometheusAppConfig() AppConfig {
	return AppConfig{
		Env: env.Option{
			AppName: "test",
			RunMode: env.RunModeTest,
		},
		Prometheus: Prometheus{
			Enabled:             true,
			ServiceName:         testPrometheusServiceName,
			ExcludedPaths:       []string{"/metrics", "/health"},
			ExcludedStatusCodes: []int{http.StatusNotFound},
		},
	}
}
