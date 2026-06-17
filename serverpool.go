package main

import (
	"net/http/httputil"
	"net/url"
)

type Backend struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
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
		})
	}
	return &ServerPool{backends: backends}, nil
}

func (s *ServerPool) NextIndex() int {
	idx := s.current % uint64(len(s.backends))
	s.current++
	return int(idx)
}

func (s *ServerPool) GetNextBackend() *Backend {
	return s.backends[s.NextIndex()]
}
