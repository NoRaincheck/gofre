"""Flask + Werkzeug benchmark server with TFB-style endpoints.

Routes:
  GET  /plaintext   — Plain text response
  GET  /json        — JSON serialization

Usage: python3 server_flask.py [port]
"""

import sys

from flask import Flask, Response, jsonify

app = Flask(__name__)


# ── Plaintext ─────────────────────────────────────────────────────
@app.route("/plaintext")
def plaintext():
    return Response("Hello, World!", status=200, content_type="text/plain")


# ── JSON ──────────────────────────────────────────────────────────
@app.route("/json")
def json_endpoint():
    return jsonify(
        message="Hello, World!",
        timestamp=1234567890,
        random=42,
        data={
            "name": "benchmark",
            "version": "1.0.0",
            "features": ["json", "db", "template"],
            "metadata": {"host": "localhost", "port": 8080},
        },
    )


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8084
    app.run(host="0.0.0.0", port=port, debug=False, threaded=True)
