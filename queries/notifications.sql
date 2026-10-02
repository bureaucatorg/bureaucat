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
INSERT INTO notifications (recipient_id, task_id, activity_type, actor_id, comment_id)
VALUES (sqlc.arg('recipient_id'), sqlc.arg('task_id'), sqlc.arg('activity_type'), sqlc.arg('actor_id'), sqlc.narg('comment_id'))
RETURNING id, recipient_id, task_id, activity_type, actor_id, event_count, read_at, created_at, updated_at;

-- name: CoalesceNotification :exec
-- Merge a new activity into an existing open notification: bump the count,
-- update the latest actor/type/comment, and re-surface as unread.
UPDATE notifications
SET event_count   = event_count + 1,
    activity_type = sqlc.arg('activity_type'),
    actor_id      = sqlc.arg('actor_id'),
    comment_id    = sqlc.narg('comment_id'),
    read_at       = NULL,
    updated_at    = NOW()
WHERE id = sqlc.arg('id');

-- name: ListNotifications :many
-- A recipient's notifications, newest first, with task/project/actor display fields.
SELECT n.id, n.task_id, n.activity_type, n.actor_id, n.comment_id, n.event_count, n.read_at, n.created_at, n.updated_at,
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

-- ==================== EMAIL DIGEST ====================

-- name: ClaimEmailNotifications :many
-- Claim unread notifications whose coalescing window has closed, for opted-in
-- recipients, marking them emailed. Atomic, so concurrent workers never double-send.
UPDATE notifications n
SET emailed_at = NOW()
FROM users r, users a, tasks t, projects p, project_states s
WHERE n.emailed_at IS NULL
  AND n.created_at <= sqlc.arg('cutoff')
  AND n.read_at IS NULL
  AND r.id = n.recipient_id
  AND r.email_notifications
  AND a.id = n.actor_id
  AND t.id = n.task_id
  AND t.deleted_at IS NULL
  AND p.id = t.project_id
  AND s.id = t.state_id
RETURNING n.id, n.task_id, n.recipient_id, n.activity_type, n.comment_id, n.event_count, n.created_at, n.updated_at,
          r.email AS recipient_email,
          a.first_name AS actor_first_name, a.last_name AS actor_last_name,
          t.task_number, t.title AS task_title,
          p.project_key, p.name AS project_name,
          s.name AS state_name, s.color AS state_color;

-- name: ListEmailActivity :many
-- The activity batched into one notification: other users' changes to the task
-- within the notification's lifetime, oldest first.
SELECT al.activity_type, al.field_name, al.old_value, al.new_value, u.first_name, u.last_name
FROM activity_log al
JOIN users u ON u.id = al.actor_id
WHERE al.task_id = sqlc.arg('task_id')
  AND al.actor_id <> sqlc.arg('recipient_id')
  AND al.created_at >= sqlc.arg('since')
  AND al.created_at <= sqlc.arg('until')
ORDER BY al.created_at ASC
LIMIT 20;

-- name: SkipEmailNotifications :exec
-- Mark the remaining closed-window rows processed so they are never emailed later.
UPDATE notifications
SET emailed_at = NOW()
WHERE emailed_at IS NULL
  AND created_at <= sqlc.arg('cutoff');

-- name: GetUserEmailNotifications :one
SELECT email_notifications FROM users WHERE id = $1;

-- name: UpdateUserEmailNotifications :exec
UPDATE users
SET email_notifications = $2, updated_at = NOW()
WHERE id = $1;
