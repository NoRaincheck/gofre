#!/bin/bash
# Benchmark: pocketpy binary vs CPython+cffi vs pure Python

set -e

GO_BINARY="goforge/examples/webserver_binary/webserver_binary"
CPYTHON_SCRIPT="examples/webserver/server.py"
PURE_SCRIPT="examples/webserver/server_pure.py"
REQUESTS=${1:-1000}
CONCURRENCY=${2:-10}

echo "========================================"
echo "Webserver Benchmark"
echo "  Requests: $REQUESTS (per test)"
echo "  Concurrency: $CONCURRENCY"
echo "========================================"
echo ""

# Binary sizes
echo "--- Binary Sizes ---"
if [ -f "$GO_BINARY" ]; then
    ls -lh "$GO_BINARY" 2>/dev/null | awk '{print "  Go binary (pocketpy):", $5}'
else
    echo "  Go binary (pocketpy): not built (build with no_pocketpy tag excluded)"
fi
python3 -c "
import sys, os
libs = ['examples/webserver/build/goforge_webserver/_binding.dylib',
        'examples/webserver/build/goforge_webserver/_binding.so']
for lib in libs:
    if os.path.exists(lib):
        size = os.path.getsize(lib)
        print(f'  Go shared lib (cffi): {size/1024:.0f} KB')
"
echo ""

bench() {
    local name=$1
    local port=$2
    local cmd=$3

    echo "--- $name ---"

    # Start server
    eval "$cmd" > /dev/null 2>&1 &
    local pid=$!
    sleep 1

    # Warmup
    curl -s http://localhost:$port/ > /dev/null 2>&1 || true

    # Benchmark with hey or wrk or ab
    if command -v hey &> /dev/null; then
        result=$(hey -n $REQUESTS -c $CONCURRENCY http://localhost:$port/ 2>&1)
        echo "$result" | grep -E "Requests|Total:|Slowest|Fastest|Average|Requests/sec"
    elif command -v wrk &> /dev/null; then
        wrk -t$CONCURRENCY -c$CONCURRENCY -d5s http://localhost:$port/ 2>&1
    else
        # Simple curl-based timing
        start=$(date +%s.%N)
        for i in $(seq 1 $REQUESTS); do
            curl -s http://localhost:$port/ > /dev/null 2>&1
        done
        end=$(date +%s.%N)
        elapsed=$(echo "$end - $start" | bc)
        rps=$(echo "$REQUESTS / $elapsed" | bc -l)
        printf "  Requests: %d in %.2fs (%.0f req/s)\n" $REQUESTS $elapsed $rps
    fi

    # Test /api/data
    if command -v hey &> /dev/null; then
        result=$(hey -n $REQUESTS -c $CONCURRENCY http://localhost:$port/api/data 2>&1)
        echo "$result" | grep -E "Requests/sec"
    fi

    # Test POST /api/echo
    if command -v hey &> /dev/null; then
        result=$(hey -n $REQUESTS -c $CONCURRENCY -m POST -d '{"test":"data"}' http://localhost:$port/api/echo 2>&1)
        echo "$result" | grep -E "Requests/sec"
    fi

    kill $pid 2>/dev/null || true
    wait $pid 2>/dev/null || true
    echo ""
}

# Test 1: Go binary (pocketpy) - skip if not built
if [ -f "$GO_BINARY" ]; then
    bench "Go Binary (pocketpy embedded)" 8080 "$GO_BINARY"
else
    echo "--- Go Binary (pocketpy embedded) ---"
    echo "  SKIPPED: binary not found at $GO_BINARY"
    echo "  Build without no_pocketpy tag to include this benchmark"
    echo ""
fi

# Test 2: CPython + Go cffi
bench "CPython + Go cffi" 8081 "python3 $CPYTHON_SCRIPT 8081"

# Test 3: Pure Python baseline
bench "Pure Python (baseline)" 8082 "python3 $PURE_SCRIPT 8082"

echo "========================================"
echo "Benchmark complete!"
echo "========================================"
