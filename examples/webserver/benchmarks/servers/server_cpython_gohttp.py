"""CPython webserver using Go HTTP (gohttp) via cffi for serving.

Routes:
  GET  /plaintext   — Plain text response
  GET  /json        — JSON serialization
  GET  /db          — Single random row query
  GET  /queries     — Multiple random row queries
  POST /updates     — Update random rows

Usage: python3 server_cpython_gohttp.py [port]
Requires: gofre build (in examples/webserver/)
"""

import json
import os
import random
import sqlite3
import sys

import cffi

ffi = cffi.FFI()

# Find the shared library
pkg_dir = os.path.join(os.path.dirname(__file__), "..", "..", "build", "gofre_webserver")
lib_path = os.path.join(pkg_dir, "_binding.dylib")
if not os.path.exists(lib_path):
    lib_path = os.path.join(pkg_dir, "_binding.so")
if not os.path.exists(lib_path):
    print("Go shared library not found. Run 'gofre build' in examples/webserver/ first.")
    print(f"Expected at: {lib_path}")
    sys.exit(1)

ffi.cdef("""
    char* Dumps(char* obj);
    char* Loads(char* s);
    int   HTTPCreateServer(void);
    void  HTTPAddRoute(int serverID, char* method, char* path, int handlerID);
    void  HTTPStartServer(int serverID, char* addr);
    void  HTTPSetDispatch(void* fn);
""")

lib = ffi.dlopen(lib_path)

DB_PATH = "benchmark_cpython_gohttp.db"


def init_db():
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()
    c.execute("PRAGMA journal_mode=WAL")
    c.execute("CREATE TABLE IF NOT EXISTS world (id INTEGER PRIMARY KEY, randomNumber INTEGER NOT NULL)")
    c.execute("SELECT COUNT(*) FROM world")
    if c.fetchone()[0] == 0:
        data = [(i, random.randint(1, 10000)) for i in range(1, 10001)]
        c.executemany("INSERT INTO world (id, randomNumber) VALUES (?, ?)", data)
        conn.commit()
    conn.close()


init_db()


# ── Handler functions ─────────────────────────────────────────────


def handle_plaintext(body):
    return "Hello, World!"


def handle_json(body):
    return json.dumps(
        {
            "message": "Hello, World!",
            "timestamp": 1234567890,
            "random": 42,
            "data": {
                "name": "benchmark",
                "version": "1.0.0",
                "features": ["json", "db", "template"],
                "metadata": {"host": "localhost", "port": 8080},
            },
        }
    )


def handle_db(body):
    conn = sqlite3.connect(DB_PATH)
    row = conn.execute("SELECT id, randomNumber FROM world ORDER BY RANDOM() LIMIT 1").fetchone()
    conn.close()
    return json.dumps({"id": row[0], "randomNumber": row[1]})


def handle_queries(body):
    n = 1
    if body:
        for part in body.split("&"):
            if part.startswith("N="):
                try:
                    n = max(1, min(500, int(part[2:])))
                except ValueError:
                    pass
    conn = sqlite3.connect(DB_PATH)
    rows = conn.execute("SELECT id, randomNumber FROM world ORDER BY RANDOM() LIMIT ?", (n,)).fetchall()
    conn.close()
    return json.dumps([{"id": r[0], "randomNumber": r[1]} for r in rows])


def handle_updates(body):
    updates = json.loads(body)
    conn = sqlite3.connect(DB_PATH)
    for u in updates:
        conn.execute(
            "UPDATE world SET randomNumber = ? WHERE id = ?",
            (u["randomNumber"], u["id"]),
        )
    conn.commit()
    conn.close()
    return json.dumps(updates)


# Handler dispatch table: handler_id -> handler function
HANDLERS = {
    0: handle_plaintext,
    1: handle_json,
    2: handle_db,
    3: handle_queries,
    4: handle_updates,
}


# ── CFFI dispatch callback ────────────────────────────────────────
# Signature: void dispatch(int handler_id, char* body, char* out_buf)


@ffi.callback("void(int, char*, char*)")
def dispatch(handler_id, body, out_buf):
    body_str = ffi.string(body).decode() if body != ffi.NULL else ""
    handler = HANDLERS.get(handler_id)
    if handler is None:
        result = json.dumps({"error": f"unknown handler {handler_id}"})
    else:
        try:
            result = handler(body_str)
        except Exception as e:
            result = json.dumps({"error": str(e)})
    result_bytes = result.encode()
    ffi.memmove(out_buf, result_bytes, len(result_bytes) + 1)


# ── Server startup ────────────────────────────────────────────────

if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8087

    lib.HTTPSetDispatch(dispatch)
    server_id = lib.HTTPCreateServer()

    lib.HTTPAddRoute(server_id, b"GET", b"/plaintext", 0)
    lib.HTTPAddRoute(server_id, b"GET", b"/json", 1)
    lib.HTTPAddRoute(server_id, b"GET", b"/db", 2)
    lib.HTTPAddRoute(server_id, b"GET", b"/queries", 3)
    lib.HTTPAddRoute(server_id, b"POST", b"/updates", 4)

    addr = f"0.0.0.0:{port}"
    lib.HTTPStartServer(server_id, addr.encode())

    print(f"CPython+Go HTTP benchmark server running on :{port}")
    try:
        import time

        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        print("\nShutting down...")
