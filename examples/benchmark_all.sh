#!/bin/bash
set -e

REQUESTS=5000
CONCURRENCY=10

export PATH=$PATH:~/go/bin
cd "$(dirname "$0")"

echo "================================================"
echo "Full Webserver Benchmark: pocketpy vs Python frameworks"
echo "  Requests: $REQUESTS per test"
echo "  Concurrency: $CONCURRENCY"
echo "================================================"
echo ""

bench() {
    local name=$1
    local port=$2
    local cmd=$3
    local wait=${4:-2}

    echo "--- $name ---"
    eval "$cmd" > /dev/null 2>&1 &
    local pid=$!
    sleep $wait

    hey -n $REQUESTS -c $CONCURRENCY http://localhost:$port/ 2>&1 | grep -E "Requests/sec:|Average|Slowest"
    
    hey -n $REQUESTS -c $CONCURRENCY http://localhost:$port/api/data 2>&1 | grep "Requests/sec:"

    hey -n $REQUESTS -c $CONCURRENCY -m POST -d '{"test":"data","n":42}' http://localhost:$port/api/echo 2>&1 | grep "Requests/sec:"

    kill $pid 2>/dev/null || true
    wait $pid 2>/dev/null || true
    echo ""
}

# 1. Go binary (pocketpy embedded)
bench "Go Binary (pocketpy embedded)" 8080 \
  "./goforge/examples/webserver_binary/webserver_binary"

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

echo "================================================"
echo "Benchmark complete!"
echo "================================================"
