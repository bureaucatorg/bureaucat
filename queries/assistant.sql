-- name: CreateAssistantConversation :one
INSERT INTO assistant_conversations (user_id, title)
VALUES ($1, $2)
RETURNING *;

-- name: ListAssistantConversations :many
SELECT * FROM assistant_conversations
WHERE user_id = $1
ORDER BY updated_at DESC
LIMIT 100;

-- name: GetAssistantConversation :one
SELECT * FROM assistant_conversations
WHERE id = $1 AND user_id = $2;

-- name: DeleteAssistantConversation :execrows
DELETE FROM assistant_conversations
WHERE id = $1 AND user_id = $2;

-- name: TouchAssistantConversation :exec
UPDATE assistant_conversations
SET updated_at = NOW(),
    title = CASE WHEN title = '' THEN sqlc.arg(title)::text ELSE title END
WHERE id = sqlc.arg(id);

-- name: CreateAssistantMessage :one
INSERT INTO assistant_messages (conversation_id, role, content, tool_calls, tool_call_id, tool_name, attachments, meta)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListAssistantMessages :many
SELECT * FROM assistant_messages
WHERE conversation_id = $1
ORDER BY seq;

-- name: CreateAssistantAction :one
INSERT INTO assistant_actions (conversation_id, tool_call_id, tool_name, args)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAssistantAction :one
SELECT * FROM assistant_actions
WHERE id = $1 AND conversation_id = $2;

-- name: ListAssistantActions :many
SELECT * FROM assistant_actions
WHERE conversation_id = $1
ORDER BY created_at;

-- name: ListPendingAssistantActions :many
SELECT * FROM assistant_actions
WHERE conversation_id = $1 AND status = 'pending'
ORDER BY created_at;

-- name: ResolveAssistantAction :one
UPDATE assistant_actions
SET status = $2, result = $3, resolved_at = NOW()
WHERE id = $1 AND status = 'pending'
RETURNING *;
