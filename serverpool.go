package main

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
	mux          sync.RWMutex
	alive        bool
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
	return s.backends[s.NextIndex()]
}
