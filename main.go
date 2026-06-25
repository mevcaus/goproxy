package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func NewProxy(targetURL string) (http.Handler, error) {
	pool, err := NewServerPool([]string{targetURL})
	if err != nil {
		return nil, err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backend := pool.GetNextBackend()
		if backend == nil {
			http.Error(w, "all backends are down", http.StatusServiceUnavailable)
			return
		}
		backend.ReverseProxy.ServeHTTP(w, r)
	}), nil
}

func main() {
	cfg, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool, err := NewServerPool(cfg.Backends)
	if err != nil {
		log.Fatalf("Failed to create server pool: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var backend *Backend
		switch cfg.Strategy {
		case "least_connections":
			backend = pool.GetLeastConnBackend()
		default:
			backend = pool.GetNextBackend()
		}
		if backend == nil {
			http.Error(w, "all backends are down", http.StatusServiceUnavailable)
			return
		}
		backend.AddConn()
		defer backend.RemoveConn()
		log.Printf("[%s] Forwarding request to %s (active: %d)", cfg.Strategy, backend.URL, backend.ActiveConnections())
		backend.ReverseProxy.ServeHTTP(w, r)
	})

	StartHealthCheck(pool, 10*time.Second)
	log.Println("Health checker started (every 10s)")

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting load balancer on %s with %d backends", addr, len(cfg.Backends))
	for _, b := range cfg.Backends {
		log.Printf("  -> %s", b)
	}

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
