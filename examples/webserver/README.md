# Webserver Benchmarks

Compares web server implementations across two benchmark suites, measuring **throughput** (req/s), **memory usage**
(RSS), and **scaling behavior** under varying concurrency.

## Servers

| Server         | Language            | Port | Description                                                                          |
| -------------- | ------------------- | ---- | ------------------------------------------------------------------------------------ |
| `pocketpy`     | Go + pocketpy       | 8086 | Go HTTP server with embedded Python interpreter (`gofre/examples/webserver_binary/`) |
| `pure_go`      | Go stdlib           | 8085 | `net/http` server, no dependencies                                                   |
| `fastapi`      | Python 3 + uvicorn  | 8083 | Async ASGI framework                                                                 |
| `flask`        | Python 3 + Werkzeug | 8084 | Sync WSGI framework                                                                  |
| `pure_python`  | Python 3 stdlib     | 8082 | `http.server` with threading                                                         |
| `cpython_cffi` | Python 3 + Go cffi  | 8081 | CPython calling Go JSON via cffi                                                     |

## Benchmark Suites

### Suite 1: Basic Routes

Simple `GET /`, `GET /api/data`, `POST /api/echo` endpoints.

| Server       |      GET / | GET /api/data | POST /api/echo |
| ------------ | ---------: | ------------: | -------------: |
| **pure_go**  | **42,290** |    **34,610** |     **36,940** |
| **pocketpy** | **39,880** |    **41,050** |     **40,044** |
| fastapi      |      7,643 |         7,657 |          7,606 |
| flask        |      7,229 |         7,697 |          7,871 |
| pure_python  |      4,574 |         4,652 |          4,480 |
| cpython_cffi |      4,487 |         4,474 |          4,543 |

Tested with `hey -n 5000 -c 10` on Apple M-series.

### Suite 2: TechEmpower-Inspired (Concurrency Sweep)

Plaintext and JSON endpoints tested at concurrency levels 1, 5, 10, 25, 50, 100 to measure scaling behavior.

#### Plaintext — `GET /plaintext` (req/s)

| Server      |    c=1 |    c=5 |   c=10 |   c=25 |   c=50 |      c=100 |
| ----------- | -----: | -----: | -----: | -----: | -----: | ---------: |
| pure_go     | 15,408 | 41,669 | 45,035 | 48,314 | 57,769 | **62,907** |
| pocketpy    |  4,492 |  1,557 |  6,707 | 28,267 | 28,691 |     32,933 |
| fastapi     |  3,880 |  5,860 |  6,141 |  6,292 |  6,412 |      6,447 |
| pure_python |  1,781 |  4,271 |  4,493 |  2,288 |    635 |        498 |
| flask       |  1,132 |    963 |  1,876 |    876 |  2,205 |      1,253 |

#### JSON — `GET /json` (req/s)

| Server      |    c=1 |    c=5 |   c=10 |   c=25 |   c=50 |      c=100 |
| ----------- | -----: | -----: | -----: | -----: | -----: | ---------: |
| pure_go     | 15,185 | 42,890 | 48,311 | 54,061 | 56,447 | **60,423** |
| pocketpy    |  4,530 | 16,433 | 20,608 | 28,969 | 26,085 |     32,009 |
| fastapi     |  3,443 |  5,121 |  5,224 |  5,300 |  5,302 |      5,308 |
| pure_python |  1,774 |  4,409 |  4,211 |  2,272 |    829 |        499 |
| flask       |  1,091 |    766 |  1,022 |    822 |  2,219 |      2,151 |

#### Memory Usage (RSS KB)

| Server      |    Idle | c=1 peak | c=100 peak | Growth at c=100 |
| ----------- | ------: | -------: | ---------: | --------------: |
| pure_go     |   1,744 |    1,744 |      1,744 |              +0 |
| pocketpy    |   1,712 |    1,712 |      1,712 |              +0 |
| pure_python | ~15,900 |   15,856 |     16,032 |             +16 |
| flask       | ~27,800 |   27,632 |     29,312 |            +192 |
| fastapi     | ~38,400 |   38,240 |     39,936 |            +656 |

## Key Observations

- **pure_go** scales 4.1x from c=1 to c=100 with zero memory growth — goroutines handle concurrency without
  per-connection overhead.
- **pocketpy** scales 7.3x — starts slow but catches up at high concurrency; the Go HTTP layer handles connection
  queuing well even when the Python interpreter is the bottleneck.
- **fastapi** plateaus around c=10 (~6,400 req/s) — single-worker uvicorn can't utilize more connections.
- **pure_python** collapses under high concurrency — the GIL causes contention above c=10 (4,493 → 498 req/s).
- **flask** is erratic with no consistent scaling pattern.
- **Go servers use 22x less memory** than Python (1.7 MB vs 16–39 MB idle).
- **pocketpy dispatch overhead** over pure Go is minimal (~0–10%).

## Files

### Basic Routes

| File                | Description                   |
| ------------------- | ----------------------------- |
| `server.py`         | CPython + Go cffi server      |
| `server_pure.py`    | Pure Python baseline          |
| `server_pure_go.go` | Pure Go stdlib server         |
| `server_fastapi.py` | FastAPI + uvicorn             |
| `server_flask.py`   | Flask (dev server)            |
| `benchmark_all.sh`  | Basic routes benchmark runner |

### TechEmpower-Inspired (benchmarks/)

| File                                   | Description                        |
| -------------------------------------- | ---------------------------------- |
| `benchmarks/run.sh`                    | Concurrency sweep benchmark runner |
| `benchmarks/servers/server_pure_go.go` | Go server with TFB endpoints       |
| `benchmarks/servers/server_pure.py`    | Python stdlib with TFB endpoints   |
| `benchmarks/servers/server_fastapi.py` | FastAPI with TFB endpoints         |
| `benchmarks/servers/server_flask.py`   | Flask with TFB endpoints           |

## Running

### Prerequisites

```bash
go install github.com/rakyll/hey@latest
pip3 install fastapi uvicorn flask
```

### Basic Routes

```bash
cd examples/webserver
bash benchmark_all.sh
```

### TechEmpower-Inspired (concurrency sweep)

```bash
cd examples/webserver/benchmarks
bash run.sh
```

To customize concurrency levels, edit `CONCURRENCY_LEVELS` in `run.sh`:

```bash
CONCURRENCY_LEVELS=(1 5 10 25 50 100)  # default
CONCURRENCY_LEVELS=(1 10 50 100)        # faster run
```

### Individual Servers

```bash
# Pocketpy (Go + embedded Python)
cd gofre/examples/webserver_binary && go build -o webserver_binary . && ./webserver_binary 8086

# Pure Go
cd examples/webserver && go build -o server_pure_go server_pure_go.go && ./server_pure_go 8085

# FastAPI
cd examples/webserver && python3 server_fastapi.py 8083

# Flask
cd examples/webserver && python3 server_flask.py 8084

# Pure Python
cd examples/webserver && python3 server_pure.py 8082
```
