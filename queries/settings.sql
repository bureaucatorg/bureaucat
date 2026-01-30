-- name: GetSetting :one
SELECT key, value, updated_at
FROM settings
WHERE key = $1;

-- name: UpsertSetting :one
INSERT INTO settings (key, value, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value, updated_at = NOW()
RETURNING key, value, updated_at;

-- name: ListSettings :many
SELECT key, value, updated_at
FROM settings
ORDER BY key;
