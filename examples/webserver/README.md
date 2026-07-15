# Webserver Benchmark: pocketpy vs Python frameworks

This directory benchmarks six approaches to running web application code:

1. **Go binary + pocketpy** — Go HTTP server with embedded pocketpy interpreter (`gofre/examples/webserver_binary/`)
2. **Pure Go (stdlib)** — Standard library `net/http` server, no dependencies
3. **CPython + Go cffi** — CPython `http.server` calling Go JSON functions via cffi
4. **Pure Python** — Baseline CPython `http.server` with no Go
5. **FastAPI + uvicorn** — Mature ASGI framework
6. **Flask** — Mature WSGI framework

All servers serve identical routes:

- `GET /` — returns `{"status": "ok", "server": "..."}`
- `GET /api/data` — returns `{"data": [1, 2, 3, 4, 5]}`
- `POST /api/echo` — echoes the JSON body back

## Files

| File                | Description                  |
| ------------------- | ---------------------------- |
| `server.py`         | CPython + Go cffi server     |
| `server_pure.py`    | Pure Python baseline         |
| `server_pure_go.go` | Pure Go stdlib server        |
| `server_fastapi.py` | FastAPI + uvicorn            |
| `server_flask.py`   | Flask (dev server)           |
| `benchmark_all.sh`  | Benchmark runner using `hey` |

The Go binary lives at `gofre/examples/webserver_binary/` and uses a Python `@app.route()` decorator API identical to
Flask/FastAPI patterns.

## Binary Sizes

| Artifact                      | Size   |
| ----------------------------- | ------ |
| Pure Go (stdlib)              | 8.0 MB |
| Go binary (pocketpy embedded) | 9.7 MB |
| Shared library (cffi)         | 2.1 MB |
| Python wheel                  | 903 KB |

## Results

Tested with `hey -n 5000 -c 10` on an Apple M-series machine.

| Server               |            GET / |    GET /api/data |   POST /api/echo |       Binary |
| -------------------- | ---------------: | ---------------: | ---------------: | -----------: |
| **Go + pocketpy**    | **39,880 req/s** | **41,050 req/s** | **40,044 req/s** |   **9.7 MB** |
| **Pure Go (stdlib)** | **42,290 req/s** | **34,610 req/s** | **36,940 req/s** |   **8.0 MB** |
| FastAPI + uvicorn    |      7,643 req/s |      7,657 req/s |      7,606 req/s |            — |
| Flask (dev server)   |      7,229 req/s |      7,697 req/s |      7,871 req/s |            — |
| CPython + Go cffi    |      4,487 req/s |      4,474 req/s |      4,543 req/s | 2.1 MB (lib) |
| Pure Python          |      4,574 req/s |      4,652 req/s |      4,480 req/s |            — |

## Memory Usage

Measured with RSS sampling during `hey -n 5000 -c 10` load.

| Server               |    Idle RSS |    Peak RSS |
| -------------------- | ----------: | ----------: |
| **Pure Go (stdlib)** | **10.3 MB** | **19.0 MB** |
| **Go + pocketpy**    | **13.3 MB** | **21.8 MB** |
| Pure Python          |     15.2 MB |     15.4 MB |
| CPython + Go cffi    |     23.1 MB |     29.6 MB |
| Flask (dev server)   |     26.5 MB |     27.8 MB |
| FastAPI + uvicorn    |     36.9 MB |     37.4 MB |

### Key takeaways

- **Go + pocketpy is ~5.2x faster than FastAPI** and ~5.5x faster than Flask.
- **Pure Go stdlib uses the least memory** of all servers (10.3 MB idle, 19.0 MB peak) and is the fastest overall.
- The single 9.7 MB Go+Pocketpy binary contains the Go runtime, HTTP server, pocketpy interpreter, and the Python
  application source.
- Pocketpy dispatch overhead over pure Go is minimal (~0-10%).
- CPython + cffi showed no improvement over pure Python — the bottleneck is CPython's `http.server`, not JSON
  processing.

## Running

```bash
# Go binary (pocketpy embedded)
cd gofre/examples/webserver_binary && go build -o webserver_binary . && ./webserver_binary

# CPython + Go cffi
cd examples/webserver && gofre build && python3 server.py 8081

# Pure Python
cd examples/webserver && python3 server_pure.py 8082

# FastAPI
cd examples/webserver && python3 server_fastapi.py 8083

# Flask
cd examples/webserver && python3 server_flask.py 8084

# Pure Go (stdlib)
cd examples/webserver && go build -o server_pure_go server_pure_go.go && ./server_pure_go 8085
```

Requires `hey` for benchmarking:

```bash
go install github.com/rakyll/hey@latest
bash examples/benchmark_all.sh
```

To build the Pure Go server:

```bash
cd examples/webserver && go build -o server_pure_go server_pure_go.go
```
