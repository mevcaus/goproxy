package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsBackendAliveHTTP(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer backend.Close()

	if !isBackendAlive(backend.URL, "/health") {
		t.Error("expected alive when /health returns 200")
	}
}

func TestIsBackendDeadHTTP(t *testing.T) {
	if isBackendAlive("http://127.0.0.1:1", "/health") {
		t.Error("expected dead for unreachable host")
	}
}

func TestIsBackendDeadOn500(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer backend.Close()

	if isBackendAlive(backend.URL, "/health") {
		t.Error("expected dead when /health returns 500")
	}
}

func TestHealthCheckMarksDeadBackend(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	pool, err := NewServerPool([]string{backend.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pool.HealthCheck("/health")
	if !pool.backends[0].IsAlive() {
		t.Error("expected backend to be alive after health check")
	}

	backend.Close()

	pool.HealthCheck("/health")
	if pool.backends[0].IsAlive() {
		t.Error("expected backend to be dead after server closed")
	}
}

func TestGetNextBackendSkipsDead(t *testing.T) {
	pool, err := NewServerPool([]string{
		"http://host-a:8081",
		"http://host-b:8082",
		"http://host-c:8083",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pool.backends[0].SetAlive(false)

	b := pool.GetNextBackend()
	if b == nil {
		t.Fatal("expected a backend, got nil")
	}
	if b.URL.Hostname() == "host-a" {
		t.Error("should have skipped dead backend host-a")
	}
}

func TestGetNextBackendAllDead(t *testing.T) {
	pool, err := NewServerPool([]string{
		"http://host-a:8081",
		"http://host-b:8082",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pool.backends[0].SetAlive(false)
	pool.backends[1].SetAlive(false)

	b := pool.GetNextBackend()
	if b != nil {
		t.Errorf("expected nil when all backends are dead, got %s", b.URL)
	}
}
