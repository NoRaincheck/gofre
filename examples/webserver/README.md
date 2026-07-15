# Webserver Benchmarks

Compares web server implementations across two benchmark suites, measuring **throughput** (req/s), **memory usage**
(RSS), and **scaling behavior** under varying concurrency.

## Servers

| Server           | Language            | Port | Description                                                                          |
| ---------------- | ------------------- | ---- | ------------------------------------------------------------------------------------ |
| `pocketpy`       | Go + pocketpy       | 8086 | Go HTTP server with embedded Python interpreter (`gofre/examples/webserver_binary/`) |
| `pure_go`        | Go stdlib           | 8085 | `net/http` server, no dependencies                                                   |
| `fastapi`        | Python 3 + uvicorn  | 8083 | Async ASGI framework                                                                 |
| `flask`          | Python 3 + Werkzeug | 8084 | Sync WSGI framework                                                                  |
| `pure_python`    | Python 3 stdlib     | 8082 | `http.server` with threading                                                         |
| `cpython_cffi`   | Python 3 + Go cffi  | 8081 | CPython calling Go JSON via cffi                                                     |
| `cpython_gohttp` | Python 3 + Go cffi  | 8087 | CPython calling Go JSON (gojson) and Go HTTP (gohttp) via cffi                       |

## Benchmark Suites

### Suite 1: Basic Routes

Simple `GET /`, `GET /api/data`, `POST /api/echo` endpoints.

| Server       |      GET / | GET /api/data | POST /api/echo |
| ------------ | ---------: | ------------: | -------------: |
| **pure_go**  | **42,290** |    **34,610** |     **36,940** |
| **pocketpy** | **39,880** |    **41,050** |     **40,044** |
| fastapi      |      7,643 |         7,657 |          7,606 |
| flask        |      7,229 |         7,697 |          7,871 |
| cpython_cffi |      4,487 |         4,474 |          4,543 |
| pure_python  |      4,574 |         4,652 |          4,480 |

Tested with `hey -n 5000 -c 10` on Apple M-series.

### Suite 2: TechEmpower-Inspired (Concurrency Sweep)

Plaintext, JSON, and SQLite I/O endpoints tested at concurrency levels 1, 5, 10, 25, 50, 100. SQLite uses per-server
WAL-mode databases with a 10,000-row `world` table.

#### Plaintext — `GET /plaintext` (req/s)

| Server         | c=1 |   c=5 |  c=10 |  c=25 |  c=50 |     c=100 |
| -------------- | --: | ----: | ----: | ----: | ----: | --------: |
| pocketpy       | 709 | 6,056 | 6,754 | 7,889 | 7,585 | **8,060** |
| fastapi        | 652 | 6,011 | 7,439 | 7,769 | 8,333 |     7,811 |
| pure_go        | 692 | 3,112 | 7,324 | 5,371 | 5,716 |     4,613 |
| flask          | 468 | 2,033 | 1,999 | 2,027 | 2,091 |     2,105 |
| cpython_gohttp | 700 | 3,588 | 3,604 | 3,156 |   626 |       421 |
| pure_python    | 525 | 3,521 | 3,452 | 2,152 |   626 |       362 |

#### JSON — `GET /json` (req/s)

| Server         | c=1 |   c=5 |  c=10 |  c=25 |  c=50 |     c=100 |
| -------------- | --: | ----: | ----: | ----: | ----: | --------: |
| fastapi        | 739 | 5,725 | 7,823 | 7,439 | 7,990 | **7,757** |
| pocketpy       | 703 | 6,095 | 7,492 | 8,027 | 7,859 |     7,566 |
| pure_go        | 713 | 5,652 | 7,199 | 7,766 | 7,705 |     6,900 |
| flask          | 474 | 1,903 | 1,973 | 2,041 | 1,982 |     2,004 |
| cpython_gohttp | 674 | 3,322 | 3,409 | 2,094 |   624 |       589 |
| pure_python    | 535 | 3,529 | 3,518 | 2,148 |   624 |       592 |

#### DB Single Query — `GET /db` (req/s)

| Server         | c=1 |   c=5 |  c=10 |  c=25 |  c=50 |     c=100 |
| -------------- | --: | ----: | ----: | ----: | ----: | --------: |
| fastapi        | 722 | 5,896 | 7,485 | 8,118 | 7,688 | **8,319** |
| pocketpy       | 721 | 6,082 | 7,294 | 7,637 | 7,578 |     8,157 |
| pure_go        | 701 | 5,726 | 7,405 | 7,909 | 7,766 |     4,916 |
| flask          | 472 | 1,198 | 1,057 | 1,254 | 1,264 |     1,371 |
| cpython_gohttp | 365 |   385 |   493 |   440 |   418 |       387 |
| pure_python    | 332 |   383 |   465 |   422 |   431 |       378 |

#### DB Multiple Queries — `GET /queries?N=20` (req/s)

| Server         | c=1 |   c=5 |  c=10 |  c=25 |  c=50 |     c=100 |
| -------------- | --: | ----: | ----: | ----: | ----: | --------: |
| pocketpy       | 704 | 6,009 | 6,961 | 7,615 | 8,018 | **8,089** |
| fastapi        | 681 | 6,025 | 7,560 | 7,832 | 7,810 |     8,021 |
| pure_go        | 690 | 5,769 | 6,997 | 8,152 | 8,038 |     7,823 |
| flask          | 467 |   444 |   534 |   639 |   640 |       632 |
| cpython_gohttp | 348 |   377 |   477 |   402 |   377 |       377 |
| pure_python    | 309 |   372 |   452 |   390 |   411 |       360 |

#### DB Updates — `POST /updates` (req/s)

| Server         | c=1 |   c=5 |  c=10 |  c=25 |  c=50 |     c=100 |
| -------------- | --: | ----: | ----: | ----: | ----: | --------: |
| pure_go        | 707 | 5,743 | 7,141 | 7,529 | 7,937 | **8,190** |
| pocketpy       | 696 | 6,001 | 7,293 | 8,019 | 8,341 |     7,912 |
| fastapi        | 664 | 5,816 | 7,770 | 8,075 | 8,105 |     7,887 |
| flask          | 463 |   575 |   635 |   637 |   665 |       826 |
| cpython_gohttp | 499 | 1,986 | 2,322 | 2,300 |   618 |       497 |
| pure_python    | 342 | 1,987 | 2,443 | 1,872 | 1,143 |       492 |

#### Memory Usage (RSS KB)

| Server         |    Idle | c=1 peak | c=100 peak | Growth at c=100 |
| -------------- | ------: | -------: | ---------: | --------------: |
| pure_go        | ~19,700 |   19,744 |     29,216 |          +9,472 |
| pure_python    | ~16,600 |   16,592 |     29,344 |         +12,752 |
| pocketpy       | ~23,000 |   23,024 |     35,056 |         +12,032 |
| cpython_gohttp | ~24,400 |   24,448 |     24,448 |               0 |
| flask          | ~29,100 |   27,472 |     29,664 |            +192 |
| fastapi        | ~66,700 |   66,656 |     68,176 |          +1,520 |

## Key Observations

- **pocketpy** leads on plaintext, queries, and matches fastapi on db/updates. Uses ~23 MB idle (Go runtime + embedded
  Python VM), growing to ~35 MB after sustained DB load. The Go HTTP layer handles connection queuing while the embedded
  Python interpreter serves requests efficiently.
- **fastapi** dominates db reads (8,319 req/s at c=100) and is competitive across all endpoints, but uses the most
  memory (~67 MB idle) due to uvicorn + SQLAlchemy + sqlmodel stack overhead.
- **pure_go** excels at writes (8,190 req/s at c=100) and is strong on db reads at mid-concurrency. Lowest idle memory
  (~20 MB) of the high-throughput servers, growing to ~29 MB after sustained DB load (SQLite page cache).
- **cpython_gohttp** demonstrates CPython + Go CFFI integration: Go JSON (gojson) handles serialization while Python's
  `http.server` handles HTTP. Peak plaintext throughput (3,604 req/s at c=10) matches `pure_python` but degrades sharply
  under high concurrency (421 req/s at c=100) due to GIL contention. Memory usage is flat at ~25 MB with zero growth
  under load — Python's threading model doesn't allocate per-connection state like Go's goroutines.
- **pure_python** collapses under concurrency — the GIL limits db throughput to ~400 req/s regardless of load level, and
  plaintext drops from 3,521 (c=5) to 362 (c=100). Memory grows significantly (+13 MB) from per-thread allocations.
- **flask** plateaus early (~2,000 req/s plaintext) and degrades on queries (632 req/s at c=100) due to synchronous WSGI
  blocking.
- **fastapi uses ~3x more memory** than pocketpy (67 MB vs 23 MB idle) for comparable throughput.
- **SQLite writes scale well** for Go-native and pocketpy (7,900–8,200 req/s) but pure_python and flask bottleneck on
  Python's sqlite3 module under concurrency.

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

| File                                          | Description                        |
| --------------------------------------------- | ---------------------------------- |
| `benchmarks/run.sh`                           | Concurrency sweep benchmark runner |
| `benchmarks/servers/server_pure_go.go`        | Go server with TFB endpoints       |
| `benchmarks/servers/server_pure.py`           | Python stdlib with TFB endpoints   |
| `benchmarks/servers/server_fastapi.py`        | FastAPI with TFB endpoints         |
| `benchmarks/servers/server_flask.py`          | Flask with TFB endpoints           |
| `benchmarks/servers/server_cpython_gohttp.py` | CPython using gojson for JSON      |
| `gofre/examples/webserver_binary/`            | Pocketpy Go+Python binary          |

## Running

### Prerequisites

```bash
go install github.com/rakyll/hey@latest
pip3 install fastapi uvicorn flask sqlmodel
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

# CPython + gojson
cd examples/webserver/benchmarks && python3 servers/server_cpython_gohttp.py 8087
```
