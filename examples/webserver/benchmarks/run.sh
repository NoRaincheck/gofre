#!/usr/bin/env bash
# Benchmark runner — plaintext & json endpoints with concurrency sweep.
# Compatible with macOS bash 3.2.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVERS_DIR="$SCRIPT_DIR/servers"
RESULTS_DIR="$SCRIPT_DIR/results"
mkdir -p "$RESULTS_DIR"

NUM_REQUESTS=5000
CONCURRENCY_LEVELS=(1 5 10 25 50 100)

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BOLD='\033[1m'; CYAN='\033[0;36m'; NC='\033[0m'

log()  { echo -e "${BOLD}[INFO]${NC} $*"; }
ok()   { echo -e "${GREEN}[OK]${NC} $*"; }
fail() { echo -e "${RED}[FAIL]${NC} $*"; }

GO_FRE_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)/gofre"
POCKETPY_BINARY="$GO_FRE_ROOT/examples/webserver_binary/webserver_binary"

SERVERS_FILE="$RESULTS_DIR/.servers"
TESTS_FILE="$RESULTS_DIR/.tests"

start_server() {
    local name=$1 port=$2 cmd=$3
    local pid
    pid=$(lsof -ti :"$port" 2>/dev/null || true)
    if [ -n "$pid" ]; then
        kill -9 "$pid" 2>/dev/null || true
        sleep 0.3
    fi

    if [ "$name" = "pocketpy" ]; then
        rm -f "$GO_FRE_ROOT/examples/webserver_binary/benchmark.db"
        (cd "$GO_FRE_ROOT/examples/webserver_binary" && "$POCKETPY_BINARY" "$port") \
            > "$RESULTS_DIR/${name}.log" 2>&1 &
    else
        eval "$cmd" > "$RESULTS_DIR/${name}.log" 2>&1 &
    fi
    local server_pid=$!
    # Wait briefly for the process to fork, then use lsof to get the actual PID
    sleep 0.3
    local actual_pid
    actual_pid=$(lsof -ti :"$port" 2>/dev/null | head -1)
    if [ -n "$actual_pid" ]; then
        server_pid=$actual_pid
    fi
    echo "$server_pid" > "$RESULTS_DIR/${name}.pid"

    local retries=30
    while [ $retries -gt 0 ]; do
        if curl -s "http://localhost:$port/plaintext" > /dev/null 2>&1; then
            echo -e "${GREEN}[OK]${NC} Server '$name' ready on port $port (PID: $server_pid)" >&2
            echo "$server_pid"
            return 0
        fi
        retries=$((retries - 1))
        sleep 0.5
    done
    echo -e "${RED}[FAIL]${NC} Server '$name' failed to start on port $port" >&2
    cat "$RESULTS_DIR/${name}.log" >&2
    return 1
}

stop_server() {
    local pid=$1
    if kill -0 "$pid" 2>/dev/null; then
        kill "$pid" 2>/dev/null || true
        wait "$pid" 2>/dev/null || true
        echo -e "${GREEN}[OK]${NC} Server stopped (PID: $pid)" >&2
    fi
}

get_rss() {
    local pid=$1
    local val
    if [ "$(uname)" = "Darwin" ]; then
        val=$(ps -o rss= -p "$pid" 2>/dev/null | tr -d ' ')
    else
        val=$(awk '/^VmRSS/ {print $2}' "/proc/$pid/status" 2>/dev/null)
    fi
    if [ -z "$val" ]; then
        echo "0"
    else
        echo "$val"
    fi
}

format_mb() {
    local kb=$1
    if [ "$kb" -eq 0 ] 2>/dev/null; then
        echo "N/A"
    else
        echo "scale=1; $kb / 1024" | bc 2>/dev/null || echo "$((kb / 1024)).0"
    fi
}

# Background memory sampler — writes peak RSS to a temp file.
# Usage: sample_mem_start <pid> <tempfile>
sample_mem_start() {
    local pid=$1 tmpfile=$2
    (
        local peak=0
        local now
        while [ -f "$tmpfile.running" ]; do
            now=$(get_rss "$pid")
            if [ "$now" -gt "$peak" ] 2>/dev/null; then
                peak=$now
            fi
            sleep 0.05
        done
        echo "$peak" > "$tmpfile.peak"
    ) &
    touch "$tmpfile.running"
    echo $!
}

sample_mem_stop() {
    local sampler_pid=$1
    if [ -n "$sampler_pid" ] && kill -0 "$sampler_pid" 2>/dev/null; then
        kill "$sampler_pid" 2>/dev/null || true
        wait "$sampler_pid" 2>/dev/null || true
    fi
}

main() {
    # Config: name<TAB>port<TAB>cmd
    printf "pure_python\t8082\tpython3 %s/server_pure.py\n" "$SERVERS_DIR" > "$SERVERS_FILE"
    printf "fastapi\t8083\tpython3 %s/server_fastapi.py\n" "$SERVERS_DIR" >> "$SERVERS_FILE"
    printf "flask\t8084\tpython3 %s/server_flask.py\n" "$SERVERS_DIR" >> "$SERVERS_FILE"
    printf "pure_go\t8085\tgo run %s/server_pure_go.go\n" "$SERVERS_DIR" >> "$SERVERS_FILE"
    printf "pocketpy\t8086\t%s\n" "$POCKETPY_BINARY" >> "$SERVERS_FILE"

    printf "plaintext\tGET\t/plaintext\t0\n" > "$TESTS_FILE"
    printf "json\tGET\t/json\t0\n" >> "$TESTS_FILE"

    log "=========================================="
    log "  Webserver Benchmark Suite"
    log "  Plaintext + JSON — Concurrency Sweep"
    log "  $(date)"
    log "=========================================="
    log "Requests per test: $NUM_REQUESTS"
    log "Concurrency levels: ${CONCURRENCY_LEVELS[*]}"
    log ""

    HEY_BIN=""
    if command -v hey &> /dev/null; then
        HEY_BIN="hey"
    elif [ -f "$(go env GOPATH)/bin/hey" ]; then
        HEY_BIN="$(go env GOPATH)/bin/hey"
    else
        fail "hey not found. Install: go install github.com/rakyll/hey@latest"
        exit 1
    fi

    # Clean previous results
    rm -f "$RESULTS_DIR"/results_* "$RESULTS_DIR"/.all_results.txt "$RESULTS_DIR"/*.log "$RESULTS_DIR"/*.pid "$RESULTS_DIR"/.pids "$RESULTS_DIR"/.mem_*

    # Start servers
    log "Starting servers..."
    SERVER_COUNT=0
    while IFS=$'\t' read -r name port cmd; do
        pid=$(start_server "$name" "$port" "$cmd" 2>&1 || true)
        if [ -n "$pid" ] && echo "$pid" | grep -q '^[0-9]'; then
            echo "$pid" >> "$RESULTS_DIR/.pids"
            SERVER_COUNT=$((SERVER_COUNT + 1))
        else
            echo "" >&2
            echo -e "${YELLOW}[SKIP]${NC} Server '$name' failed to start, skipping" >&2
        fi
    done < "$SERVERS_FILE"
    echo ""

    # Master results file: server|test|concurrency|throughput|mem_before|mem_peak|mem_after|mem_growth
    RESULTS_FILE="$RESULTS_DIR/.all_results.txt"
    > "$RESULTS_FILE"

    # Run benchmarks for each concurrency level
    for CONCURRENCY in "${CONCURRENCY_LEVELS[@]}"; do
        log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        log "  Concurrency: $CONCURRENCY"
        log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""

        # Header
        printf "${BOLD}%-18s" "Server"
        while IFS=$'\t' read -r tname tmethod tpath text; do
            printf "%-14s" "$tname"
        done < "$TESTS_FILE"
        printf "%-10s\n" "Idle (KB)"
        printf "%-18s" "------------------"
        while IFS=$'\t' read -r tname tmethod tpath text; do
            printf "%-14s" "--------------"
        done < "$TESTS_FILE"
        printf "%-10s\n" "----------"

        while IFS=$'\t' read -r name port cmd; do
            # Skip if server didn't start
            if [ ! -f "$RESULTS_DIR/${name}.pid" ]; then
                printf "${CYAN}%-18s" "$name"
                while IFS=$'\t' read -r tname tmethod tpath text; do
                    printf "%-14s" "SKIPPED"
                done < "$TESTS_FILE"
                printf "%-10s\n" "N/A"
                continue
            fi

            printf "${CYAN}%-18s" "$name"

            while IFS=$'\t' read -r tname tmethod tpath text; do
                hey_args=("-n" "$NUM_REQUESTS" "-c" "$CONCURRENCY")
                url="http://localhost:$port$tpath"
                if [ "$text" != "0" ]; then
                    url="$url?$text"
                fi

                pid=$(cat "$RESULTS_DIR/${name}.pid" 2>/dev/null || echo "0")
                mem_before=$(get_rss "$pid")

                # Start background memory sampler DURING the load test
                mem_tmp="$RESULTS_DIR/.mem_${CONCURRENCY}_${name}_${tname}"
                > "$mem_tmp.peak"
                sampler_pid=$(sample_mem_start "$pid" "$mem_tmp")

                hey_output=$("$HEY_BIN" "${hey_args[@]}" "$url" 2>&1)

                # Stop sampler after hey completes
                sample_mem_stop "$sampler_pid"

                throughput=$(echo "$hey_output" | grep 'Requests/sec' | awk '{print $2}' | head -1 || echo "0")
                if [ -z "$throughput" ]; then
                    throughput="0"
                fi

                mem_peak=$(cat "$mem_tmp.peak" 2>/dev/null || echo "$mem_before")
                if [ -z "$mem_peak" ] || [ "$mem_peak" -eq 0 ] 2>/dev/null; then
                    mem_peak="$mem_before"
                fi

                mem_after=$(get_rss "$pid")
                mem_growth=$((mem_after - mem_before))

                echo "${name}|${tname}|${CONCURRENCY}|${throughput}|${mem_before}|${mem_peak}|${mem_after}|${mem_growth}" >> "$RESULTS_FILE"
                printf "%-14s" "${throughput} req/s"
            done < "$TESTS_FILE"

            mem_before=$(get_rss "$pid")
            printf "%-10s\n" "${mem_before} KB"
        done < "$SERVERS_FILE"

        echo ""
    done

    # ── Summary: Throughput by concurrency ─────────────────────────
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log "  THROUGHPUT SUMMARY (req/s)"
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    while IFS=$'\t' read -r tname tmethod tpath text; do
        printf "${BOLD}%-18s" "$tname"
        for c in "${CONCURRENCY_LEVELS[@]}"; do
            printf "%-12s" "c=$c"
        done
        echo ""
        printf "%-18s" "------------------"
        for c in "${CONCURRENCY_LEVELS[@]}"; do
            printf "%-12s" "------------"
        done
        echo ""

        while IFS=$'\t' read -r sname sport scmd; do
            if [ ! -f "$RESULTS_DIR/${sname}.pid" ]; then
                continue
            fi
            printf "${CYAN}%-18s" "$sname"
            for c in "${CONCURRENCY_LEVELS[@]}"; do
                line=$(grep "^${sname}|${tname}|${c}|" "$RESULTS_FILE" 2>/dev/null || echo "")
                if [ -n "$line" ]; then
                    tp=$(echo "$line" | cut -d'|' -f4)
                    printf "%-12s" "${tp}"
                else
                    printf "%-12s" "—"
                fi
            done
            echo ""
        done < "$SERVERS_FILE"
        echo ""
    done < "$TESTS_FILE"

    # ── Summary: Memory by concurrency ────────────────────────────
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log "  MEMORY SUMMARY (RSS KB: idle → peak → post, growth)"
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    while IFS=$'\t' read -r tname tmethod tpath text; do
        printf "${BOLD}%-18s  %-14s\n" "$tname" ""
        printf "${BOLD}%-18s" "Server"
        for c in "${CONCURRENCY_LEVELS[@]}"; do
            printf "%-20s" "c=$c"
        done
        echo ""
        printf "%-18s" "------------------"
        for c in "${CONCURRENCY_LEVELS[@]}"; do
            printf "%-20s" "--------------------"
        done
        echo ""

        while IFS=$'\t' read -r sname sport scmd; do
            if [ ! -f "$RESULTS_DIR/${sname}.pid" ]; then
                continue
            fi
            printf "${CYAN}%-18s" "$sname"
            for c in "${CONCURRENCY_LEVELS[@]}"; do
                line=$(grep "^${sname}|${tname}|${c}|" "$RESULTS_FILE" 2>/dev/null || echo "")
                if [ -n "$line" ]; then
                    idle=$(echo "$line" | cut -d'|' -f5)
                    peak=$(echo "$line" | cut -d'|' -f6)
                    growth=$(echo "$line" | cut -d'|' -f8)
                    if [ "$growth" -lt 0 ] 2>/dev/null; then
                        growth_str="-${growth}"
                    else
                        growth_str="+${growth}"
                    fi
                    printf "%-20s" "${idle}→${peak} ${growth_str}"
                else
                    printf "%-20s" "—"
                fi
            done
            echo ""
        done < "$SERVERS_FILE"
        echo ""
    done < "$TESTS_FILE"

    # ── JSON output ───────────────────────────────────────────────
    log "Saving results to $RESULTS_DIR/results.json..."
    {
        echo "{"
        echo "  \"timestamp\": \"$(date -Iseconds)\","
        echo "  \"parameters\": {"
        echo "    \"requests\": $NUM_REQUESTS,"
        echo "    \"concurrency_levels\": [$(printf '%s,' "${CONCURRENCY_LEVELS[@]}" | sed 's/,$//')]"
        echo "  },"
        echo "  \"tests\": ["

        first_test=true
        while IFS=$'\t' read -r tname tmethod tpath text; do
            if [ "$first_test" = true ]; then
                first_test=false
            else
                echo ","
            fi
            printf "    {\n      \"name\": \"%s\",\n      \"results\": [" "$tname"

            first_server=true
            while IFS=$'\t' read -r sname sport scmd; do
                if [ ! -f "$RESULTS_DIR/${sname}.pid" ]; then
                    continue
                fi

                if [ "$first_server" = true ]; then
                    first_server=false
                else
                    printf ","
                fi
                printf "\n        {\"server\": \"%s\", \"concurrency\": [" "$sname"

                first_c=true
                for c in "${CONCURRENCY_LEVELS[@]}"; do
                    line=$(grep "^${sname}|${tname}|${c}|" "$RESULTS_FILE" 2>/dev/null || echo "")
                    if [ -z "$line" ]; then
                        continue
                    fi
                    tp=$(echo "$line" | cut -d'|' -f4)
                    idle=$(echo "$line" | cut -d'|' -f5)
                    peak=$(echo "$line" | cut -d'|' -f6)
                    growth=$(echo "$line" | cut -d'|' -f8)

                    if [ "$first_c" = true ]; then
                        first_c=false
                    else
                        printf ","
                    fi
                    printf "\n          {\"level\": %s, \"throughput\": %s, \"memory\": {\"idle_kb\": %s, \"peak_kb\": %s, \"growth_kb\": %s}}" "$c" "$tp" "$idle" "$peak" "$growth"
                done
                printf "\n        ]}"
            done < "$SERVERS_FILE"
            printf "\n      ]\n    }"
        done < "$TESTS_FILE"
        echo ""
        echo "  ]"
        echo "}"
    } > "$RESULTS_DIR/results.json"

    # Cleanup
    log "Stopping servers..."
    if [ -f "$RESULTS_DIR/.pids" ]; then
        while read -r pid; do
            stop_server "$pid"
        done < "$RESULTS_DIR/.pids"
    fi

    log ""
    log "=========================================="
    log "  Benchmarks complete!"
    log "  Results: $RESULTS_DIR/"
    log "=========================================="
}

main "$@"
