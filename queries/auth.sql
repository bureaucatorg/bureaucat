-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, first_name, last_name, user_type)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, username, email, first_name, last_name, user_type, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, username, email, first_name, last_name, user_type, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmailOrUsername :one
SELECT id, username, email, password_hash, first_name, last_name, user_type, created_at, updated_at
FROM users
WHERE email = $1 OR username = $1;

-- name: UserExistsByEmailOrUsername :one
SELECT EXISTS (
    SELECT 1 FROM users
    WHERE email = $1 OR username = $2
) AS exists;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token_hash, expires_at, created_at, revoked_at;

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
FROM refresh_tokens
WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW();

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE id = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;
