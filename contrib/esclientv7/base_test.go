package esclientv7

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bpcoder16/Chestnut/v4/appconfig/env"
	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/elastic/go-elasticsearch/v7"
)

type capturedLogEntry struct {
	level     log.Level
	keyValues []interface{}
}

type captureLogger struct {
	entries []capturedLogEntry
}

func (l *captureLogger) Log(level log.Level, keyValues ...interface{}) error {
	kvs := append([]interface{}{}, keyValues...)
	l.entries = append(l.entries, capturedLogEntry{level: level, keyValues: kvs})
	return nil
}

func TestShouldLogSearchDebugByRunMode(t *testing.T) {
	oldEnv := env.Default
	t.Cleanup(func() { env.Default = oldEnv })

	cases := []struct {
		name    string
		runMode string
		want    bool
	}{
		{name: "debug", runMode: env.RunModeDebug, want: true},
		{name: "test", runMode: env.RunModeTest, want: true},
		{name: "release", runMode: env.RunModeRelease, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env.Default = env.New(env.Option{RunMode: tc.runMode})
			if got := shouldLogSearchDebug(); got != tc.want {
				t.Fatalf("shouldLogSearchDebug() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSearchLogsDebugOutsideRelease(t *testing.T) {
	oldEnv := env.Default
	t.Cleanup(func() { env.Default = oldEnv })
	env.Default = env.New(env.Option{RunMode: env.RunModeDebug})

	logger := &captureLogger{}
	manager, cleanup := newTestSearchManager(t, logger)
	defer cleanup()

	var hits []struct {
		Source map[string]any `json:"_source"`
	}
	total, maxScore, err := manager.Search(context.Background(), "groupbuy_items_v1_test", map[string]any{
		"query": map[string]any{"match_all": map[string]any{}},
	}, &hits, nil)
	if err != nil {
		t.Fatalf("Search error = %v", err)
	}
	if total.Value != 1 || maxScore != 1.5 || len(hits) != 1 {
		t.Fatalf("Search result total=%#v maxScore=%v hits=%#v", total, maxScore, hits)
	}
	if len(logger.entries) != 1 {
		t.Fatalf("log entries len = %d, want 1", len(logger.entries))
	}
	entry := logger.entries[0]
	if entry.level != log.LevelDebug {
		t.Fatalf("log level = %v, want debug", entry.level)
	}
	assertLogValue(t, entry.keyValues, "ESSearch", "Search")
	assertLogValue(t, entry.keyValues, "index", "groupbuy_items_v1_test")
	assertLogValue(t, entry.keyValues, "total", uint64(1))
	assertLogValue(t, entry.keyValues, "relation", "eq")
}

func TestSearchSkipsDebugInRelease(t *testing.T) {
	oldEnv := env.Default
	t.Cleanup(func() { env.Default = oldEnv })
	env.Default = env.New(env.Option{RunMode: env.RunModeRelease})

	logger := &captureLogger{}
	manager, cleanup := newTestSearchManager(t, logger)
	defer cleanup()

	var hits []struct {
		Source map[string]any `json:"_source"`
	}
	if _, _, err := manager.Search(context.Background(), "groupbuy_items_v1_test", map[string]any{
		"query": map[string]any{"match_all": map[string]any{}},
	}, &hits, nil); err != nil {
		t.Fatalf("Search error = %v", err)
	}
	if len(logger.entries) != 0 {
		t.Fatalf("log entries len = %d, want 0", len(logger.entries))
	}
}

func TestBulkIndexUsesIndexAction(t *testing.T) {
	var gotPath string
	var gotBody string
	manager, cleanup := newTestBulkManager(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.URL.Path != "/_bulk" {
			t.Fatalf("request path = %q, want /_bulk", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		gotBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":false,"items":[{"index":{"_id":"doc-1","status":200}}]}`))
	})
	defer cleanup()

	err := manager.BulkIndex(context.Background(), "checklist_confirmed_cards_v1_test", []Document{
		{ID: "doc-1", Content: map[string]any{"uuid": "doc-1"}},
	})
	if err != nil {
		t.Fatalf("BulkIndex error = %v", err)
	}
	if gotPath != "/_bulk" {
		t.Fatalf("path = %q, want /_bulk", gotPath)
	}
	lines := strings.Split(strings.TrimSpace(gotBody), "\n")
	if len(lines) != 2 {
		t.Fatalf("bulk body lines = %d, want 2: %q", len(lines), gotBody)
	}
	var meta map[string]map[string]string
	if err := json.Unmarshal([]byte(lines[0]), &meta); err != nil {
		t.Fatalf("unmarshal meta line: %v", err)
	}
	if _, ok := meta["update"]; ok {
		t.Fatalf("bulk meta uses update action: %#v", meta)
	}
	indexMeta, ok := meta["index"]
	if !ok {
		t.Fatalf("bulk meta = %#v, want index action", meta)
	}
	if indexMeta["_index"] != "checklist_confirmed_cards_v1_test" || indexMeta["_id"] != "doc-1" {
		t.Fatalf("index meta = %#v", indexMeta)
	}
	var doc map[string]string
	if err := json.Unmarshal([]byte(lines[1]), &doc); err != nil {
		t.Fatalf("unmarshal doc line: %v", err)
	}
	if doc["uuid"] != "doc-1" {
		t.Fatalf("doc line = %#v", doc)
	}
}

func newTestBulkManager(t *testing.T, handler http.HandlerFunc) (*Manager, func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"version":{"number":"7.17.0"}}`))
			return
		}
		handler(w, r)
	}))
	client, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{server.URL}})
	if err != nil {
		server.Close()
		t.Fatalf("new elasticsearch client: %v", err)
	}
	return &Manager{
		client: client,
	}, server.Close
}

func newTestSearchManager(t *testing.T, logger log.Logger) (*Manager, func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"version":{"number":"7.17.0"}}`))
			return
		}
		if r.URL.Path != "/groupbuy_items_v1_test/_search" {
			t.Fatalf("request path = %q, want /groupbuy_items_v1_test/_search", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hits":{"total":{"value":1,"relation":"eq"},"max_score":1.5,"hits":[{"_source":{"id":7}}]}}`))
	}))
	client, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{server.URL}})
	if err != nil {
		server.Close()
		t.Fatalf("new elasticsearch client: %v", err)
	}
	return &Manager{
		logger: log.NewHelper(logger),
		client: client,
	}, server.Close
}

func assertLogValue(t *testing.T, keyValues []interface{}, key string, want interface{}) {
	t.Helper()
	for i := 0; i+1 < len(keyValues); i += 2 {
		if keyValues[i] == key {
			if keyValues[i+1] != want {
				t.Fatalf("log value %s = %#v, want %#v", key, keyValues[i+1], want)
			}
			return
		}
	}
	t.Fatalf("log key %s missing in %#v", key, keyValues)
}
