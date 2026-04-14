-- name: CreatePersonalAccessToken :one
INSERT INTO personal_access_tokens (user_id, name, token_hash, expires_at, scope)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, name, token_hash, expires_at, last_used_at, created_at, scope;

-- name: DeletePersonalAccessToken :exec
DELETE FROM personal_access_tokens
WHERE id = $1 AND user_id = $2;

-- name: GetPersonalAccessTokenByHash :one
SELECT pat.id, pat.user_id, pat.expires_at, pat.last_used_at, pat.created_at, pat.scope,
       u.username, u.user_type
FROM personal_access_tokens pat
JOIN users u ON pat.user_id = u.id
WHERE pat.token_hash = $1
  AND (pat.expires_at IS NULL OR pat.expires_at > NOW());

-- name: ListPersonalAccessTokensByUser :many
SELECT id, name, expires_at, last_used_at, created_at, scope
FROM personal_access_tokens
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdatePersonalAccessTokenLastUsed :exec
UPDATE personal_access_tokens
SET last_used_at = NOW()
WHERE id = $1;

-- name: UpdatePersonalAccessTokenScope :one
UPDATE personal_access_tokens
SET scope = $3
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, name, expires_at, last_used_at, created_at, scope;
