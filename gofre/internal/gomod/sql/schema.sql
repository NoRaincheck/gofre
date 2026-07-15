CREATE TABLE IF NOT EXISTS world (
    id INTEGER PRIMARY KEY,
    randomNumber INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_world_random ON world(randomNumber);
