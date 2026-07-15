# TechEmpower-Inspired Benchmarks

Part of the [webserver benchmark suite](../README.md). Runs plaintext, JSON, and SQLite I/O endpoints at varying
concurrency levels (1, 5, 10, 25, 50, 100) to measure throughput scaling and memory usage.

## Endpoints

| Endpoint        | Method | Description                      |
| --------------- | ------ | -------------------------------- |
| `/plaintext`    | GET    | Plain text response              |
| `/json`         | GET    | JSON serialization               |
| `/db`           | GET    | Single random row from SQLite    |
| `/queries?N=20` | GET    | Multiple random rows from SQLite |
| `/updates`      | POST   | Update random rows in SQLite     |

Each server gets its own WAL-mode SQLite database with a 10,000-row `world` table.

## Quick Start

```bash
cd examples/webserver/benchmarks
bash run.sh
```

Results are saved to `results/results.json`.

## Customizing

Edit `CONCURRENCY_LEVELS` in `run.sh` to test different concurrency profiles:

```bash
CONCURRENCY_LEVELS=(1 5 10 25 50 100)  # default
CONCURRENCY_LEVELS=(1 10 50 100)        # fewer levels
```
