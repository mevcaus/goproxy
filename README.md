# goproxy

A Layer 7 HTTP load balancer built from scratch in Go using only the standard library.

I built this to deepen my understanding of reverse proxies, load balancing algorithms, and Go concurrency primitives — the kind of systems design knowledge that matters in production infrastructure.

## What It Does

`goproxy` sits in front of a pool of backend HTTP servers and distributes incoming requests across them. It handles backend failures gracefully and supports multiple routing strategies.

```
        ┌────────────────┐
        │    Clients     │
        └───────┬────────┘
                │
        ┌───────▼────────┐
        │    goproxy     │
        │     :8080      │
        └──┬────┬────┬───┘
           │    │    │
  ┌────────┘    │    └────────┐
  │             │             │
  ▼             ▼             ▼
┌─────────┐ ┌─────────┐ ┌─────────┐
│ :8081   │ │ :8082   │ │ :8083   │
│ Backend │ │ Backend │ │ Backend │
└─────────┘ └─────────┘ └─────────┘
```

## Features

- **Reverse Proxying** — Built on `net/http/httputil.ReverseProxy` from the standard library. Handles header forwarding, chunked transfers, and connection management out of the box.
- **Round Robin** — Cycles through backends sequentially using an atomic counter (`sync/atomic`) so concurrent requests never cause a data race on the routing index.
- **Least Connections** — Tracks active connections per backend with atomic counters and routes each new request to the backend currently handling the fewest.
- **Active Health Checking** — A background goroutine hits each backend's `/health` endpoint via HTTP every 10 seconds. Unlike a TCP dial (which only proves the port is open), this verifies the application is actually responding. Dead backends are marked with a `sync.RWMutex`-protected flag so the router skips them without blocking readers.
- **JSON Configuration** — Backends, port, and routing strategy are defined in `config.json`. No recompilation needed to change the topology.

## Architecture

```
config.json ──▶ LoadConfig() ──▶ Config
                                    │
                              NewServerPool()
                                    │
                               ServerPool
                              ┌─────┴─────┐
                              │  backends  │──▶ []*Backend
                              │  current   │       │
                              └────────────┘       ├── URL
                                    │              ├── ReverseProxy
                          ┌─────────┼──────────┐   ├── alive (RWMutex)
                          ▼         ▼          ▼   └── activeConnections (atomic)
                   GetNextBackend  GetLeastConn  HealthCheck
                   (round robin)  (least conns)  (background goroutine)
```

| File | Responsibility |
|---|---|
| `main.go` | HTTP server, request handler, strategy dispatch |
| `config.go` | JSON config parsing and validation |
| `serverpool.go` | `Backend` struct, `ServerPool`, routing algorithms |
| `healthcheck.go` | TCP liveness probes, background health checker |

## Concurrency Model

Go's `net/http` server spawns a new goroutine for every incoming request. This means every shared data structure needs to be safe for concurrent access:

| Shared State | Protection | Why |
|---|---|---|
| Round robin index (`current`) | `sync/atomic.AddUint64` | Lock-free increment; a mutex would be overkill for a single counter |
| Backend alive status (`alive`) | `sync.RWMutex` | Health checker writes infrequently, request handlers read constantly — `RWMutex` allows concurrent reads without blocking |
| Active connection count | `sync/atomic.AddInt64` | Same rationale as the round robin index — a simple counter that needs atomic increment/decrement |

## Getting Started

```bash
git clone https://github.com/mevcaus/goproxy.git
cd goproxy
go build -o goproxy .
```

### Configuration

Edit `config.json`:

```json
{
  "port": 8080,
  "strategy": "round_robin",
  "health_path": "/health",
  "backends": [
    "http://localhost:8081",
    "http://localhost:8082",
    "http://localhost:8083"
  ]
}
```

- `strategy` accepts `round_robin` or `least_connections`.
- `health_path` is the endpoint the health checker will GET on each backend (defaults to `/health`).

### Run

```bash
./goproxy
```

```
2026/06/27 10:00:00 Health checker started (every 10s, path: /health)
2026/06/27 10:00:00 Starting load balancer on :8080 with 3 backends
2026/06/25 12:00:00   -> http://localhost:8081
2026/06/25 12:00:00   -> http://localhost:8082
2026/06/25 12:00:00   -> http://localhost:8083
```

## Testing

```bash
go test -v -race ./...
```

There are 26 tests covering:

- Config parsing (valid input, defaults, strategy validation, health path)
- Round robin cycling and wraparound
- Concurrent access to the round robin counter (1000 goroutines)
- Backend alive status toggling under concurrent read/write
- HTTP health endpoint probing (200 = alive, 500 = dead, unreachable = dead)
- Health checker marking backends dead after server shutdown
- Routing that skips dead backends
- Least connections picking the backend with fewest active connections
- Least connections skipping dead backends
- Graceful 503 when all backends are down

## Built With

Only the Go standard library:

- `net/http` — HTTP server, client, and health check probes
- `net/http/httputil` — `ReverseProxy` implementation
- `net/url` — URL parsing
- `sync` — `RWMutex` for alive status
- `sync/atomic` — Lock-free counters
- `encoding/json` — Config parsing
- `time` — Health check ticker
- `math` — `MaxInt64` for least connections initialization
