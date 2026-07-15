"""CPython webserver using Go JSON extensions via cffi.

Run with: python3 server.py [port]
Requires the Go shared library to be built first:
  gofre build
"""

import http.server
import os
import sys

import cffi

ffi = cffi.FFI()

lib_path = os.path.join(os.path.dirname(__file__), "build", "gofre_webserver", "_binding.dylib")
if not os.path.exists(lib_path):
    lib_path = os.path.join(os.path.dirname(__file__), "build", "gofre_webserver", "_binding.so")
if not os.path.exists(lib_path):
    print("Go shared library not found. Run 'gofre build' first.")
    print(f"Expected at: {lib_path}")
    sys.exit(1)

ffi.cdef("""
    char* Dumps(char* obj);
    char* Loads(char* s);
""")

lib = ffi.dlopen(lib_path)


def go_dumps(obj_str):
    c_str = ffi.new("char[]", obj_str.encode())
    result = lib.Dumps(c_str)
    return ffi.string(result).decode()


def go_loads(s):
    c_str = ffi.new("char[]", s.encode())
    result = lib.Loads(c_str)
    return ffi.string(result).decode()


class GoHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/":
            self._respond(200, go_dumps('{"message": "hello from CPython + Go cffi!"}'))
        elif self.path == "/api/data":
            self._respond(200, go_dumps('{"items": [1, 2, 3], "service": "CPython"}'))
        else:
            self._respond(404, '{"error": "not found"}')

    def do_POST(self):
        content_length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(content_length).decode()
        if self.path == "/api/echo":
            self._respond(200, go_dumps(body))
        else:
            self._respond(404, '{"error": "not found"}')

    def _respond(self, status, body):
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(body.encode())

    def log_message(self, format, *args):
        pass  # suppress logs for benchmark


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8081
    server = http.server.HTTPServer(("", port), GoHandler)
    print(f"CPython+Go webserver running on :{port}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down...")
        server.server_close()
