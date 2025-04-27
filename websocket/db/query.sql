-- name: CreateTableIfNotExistsTemperature :exec
CREATE TABLE IF NOT EXISTS temperatures (
  id SERIAL PRIMARY KEY,
  temperature REAL NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- name: CreateTemperature :one
INSERT INTO temperatures (
  temperature
) VALUES (
  ?
)
RETURNING id, temperature, created_at;
