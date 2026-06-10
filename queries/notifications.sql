-- ==================== NOTIFICATIONS ====================

-- name: GetOpenNotification :one
-- Most recent notification for a (recipient, task) pair still within the
-- coalescing window. Used to merge new activity into an existing notification
-- instead of creating a new one (max 1 notification per task per window).
SELECT id, recipient_id, task_id, activity_type, actor_id, event_count, read_at, created_at, updated_at
FROM notifications
WHERE recipient_id = sqlc.arg('recipient_id')
  AND task_id = sqlc.arg('task_id')
  AND created_at > sqlc.arg('cutoff')
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateNotification :one
INSERT INTO notifications (recipient_id, task_id, activity_type, actor_id)
VALUES (sqlc.arg('recipient_id'), sqlc.arg('task_id'), sqlc.arg('activity_type'), sqlc.arg('actor_id'))
RETURNING id, recipient_id, task_id, activity_type, actor_id, event_count, read_at, created_at, updated_at;

-- name: CoalesceNotification :exec
-- Merge a new activity into an existing open notification: bump the count,
-- update the latest actor/type, and re-surface as unread.
UPDATE notifications
SET event_count   = event_count + 1,
    activity_type = sqlc.arg('activity_type'),
    actor_id      = sqlc.arg('actor_id'),
    read_at       = NULL,
    updated_at    = NOW()
WHERE id = sqlc.arg('id');

-- name: ListNotifications :many
-- A recipient's notifications, newest first, with task/project/actor display fields.
SELECT n.id, n.task_id, n.activity_type, n.actor_id, n.event_count, n.read_at, n.created_at, n.updated_at,
       u.username, u.first_name, u.last_name, u.avatar_url,
       t.task_number, t.title AS task_title, p.project_key
FROM notifications n
JOIN users u ON n.actor_id = u.id
JOIN tasks t ON n.task_id = t.id
JOIN projects p ON t.project_id = p.id
WHERE n.recipient_id = sqlc.arg('recipient_id')
  AND t.deleted_at IS NULL
ORDER BY n.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountNotifications :one
SELECT COUNT(*)
FROM notifications n
JOIN tasks t ON n.task_id = t.id
WHERE n.recipient_id = sqlc.arg('recipient_id')
  AND t.deleted_at IS NULL;

-- name: CountUnreadNotifications :one
SELECT COUNT(*)
FROM notifications n
JOIN tasks t ON n.task_id = t.id
WHERE n.recipient_id = sqlc.arg('recipient_id')
  AND t.deleted_at IS NULL
  AND n.read_at IS NULL;

-- name: MarkNotificationRead :exec
UPDATE notifications
SET read_at = NOW()
WHERE id = sqlc.arg('id')
  AND recipient_id = sqlc.arg('recipient_id')
  AND read_at IS NULL;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read_at = NOW()
WHERE recipient_id = sqlc.arg('recipient_id')
  AND read_at IS NULL;
