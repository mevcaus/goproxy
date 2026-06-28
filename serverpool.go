package main

import (
	"math"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL               *url.URL
	ReverseProxy      *httputil.ReverseProxy
	mux               sync.RWMutex
	alive             bool
	activeConnections int64
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.alive = alive
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.alive
}

func (b *Backend) AddConn() {
	atomic.AddInt64(&b.activeConnections, 1)
}

func (b *Backend) RemoveConn() {
	atomic.AddInt64(&b.activeConnections, -1)
}

func (b *Backend) ActiveConnections() int64 {
	return atomic.LoadInt64(&b.activeConnections)
}

type ServerPool struct {
	backends []*Backend
	current  uint64
}

func NewServerPool(urls []string) (*ServerPool, error) {
	var backends []*Backend
	for _, rawURL := range urls {
		u, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}
		backends = append(backends, &Backend{
			URL:          u,
			ReverseProxy: httputil.NewSingleHostReverseProxy(u),
			alive:        true,
		})
	}
	return &ServerPool{backends: backends}, nil
}

func (s *ServerPool) NextIndex() int {
	idx := (atomic.AddUint64(&s.current, 1) - 1) % uint64(len(s.backends))
	return int(idx)
}

func (s *ServerPool) GetNextBackend() *Backend {
	total := len(s.backends)
	next := s.NextIndex()
	for i := 0; i < total; i++ {
		idx := (next + i) % total
		if s.backends[idx].IsAlive() {
			return s.backends[idx]
		}
	}
	return nil
}

func (s *ServerPool) GetLeastConnBackend() *Backend {
	var best *Backend
	var minConn int64 = math.MaxInt64

	for _, b := range s.backends {
		if !b.IsAlive() {
			continue
		}
		conns := b.ActiveConnections()
		if conns < minConn {
			minConn = conns
			best = b
		}
	}
	return best
}
