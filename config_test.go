package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	validJSON := `{"port": 9090, "backends": ["http://localhost:8081", "http://localhost:8082"]}`
	path := writeTemp(t, validJSON)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if len(cfg.Backends) != 2 {
		t.Errorf("expected 2 backends, got %d", len(cfg.Backends))
	}
}

func TestLoadConfigDefaultPort(t *testing.T) {
	json := `{"backends": ["http://localhost:8081"]}`
	path := writeTemp(t, json)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Port)
	}
}

func TestLoadConfigNoBackends(t *testing.T) {
	json := `{"port": 8080, "backends": []}`
	path := writeTemp(t, json)

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for empty backends, got nil")
	}
}

func TestLoadConfigBadJSON(t *testing.T) {
	path := writeTemp(t, `{not valid json}`)

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for bad JSON, got nil")
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadConfigDefaultStrategy(t *testing.T) {
	json := `{"backends": ["http://localhost:8081"]}`
	path := writeTemp(t, json)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Strategy != "round_robin" {
		t.Errorf("expected default strategy round_robin, got %s", cfg.Strategy)
	}
}

func TestLoadConfigLeastConnections(t *testing.T) {
	json := `{"backends": ["http://localhost:8081"], "strategy": "least_connections"}`
	path := writeTemp(t, json)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Strategy != "least_connections" {
		t.Errorf("expected strategy least_connections, got %s", cfg.Strategy)
	}
}

func TestLoadConfigBadStrategy(t *testing.T) {
	json := `{"backends": ["http://localhost:8081"], "strategy": "random"}`
	path := writeTemp(t, json)

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for unknown strategy, got nil")
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}
