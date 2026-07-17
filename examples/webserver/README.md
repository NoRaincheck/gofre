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
| `cpython_gohttp` | Python 3 + Go cffi  | 8087 | CPython with Go HTTP server (gohttp) via cffi — Python handles business logic only   |

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

| Server         |        c=1 |        c=5 |       c=10 |       c=25 |       c=50 |      c=100 |
| -------------- | ---------: | ---------: | ---------: | ---------: | ---------: | ---------: |
| **pure_go**    | **12,667** | **36,118** | **44,093** | **50,499** | **55,805** | **61,481** |
| cpython_gohttp |     11,774 |     39,152 |     44,493 |     46,441 |     47,113 |     51,008 |
| **pocketpy**   | **11,980** | **15,718** | **24,788** | **31,140** | **29,779** | **34,379** |
| fastapi        |        652 |      6,011 |      7,439 |      7,769 |      8,333 |      7,811 |
| flask          |        468 |      2,033 |      1,999 |      2,027 |      2,091 |      2,105 |
| pure_python    |        525 |      3,521 |      3,452 |      2,152 |        626 |        362 |

#### JSON — `GET /json` (req/s)

| Server         |        c=1 |        c=5 |       c=10 |       c=25 |       c=50 |      c=100 |
| -------------- | ---------: | ---------: | ---------: | ---------: | ---------: | ---------: |
| **pure_go**    | **12,497** | **36,054** | **48,516** | **50,573** | **53,234** | **59,035** |
| cpython_gohttp |     11,807 |     34,100 |     34,703 |     36,759 |     35,527 |     37,445 |
| **pocketpy**   | **11,357** | **16,738** | **24,781** | **30,517** | **30,532** | **33,282** |
| fastapi        |        739 |      5,725 |      7,823 |      7,439 |      7,990 |      7,757 |
| flask          |        474 |      1,903 |      1,973 |      2,041 |      1,982 |      2,004 |
| pure_python    |        535 |      3,529 |      3,518 |      2,148 |        624 |        592 |

#### DB Single Query — `GET /db` (req/s)

| Server         |     c=1 |        c=5 |       c=10 |       c=25 |       c=50 |      c=100 |
| -------------- | ------: | ---------: | ---------: | ---------: | ---------: | ---------: |
| **pocketpy**   | **256** | **15,234** | **24,943** | **30,557** | **29,663** | **31,640** |
| fastapi        |     722 |      5,896 |      7,485 |      8,118 |      7,688 |      8,319 |
| cpython_gohttp |     766 |        337 |        431 |        398 |        403 |        402 |
| pure_go        |     253 |        271 |        271 |        267 |        268 |        269 |
| flask          |     472 |      1,198 |      1,057 |      1,254 |      1,264 |      1,371 |
| pure_python    |     332 |        383 |        465 |        422 |        431 |        378 |

#### DB Multiple Queries — `GET /queries?N=20` (req/s)

| Server         |       c=1 |        c=5 |       c=10 |       c=25 |       c=50 |      c=100 |
| -------------- | --------: | ---------: | ---------: | ---------: | ---------: | ---------: |
| **pocketpy**   | **8,286** | **15,086** | **19,918** | **21,186** | **26,967** | **30,335** |
| fastapi        |       681 |      6,025 |      7,560 |      7,832 |      7,810 |      8,021 |
| cpython_gohttp |       724 |        386 |        457 |        412 |        421 |        393 |
| pure_go        |       248 |        262 |        261 |        263 |        260 |        260 |
| flask          |       467 |        444 |        534 |        639 |        640 |        632 |
| pure_python    |       309 |        372 |        452 |        390 |        411 |        360 |

#### DB Updates — `POST /updates` (req/s)

| Server         |       c=1 |        c=5 |       c=10 |       c=25 |       c=50 |      c=100 |
| -------------- | --------: | ---------: | ---------: | ---------: | ---------: | ---------: |
| **pure_go**    | **8,981** | **23,431** | **22,738** | **23,301** | **23,423** | **23,609** |
| **pocketpy**   | **4,521** | **16,696** | **24,819** | **30,681** | **21,386** | **32,514** |
| fastapi        |       664 |      5,816 |      7,770 |      8,075 |      8,105 |      7,887 |
| cpython_gohttp |     1,876 |      8,136 |      8,406 |      7,298 |      7,813 |      7,218 |
| flask          |       463 |        575 |        635 |        637 |        665 |        826 |
| pure_python    |       342 |      1,987 |      2,443 |      1,872 |      1,143 |        492 |

#### Memory Usage (RSS KB)

| Server         |        Idle |   c=1 peak |  c=100 peak | Growth at c=100 |
| -------------- | ----------: | ---------: | ----------: | --------------: |
| pure_go        |     ~19,700 |     19,744 |      29,216 |          +9,472 |
| pure_python    |     ~16,600 |     16,592 |      29,344 |         +12,752 |
| **pocketpy**   | **~23,000** | **22,656** | **~33,000** |     **+10,000** |
| cpython_gohttp |     ~24,400 |     30,848 |      47,120 |         +22,720 |
| flask          |     ~29,100 |     27,472 |      29,664 |            +192 |
| fastapi        |     ~66,700 |     66,656 |      68,176 |          +1,520 |

## Key Observations

- **pure_go** dominates plaintext and JSON throughput (50–61k req/s at high concurrency) — Go's `net/http` with no
  dependencies is the fastest for CPU-bound endpoints. DB read throughput is capped at ~270 req/s due to
  `ORDER BY RANDOM()` on 10k rows serialized through a single SQLite connection (`MaxOpenConns(1)`). DB updates reach
  23k req/s at concurrency — simple indexed UPDATE statements are much cheaper than random reads.
- **pocketpy** reaches 30–34k req/s across all endpoints at high concurrency, including DB reads and writes. Its
  Go-native SQLite drives DB operations with no Python sqlite3 bottleneck. Interestingly, pocketpy's DB read throughput
  at concurrency (15–31k) is significantly higher than pure_go's (~270), despite both using `MaxOpenConns(1)` with
  `modernc.org/sqlite`. This suggests the pocketpy Go SQL bridge may be handling connection pooling differently at the
  HTTP dispatch level. ~23 MB idle, growing to ~33 MB under load.
- **cpython_gohttp** demonstrates CPython + Go CFFI integration: Go HTTP (gohttp) handles connection management while
  Python's business logic is called via CFFI dispatch. Achieves ~80% of pure_go on plaintext (51k) and JSON (37k).
  DB/queries match pure_python (~400 req/s) since both use Python's sqlite3. Updates reach ~8k req/s — Go goroutines
  serialize writes through CFFI, avoiding GIL contention. Memory grows +23 MB under load from goroutine-per-connection
  allocation.
- **fastapi** provides consistent throughput (~6–8k req/s) with async ASGI. DB/queries benefit from SQLAlchemy
  connection pooling. Uses the most memory (~67 MB idle) due to uvicorn + SQLAlchemy + sqlmodel stack overhead.
- **pure_python** collapses under concurrency — the GIL limits db throughput to ~400 req/s regardless of load level, and
  plaintext drops from 3,521 (c=5) to 362 (c=100). Memory grows significantly (+13 MB) from per-thread allocations.
- **flask** plateaus early (~2,000 req/s plaintext) and degrades on queries (632 req/s at c=100) due to synchronous WSGI
  blocking.
- **fastapi uses ~3x more memory** than pocketpy (67 MB vs 23 MB idle) for comparable throughput.
- **SQLite reads vs writes**: pure_go's updates (23k) are 86x faster than its reads (~270) at concurrency, showing that
  `ORDER BY RANDOM()` on 10k rows is the real bottleneck, not SQLite itself. pocketpy avoids this gap by handling DB
  operations at the Go level before the Python dispatch layer.

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

| File                                          | Description                          |
| --------------------------------------------- | ------------------------------------ |
| `benchmarks/run.sh`                           | Concurrency sweep benchmark runner   |
| `benchmarks/servers/server_pure_go.go`        | Go server with TFB endpoints         |
| `benchmarks/servers/server_pure.py`           | Python stdlib with TFB endpoints     |
| `benchmarks/servers/server_fastapi.py`        | FastAPI with TFB endpoints           |
| `benchmarks/servers/server_flask.py`          | Flask with TFB endpoints             |
| `benchmarks/servers/server_cpython_gohttp.py` | CPython with Go HTTP server via cffi |
| `gofre/examples/webserver_binary/`            | Pocketpy Go+Python binary            |

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

Run specific servers only (no Python deps required for pocketpy/pure_go):

```bash
bash run.sh --servers pocketpy
bash run.sh --servers pocketpy,pure_go
```

To customize concurrency levels:

```bash
bash run.sh --concurrency 1,10,50,100
```

To change request count:

```bash
bash run.sh --requests 10000
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

# CPython + gohttp
cd examples/webserver/benchmarks && python3 servers/server_cpython_gohttp.py 8087
```
