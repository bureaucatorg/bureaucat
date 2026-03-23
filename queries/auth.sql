-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, first_name, last_name, user_type)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, username, email, first_name, last_name, user_type, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, username, email, first_name, last_name, user_type, avatar_url, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmailOrUsername :one
SELECT id, username, email, password_hash, first_name, last_name, user_type,
       auth_provider, provider_user_id, created_at, updated_at
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

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: ListUsersPaginated :many
SELECT id, username, email, first_name, last_name, user_type, created_at, updated_at
FROM users
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- name: SearchUsersPaginated :many
SELECT id, username, email, first_name, last_name, user_type, created_at, updated_at
FROM users
WHERE username ILIKE '%' || $1 || '%'
   OR email ILIKE '%' || $1 || '%'
   OR first_name ILIKE '%' || $1 || '%'
   OR last_name ILIKE '%' || $1 || '%'
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;

-- name: CountSearchUsers :one
SELECT COUNT(*) FROM users
WHERE username ILIKE '%' || $1 || '%'
   OR email ILIKE '%' || $1 || '%'
   OR first_name ILIKE '%' || $1 || '%'
   OR last_name ILIKE '%' || $1 || '%';

-- name: DeleteUserByID :exec
DELETE FROM users WHERE id = $1;

-- name: ListActiveRefreshTokens :many
SELECT
    rt.id,
    rt.user_id,
    rt.expires_at,
    rt.created_at,
    u.username,
    u.email
FROM refresh_tokens rt
JOIN users u ON rt.user_id = u.id
WHERE rt.revoked_at IS NULL AND rt.expires_at > NOW()
ORDER BY rt.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountActiveRefreshTokens :one
SELECT COUNT(*)
FROM refresh_tokens
WHERE revoked_at IS NULL AND expires_at > NOW();

-- name: GetRefreshTokenByID :one
SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
FROM refresh_tokens
WHERE id = $1;

-- name: DeleteExpiredRefreshTokens :execrows
DELETE FROM refresh_tokens
WHERE expires_at <= NOW();

-- name: GetUserByEmail :one
SELECT id, username, email, password_hash, first_name, last_name, user_type,
       auth_provider, provider_user_id, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByProviderID :one
SELECT id, username, email, password_hash, first_name, last_name, user_type,
       auth_provider, provider_user_id, created_at, updated_at
FROM users
WHERE auth_provider = $1 AND provider_user_id = $2;

-- name: CreateSSOUser :one
INSERT INTO users (username, email, first_name, last_name, user_type, auth_provider, provider_user_id, avatar_url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, username, email, first_name, last_name, user_type, auth_provider, provider_user_id, avatar_url, created_at, updated_at;

-- name: LinkProviderToUser :exec
UPDATE users
SET auth_provider = $2, provider_user_id = $3, updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserAvatarURL :exec
UPDATE users
SET avatar_url = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserType :exec
UPDATE users
SET user_type = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;
