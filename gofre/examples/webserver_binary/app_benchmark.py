"""Benchmark application for Go+Pocketpy server (TFB-style endpoints).

Implements TFB-style benchmark endpoints:
  GET  /plaintext   — Plain text response
  GET  /json        — JSON serialization

Uses gohttp and gojson modules provided by the Go host.
"""

import gohttp
import gojson


app = gohttp.Server()


@app.route("/plaintext")
def plaintext(body):
    return gohttp.Response(status=200, headers=[("Content-Type", "text/plain")], body="Hello, World!")


@app.route("/json")
def json_endpoint(body):
    resp = {
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
    return gohttp.Response(status=200, headers=[("Content-Type", "application/json")], body=gojson.dumps(resp))


if __name__ == "__main__":
    app.run(addr=":8086")
