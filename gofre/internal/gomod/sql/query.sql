-- name: GetRandomWorld :one
SELECT * FROM world ORDER BY RANDOM() LIMIT 1;

-- name: GetRandomWorlds :many
SELECT * FROM world ORDER BY RANDOM() LIMIT ?;

-- name: UpdateWorld :exec
UPDATE world SET randomNumber = ? WHERE id = ?;

-- name: CountWorlds :one
SELECT COUNT(*) FROM world;
