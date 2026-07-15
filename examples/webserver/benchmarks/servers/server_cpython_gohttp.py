"""CPython webserver using gojson (Go JSON) via cffi for JSON serialization.

Routes:
  GET  /plaintext   — Plain text response
  GET  /json        — JSON serialization via Go
  GET  /db          — Single random row query
  GET  /queries     — Multiple random row queries
  POST /updates     — Update random rows

Usage: python3 server_cpython_gohttp.py [port]
Requires: gofre build (in examples/webserver/)
"""

import http.server
import json
import os
import random
import sqlite3
import sys
import threading
import urllib.parse

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
""")

lib = ffi.dlopen(lib_path)


def go_json_dumps(obj):
    """Serialize a Python object to JSON string using Go."""
    obj_str = json.dumps(obj)
    c_str = ffi.new("char[]", obj_str.encode())
    result = lib.Dumps(c_str)
    return ffi.string(result).decode()


def go_json_loads(s):
    """Deserialize a JSON string to Python object using Go."""
    c_str = ffi.new("char[]", s.encode())
    result = lib.Loads(c_str)
    return json.loads(ffi.string(result).decode())


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


class BenchmarkHandler(http.server.BaseHTTPRequestHandler):
    """HTTP handler with TFB benchmark endpoints using Go JSON via cffi."""

    def do_GET(self):
        if self.path == "/plaintext":
            self._handle_plaintext()
        elif self.path == "/json":
            self._handle_json()
        elif self.path.startswith("/db"):
            self._handle_db()
        elif self.path.startswith("/queries"):
            self._handle_queries()
        else:
            self._send_response(404, go_json_dumps({"error": "not found"}))

    def do_POST(self):
        if self.path == "/updates":
            self._handle_updates()
        else:
            self._send_response(404, go_json_dumps({"error": "not found"}))

    def _handle_plaintext(self):
        body = "Hello, World!"
        self.send_response(200)
        self.send_header("Content-Type", "text/plain")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body.encode())

    def _handle_json(self):
        body = go_json_dumps(
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
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body.encode())

    def _handle_db(self):
        conn = sqlite3.connect(DB_PATH)
        row = conn.execute("SELECT id, randomNumber FROM world ORDER BY RANDOM() LIMIT 1").fetchone()
        conn.close()
        self._send_response(200, go_json_dumps({"id": row[0], "randomNumber": row[1]}))

    def _handle_queries(self):
        n = 1
        if "?" in self.path:
            qs = urllib.parse.urlparse(self.path).query
            for part in qs.split("&"):
                if part.startswith("N="):
                    try:
                        n = max(1, min(500, int(part[2:])))
                    except ValueError:
                        pass
        conn = sqlite3.connect(DB_PATH)
        rows = conn.execute("SELECT id, randomNumber FROM world ORDER BY RANDOM() LIMIT ?", (n,)).fetchall()
        conn.close()
        self._send_response(200, go_json_dumps([{"id": r[0], "randomNumber": r[1]} for r in rows]))

    def _handle_updates(self):
        length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(length).decode()
        updates = go_json_loads(body)
        conn = sqlite3.connect(DB_PATH)
        for u in updates:
            conn.execute(
                "UPDATE world SET randomNumber = ? WHERE id = ?",
                (u["randomNumber"], u["id"]),
            )
        conn.commit()
        conn.close()
        self._send_response(200, go_json_dumps(updates))

    def _send_response(self, status, body):
        body_bytes = body.encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body_bytes)))
        self.end_headers()
        self.wfile.write(body_bytes)

    def log_message(self, format, *args):
        pass


class ThreadedHTTPServer(http.server.HTTPServer):
    allow_reuse_address = True

    def process_request(self, request, client_address):
        t = threading.Thread(target=self._handle, args=(request, client_address))
        t.daemon = True
        t.start()

    def _handle(self, request, client_address):
        try:
            self.finish_request(request, client_address)
        except Exception:
            self.handle_error(request, client_address)
        finally:
            self.shutdown_request(request)


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8087
    server = ThreadedHTTPServer(("0.0.0.0", port), BenchmarkHandler)
    print(f"CPython+Go JSON benchmark server running on :{port}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down...")
        server.server_close()
