-- name: CreateAttachment :one
INSERT INTO attachments (upload_id, entity_type, entity_id, created_by)
VALUES ($1, $2, $3, $4)
RETURNING id, upload_id, entity_type, entity_id, created_by, created_at;

-- name: DeleteAttachment :exec
DELETE FROM attachments WHERE id = $1;

-- name: DeleteAttachmentsByEntity :exec
DELETE FROM attachments WHERE entity_type = $1 AND entity_id = $2;

-- name: ListAttachmentsByEntity :many
SELECT a.id, a.upload_id, a.entity_type, a.entity_id, a.created_by, a.created_at,
       u.filename, u.mime_type, u.size_bytes
FROM attachments a
JOIN uploads u ON u.id = a.upload_id
WHERE a.entity_type = $1 AND a.entity_id = $2
ORDER BY a.created_at;
