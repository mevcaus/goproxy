package main

import (
	"testing"
)

func TestNewServerPool(t *testing.T) {
	urls := []string{"http://localhost:8081", "http://localhost:8082"}
	pool, err := NewServerPool(urls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pool.backends) != 2 {
		t.Errorf("expected 2 backends, got %d", len(pool.backends))
	}
}

func TestNewServerPoolBadURL(t *testing.T) {
	urls := []string{"://bad-url"}
	_, err := NewServerPool(urls)
	if err == nil {
		t.Fatal("expected error for bad URL, got nil")
	}
}

func TestRoundRobinCycles(t *testing.T) {
	urls := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}
	pool, err := NewServerPool(urls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []int{0, 1, 2, 0, 1, 2, 0}
	for i, want := range expected {
		got := pool.NextIndex()
		if got != want {
			t.Errorf("call %d: expected index %d, got %d", i, want, got)
		}
	}
}

func TestGetNextBackendReturnsCorrectHost(t *testing.T) {
	urls := []string{
		"http://host-a:8081",
		"http://host-b:8082",
	}
	pool, err := NewServerPool(urls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b1 := pool.GetNextBackend()
	if b1.URL.Hostname() != "host-a" {
		t.Errorf("expected host-a, got %s", b1.URL.Hostname())
	}

	b2 := pool.GetNextBackend()
	if b2.URL.Hostname() != "host-b" {
		t.Errorf("expected host-b, got %s", b2.URL.Hostname())
	}

	b3 := pool.GetNextBackend()
	if b3.URL.Hostname() != "host-a" {
		t.Errorf("expected wraparound to host-a, got %s", b3.URL.Hostname())
	}
}
