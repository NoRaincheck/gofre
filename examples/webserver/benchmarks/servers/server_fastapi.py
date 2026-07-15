"""FastAPI + uvicorn benchmark server with TFB-style endpoints.

Routes:
  GET  /plaintext   — Plain text response
  GET  /json        — JSON serialization

Usage: python3 server_fastapi.py [port]
"""

import sys

from fastapi import FastAPI
from fastapi.responses import PlainTextResponse

app = FastAPI()


# ── Plaintext ─────────────────────────────────────────────────────
@app.get("/plaintext")
async def plaintext():
    return PlainTextResponse(content="Hello, World!")


# ── JSON ──────────────────────────────────────────────────────────
@app.get("/json")
async def json_endpoint():
    return {
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


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8083
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=port, log_level="warning", workers=1)
