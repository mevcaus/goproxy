package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var healthClient = &http.Client{
	Timeout: 2 * time.Second,
}

func isBackendAlive(backendURL, healthPath string) bool {
	url := fmt.Sprintf("%s%s", backendURL, healthPath)
	resp, err := healthClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (s *ServerPool) HealthCheck(healthPath string) {
	for _, b := range s.backends {
		alive := isBackendAlive(b.URL.String(), healthPath)
		b.SetAlive(alive)
		if !alive {
			log.Printf("Backend %s is down", b.URL)
		}
	}
}

func StartHealthCheck(pool *ServerPool, interval time.Duration, healthPath string) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			pool.HealthCheck(healthPath)
		}
	}()
}
