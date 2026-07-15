"""Pure Python baseline webserver (no Go acceleration)."""

import http.server
import json
import sys


class PureHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/":
            data = {"message": "hello from pure Python!"}
            self._respond(200, json.dumps(data))
        elif self.path == "/api/data":
            data = {"items": [1, 2, 3], "service": "pure Python"}
            self._respond(200, json.dumps(data))
        else:
            self._respond(404, json.dumps({"error": "not found"}))

    def do_POST(self):
        content_length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(content_length).decode()
        if self.path == "/api/echo":
            try:
                data = json.loads(body)
                self._respond(200, json.dumps(data))
            except json.JSONDecodeError:
                self._respond(400, json.dumps({"error": "invalid json"}))
        else:
            self._respond(404, json.dumps({"error": "not found"}))

    def _respond(self, status, body):
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(body.encode())

    def log_message(self, format, *args):
        pass


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8082
    server = http.server.HTTPServer(("", port), PureHandler)
    print(f"Pure Python webserver running on :{port}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down...")
        server.server_close()
