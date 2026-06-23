package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsBackendAlive(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer ln.Close()

	if !isBackendAlive(ln.Addr().String()) {
		t.Error("expected alive for open listener")
	}
}

func TestIsBackendDead(t *testing.T) {
	if isBackendAlive("127.0.0.1:1") {
		t.Error("expected dead for closed port")
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

	pool.HealthCheck()
	if !pool.backends[0].IsAlive() {
		t.Error("expected backend to be alive after health check")
	}

	backend.Close()

	pool.HealthCheck()
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
