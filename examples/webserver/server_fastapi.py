"""FastAPI server for benchmark comparison."""

import json

from fastapi import FastAPI, Request

app = FastAPI()


@app.get("/")
async def home():
    return {"message": "hello from FastAPI!"}


@app.get("/api/data")
async def api_data():
    return {"items": [1, 2, 3], "service": "FastAPI"}


@app.post("/api/echo")
async def api_echo(request: Request):
    body = await request.body()
    data = json.loads(body)
    return data


if __name__ == "__main__":
    import sys

    import uvicorn

    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8083
    uvicorn.run(app, host="0.0.0.0", port=port, log_level="warning")
