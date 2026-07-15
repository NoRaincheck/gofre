#!/bin/bash
set -e

REQUESTS=5000
CONCURRENCY=10

export PATH=$PATH:~/go/bin
cd "$(dirname "$0")/.."

# Arrays to collect memory results
declare -a MEM_NAMES=()
declare -a MEM_IDLE_KB=()
declare -a MEM_PEAK_KB=()

sample_rss() {
    local pid=$1
    local outfile=$2
    while kill -0 "$pid" 2>/dev/null; do
        ps -o rss= -p "$pid" 2>/dev/null | tr -d ' ' >> "$outfile"
        sleep 0.1
    done
}

bench() {
    local name=$1
    local port=$2
    local cmd=$3
    local wait=${4:-2}

    echo "--- $name ---"

    # Extract server name pattern for process identification
    local server_pattern
    server_pattern=$(echo "$cmd" | grep -oE '[a-zA-Z0-9_-]+\.(py|go)$|[^/]+$' | tail -1)

    # Start server in background. Use a subshell with exec so the shell replaces itself,
    # making $! point to the actual server process.
    eval "$cmd" > /dev/null 2>&1 &
    local shell_pid=$!

    # Wait for server to be ready
    sleep $wait

    # Try to find the actual server process.
    # Strategy: find all processes matching the server pattern, exclude benchmark script,
    # then pick the one with highest RSS (actual server, not shell wrapper).
    local pid=""
    if [ -n "$server_pattern" ]; then
        pid=$(pgrep -f "$server_pattern" | grep -v "benchmark_all" | while read -r p; do
            rss=$(ps -o rss= -p "$p" 2>/dev/null | tr -d ' ')
            [ -n "$rss" ] && echo "$rss $p"
        done | sort -rn | head -1 | awk '{print $2}')
    fi

    # Fallback: if pattern matching failed, try the shell_pid directly
    if [ -z "$pid" ] || ! kill -0 "$pid" 2>/dev/null; then
        if kill -0 "$shell_pid" 2>/dev/null; then
            pid=$shell_pid
        fi
    fi

    if [ -z "$pid" ] || ! kill -0 "$pid" 2>/dev/null; then
        echo "  WARNING: Could not find server process for $name"
        echo "  Idle RSS: ? KB | Peak RSS: ? KB"
        echo ""
        MEM_NAMES+=("$name")
        MEM_IDLE_KB+=("0")
        MEM_PEAK_KB+=("0")
        return
    fi

    # Measure idle RSS (single sample after startup)
    local idle_rss
    idle_rss=$(ps -o rss= -p "$pid" 2>/dev/null | tr -d ' ')

    # Sample RSS in background during benchmark
    local rss_file
    rss_file=$(mktemp)
    sample_rss "$pid" "$rss_file" &
    local sampler_pid=$!

    hey -n $REQUESTS -c $CONCURRENCY http://localhost:$port/ 2>&1 | grep -E "Requests/sec:|Average|Slowest"

    hey -n $REQUESTS -c $CONCURRENCY http://localhost:$port/api/data 2>&1 | grep "Requests/sec:"

    hey -n $REQUESTS -c $CONCURRENCY -m POST -d '{"test":"data","n":42}' http://localhost:$port/api/echo 2>&1 | grep "Requests/sec:"

    # Kill the actual server process and its children
    kill $pid 2>/dev/null || true
    kill $(pgrep -P "$pid" 2>/dev/null) 2>/dev/null || true
    wait $pid 2>/dev/null || true

    # Stop sampler and compute peak RSS
    kill $sampler_pid 2>/dev/null || true
    wait $sampler_pid 2>/dev/null || true
    local peak_rss=0
    if [ -s "$rss_file" ]; then
        peak_rss=$(sort -n "$rss_file" | tail -1)
    fi
    rm -f "$rss_file"

    # Store results
    MEM_NAMES+=("$name")
    MEM_IDLE_KB+=("${idle_rss:-0}")
    MEM_PEAK_KB+=("${peak_rss:-0}")

    printf "  Idle RSS: %s KB | Peak RSS: %s KB\n" "${idle_rss:-?}" "${peak_rss:-?}"
    echo ""
}

# 1. Go binary (pocketpy embedded) - skip if not built
if [ -f "./gofre/examples/webserver_binary/webserver_binary" ]; then
    bench "Go Binary (pocketpy embedded)" 8080 \
      "./gofre/examples/webserver_binary/webserver_binary"
else
    echo "--- Go Binary (pocketpy embedded) ---"
    echo "  SKIPPED: binary not found"
    echo "  Build without no_pocketpy tag to include this benchmark"
    echo ""
fi

# 2. FastAPI + uvicorn (ASGI)
bench "FastAPI + uvicorn" 8083 \
  "python3 examples/webserver/server_fastapi.py 8083" 3

# 3. Flask (WSGI, dev server)
bench "Flask (dev server)" 8084 \
  "python3 examples/webserver/server_flask.py 8084" 2

# 4. CPython + Go cffi
bench "CPython + Go cffi (http.server)" 8081 \
  "python3 examples/webserver/server.py 8081"

# 5. Pure Python (baseline)
bench "Pure Python (baseline)" 8082 \
  "python3 examples/webserver/server_pure.py 8082"

# 6. Pure Go (stdlib only)
if [ -f "./examples/webserver/server_pure_go" ]; then
    bench "Pure Go (stdlib)" 8085 \
      "./examples/webserver/server_pure_go 8085"
else
    echo "--- Pure Go (stdlib) ---"
    echo "  SKIPPED: binary not found"
    echo "  Build with: cd examples/webserver && go build -o server_pure_go server_pure_go.go"
    echo ""
    MEM_NAMES+=("Pure Go (stdlib)")
    MEM_IDLE_KB+=("0")
    MEM_PEAK_KB+=("0")
fi

echo "================================================"
echo "Memory Usage Summary"
echo "================================================"
printf "%-30s %12s %12s\n" "Server" "Idle RSS" "Peak RSS"
printf "%-30s %12s %12s\n" "------------------------------" "------------" "------------"
for i in "${!MEM_NAMES[@]}"; do
    idle_mb=$(echo "scale=1; ${MEM_IDLE_KB[$i]} / 1024" | bc 2>/dev/null || echo "?")
    peak_mb=$(echo "scale=1; ${MEM_PEAK_KB[$i]} / 1024" | bc 2>/dev/null || echo "?")
    printf "%-30s %10s MB %10s MB\n" "${MEM_NAMES[$i]}" "$idle_mb" "$peak_mb"
done
echo ""
echo "================================================"
echo "Benchmark complete!"
echo "================================================"
