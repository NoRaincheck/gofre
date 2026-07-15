"""Benchmark application for Go+Pocketpy server (TFB-style endpoints).

Implements TFB-style benchmark endpoints:
  GET  /plaintext   — Plain text response
  GET  /json        — JSON serialization
  GET  /db          — Single random row query
  GET  /queries     — Multiple random row queries
  POST /updates     — Update random rows

Uses gohttp, gojson, and gosql modules provided by the Go host.
"""

import gohttp
import gojson
import gosql


class Server:
    def __init__(self):
        self._id = gohttp.create_server()
        self._handler_counter = 0

    def route(self, path, method="GET"):
        def wrapper(fn):
            self._handler_counter += 1
            handler_id = self._handler_counter
            func_name = "_handler_%d" % handler_id
            globals()[func_name] = fn
            gohttp.add_route(self._id, method, path, handler_id)
            return fn

        return wrapper

    def run(self, addr=":8086"):
        gohttp.start_server(self._id, addr)


app = Server()

# Initialize database
db = gosql.open("benchmark.db")
gosql.seed(db)


@app.route("/plaintext")
def plaintext(body):
    return "Hello, World!"


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
    return gojson.dumps(resp)


@app.route("/db")
def db_single(body):
    row = gosql.query_row(db, "SELECT id, randomNumber FROM world ORDER BY RANDOM() LIMIT 1")
    return row


@app.route("/queries")
def db_queries(body):
    # body contains query string like "N=10"
    n = 1
    if body:
        for part in body.split("&"):
            if part.startswith("N="):
                try:
                    n = int(part[2:])
                except Exception:
                    pass
    if n < 1:
        n = 1
    if n > 500:
        n = 500
    rows = gosql.query_rows(db, "SELECT id, randomNumber FROM world ORDER BY RANDOM() LIMIT ?", n)
    return rows


@app.route("/updates", method="POST")
def db_updates(body):
    updates = gojson.loads(body)
    for u in updates:
        gosql.exec(db, "UPDATE world SET randomNumber = ? WHERE id = ?", u["randomNumber"], u["id"])
    return gojson.dumps(updates)


if __name__ == "__main__":
    app.run(addr=":8086")
