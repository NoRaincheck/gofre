# Webserver Benchmark: pocketpy vs Python frameworks

This directory benchmarks three approaches to running Python web application code:

1. **Go binary + pocketpy** — Go HTTP server with embedded pocketpy interpreter (`goforge/examples/webserver_binary/`)
2. **CPython + Go cffi** — CPython `http.server` calling Go JSON functions via cffi
3. **Pure Python** — Baseline CPython `http.server` with no Go
4. **FastAPI + uvicorn** — Mature ASGI framework
5. **Flask** — Mature WSGI framework

All servers serve identical routes:

- `GET /` — returns `{"message": "..."}`
- `GET /api/data` — returns `{"items": [1, 2, 3], "service": "..."}`
- `POST /api/echo` — echoes the JSON body back

## Files

| File                | Description                  |
| ------------------- | ---------------------------- |
| `server.py`         | CPython + Go cffi server     |
| `server_pure.py`    | Pure Python baseline         |
| `server_fastapi.py` | FastAPI + uvicorn            |
| `server_flask.py`   | Flask (dev server)           |
| `benchmark_all.sh`  | Benchmark runner using `hey` |

The Go binary lives at `goforge/examples/webserver_binary/` and uses a Python `@app.route()` decorator API identical to
Flask/FastAPI patterns.

## Binary Sizes

| Artifact                      | Size   |
| ----------------------------- | ------ |
| Go binary (pocketpy embedded) | 9.7 MB |
| Shared library (cffi)         | 2.1 MB |
| Python wheel                  | 903 KB |

## Results

Tested with `hey -n 5000 -c 10` on an Apple M-series machine.

| Server             |            GET / |    GET /api/data |   POST /api/echo |       Binary |
| ------------------ | ---------------: | ---------------: | ---------------: | -----------: |
| **Go + pocketpy**  | **39,880 req/s** | **41,050 req/s** | **40,044 req/s** |   **9.7 MB** |
| FastAPI + uvicorn  |      7,643 req/s |      7,657 req/s |      7,606 req/s |            — |
| Flask (dev server) |      7,229 req/s |      7,697 req/s |      7,871 req/s |            — |
| CPython + Go cffi  |      4,487 req/s |      4,474 req/s |      4,543 req/s | 2.1 MB (lib) |
| Pure Python        |      4,574 req/s |      4,652 req/s |      4,480 req/s |            — |

### Key takeaways

- **Go + pocketpy is ~5.2x faster than FastAPI** and ~5.5x faster than Flask.
- The single 9.7 MB binary contains the Go runtime, HTTP server, pocketpy interpreter, and the Python application
  source.
- Pocketpy dispatch overhead over pure Go is minimal (~0-10%).
- CPython + cffi showed no improvement over pure Python — the bottleneck is CPython's `http.server`, not JSON
  processing.

## Running

```bash
# Go binary (pocketpy embedded)
cd goforge/examples/webserver_binary && go build -o webserver_binary . && ./webserver_binary

# CPython + Go cffi
cd examples/webserver && goforge build && python3 server.py 8081

# Pure Python
cd examples/webserver && python3 server_pure.py 8082

# FastAPI
cd examples/webserver && python3 server_fastapi.py 8083

# Flask
cd examples/webserver && python3 server_flask.py 8084
```

Requires `hey` for benchmarking:

```bash
go install github.com/rakyll/hey@latest
bash examples/benchmark_all.sh
```
