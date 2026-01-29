-- name: CreateUpload :one
INSERT INTO uploads (filename, stored_name, mime_type, size_bytes, uploaded_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, filename, stored_name, mime_type, size_bytes, uploaded_by, created_at;

-- name: GetUploadByID :one
SELECT id, filename, stored_name, mime_type, size_bytes, uploaded_by, created_at
FROM uploads
WHERE id = $1;

-- name: DeleteUpload :exec
DELETE FROM uploads WHERE id = $1;

-- name: ListUploadsByUser :many
SELECT id, filename, stored_name, mime_type, size_bytes, uploaded_by, created_at
FROM uploads
WHERE uploaded_by = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
