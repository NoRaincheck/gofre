"""FastAPI + uvicorn benchmark server with TFB-style endpoints.

Routes:
  GET  /plaintext   — Plain text response
  GET  /json        — JSON serialization
  GET  /db          — Single random row query
  GET  /queries     — Multiple random row queries
  POST /updates     — Update random rows

Usage: python3 server_fastapi.py [port]
"""

import random
import sys

from fastapi import FastAPI
from fastapi.responses import PlainTextResponse
from pydantic import BaseModel
from sqlmodel import Field, Session, SQLModel, create_engine, select


class World(SQLModel, table=True):
    id: int = Field(primary_key=True)
    randomNumber: int = Field(default=0)


DATABASE_URL = "sqlite:///benchmark_fastapi.db"
engine = create_engine(
    DATABASE_URL,
    connect_args={"check_same_thread": False},
    echo=False,
)


def init_db():
    SQLModel.metadata.create_all(engine)
    with Session(engine) as session:
        count = session.exec(select(World)).all()
        if len(count) == 0:
            for i in range(1, 10001):
                session.add(World(id=i, randomNumber=random.randint(1, 10000)))
            session.commit()


init_db()

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


# ── DB Single Query ───────────────────────────────────────────────
@app.get("/db")
async def db_single():
    with Session(engine) as session:
        # Use raw SQL for RANDOM() since sqlmodel doesn't support it directly
        result = session.exec("SELECT id, randomNumber FROM world ORDER BY RANDOM() LIMIT 1").first()
        return {"id": result[0], "randomNumber": result[1]}


# ── DB Multiple Queries ──────────────────────────────────────────
@app.get("/queries")
async def db_queries(N: int = 1):
    N = max(1, min(500, N))
    with Session(engine) as session:
        result = session.exec("SELECT id, randomNumber FROM world ORDER BY RANDOM() LIMIT ?", (N,)).all()
        return [{"id": r[0], "randomNumber": r[1]} for r in result]


# ── DB Updates ───────────────────────────────────────────────────
class UpdateItem(BaseModel):
    id: int
    randomNumber: int


@app.post("/updates")
async def db_updates(updates: list[UpdateItem]):
    with Session(engine) as session:
        for u in updates:
            session.exec(
                "UPDATE world SET randomNumber = ? WHERE id = ?",
                (u.randomNumber, u.id),
            )
        session.commit()
    return updates


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8083
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=port, log_level="warning", workers=1)
