package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewProxy(targetURL string) (http.Handler, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(target), nil
}

func main() {
	targetURL := "http://localhost:8081"
	proxy, err := NewProxy(targetURL)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	log.Printf("Starting proxy server on :8080 forwarding to %s", targetURL)
	if err := http.ListenAndServe(":8080", proxy); err != nil {
		log.Fatal(err)
	}
}
