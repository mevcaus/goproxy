package main

import (
	"log"
	"net"
	"time"
)

func isBackendAlive(u string) bool {
	conn, err := net.DialTimeout("tcp", u, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		addr := b.URL.Host
		alive := isBackendAlive(addr)
		b.SetAlive(alive)
		if !alive {
			log.Printf("Backend %s is down", b.URL)
		}
	}
}

func StartHealthCheck(pool *ServerPool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			pool.HealthCheck()
		}
	}()
}
