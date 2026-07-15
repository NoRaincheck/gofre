# TechEmpower-Inspired Benchmarks

Part of the [webserver benchmark suite](../README.md). Runs plaintext and JSON endpoints at varying concurrency levels
(1, 5, 10, 25, 50, 100) to measure throughput scaling and memory usage.

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
