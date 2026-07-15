"""Pure Python baseline webserver with TFB-style benchmark endpoints.

Routes:
  GET  /plaintext   — Plain text response
  GET  /json        — JSON serialization

Usage: python3 server_pure.py [port]
"""

import http.server
import json
import sys
import threading


class BenchmarkHandler(http.server.BaseHTTPRequestHandler):
    """HTTP handler with TFB benchmark endpoints."""

    def do_GET(self):
        if self.path == "/plaintext":
            self._handle_plaintext()
        elif self.path == "/json":
            self._handle_json()
        else:
            self._send_response(404, json.dumps({"error": "not found"}))

    # ── Plaintext ──────────────────────────────────────────────────
    def _handle_plaintext(self):
        body = "Hello, World!"
        self.send_response(200)
        self.send_header("Content-Type", "text/plain")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body.encode())

    # ── JSON ───────────────────────────────────────────────────────
    def _handle_json(self):
        response = {
            "message": "Hello, World!",
            "timestamp": 1234567890,
            "random": 42,
            "data": {
                "name": "benchmark",
                "version": "1.0.0",
                "features": ["json", "db", "template"],
                "metadata": {
                    "host": "localhost",
                    "port": 8080,
                },
            },
        }
        body = json.dumps(response)
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body.encode())

    def _send_response(self, status, body):
        body_bytes = body.encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body_bytes)))
        self.end_headers()
        self.wfile.write(body_bytes)

    def log_message(self, format, *args):
        pass  # Suppress logs for benchmark cleanliness


class ThreadedHTTPServer(http.server.HTTPServer):
    """HTTP server that handles each request in a new thread."""

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
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8082
    server = ThreadedHTTPServer(("0.0.0.0", port), BenchmarkHandler)
    print(f"Pure Python benchmark server running on :{port}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down...")
        server.server_close()
