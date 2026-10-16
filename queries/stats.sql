-- name: CountWorkspaces :one
SELECT COUNT(*) FROM workspaces WHERE deleted_at IS NULL;

-- name: CountProjects :one
SELECT COUNT(*) FROM projects WHERE deleted_at IS NULL;

-- name: CountTopLevelTasks :one
SELECT COUNT(*) FROM tasks WHERE deleted_at IS NULL AND parent_task_id IS NULL;

-- name: CountSubtasks :one
SELECT COUNT(*) FROM tasks WHERE deleted_at IS NULL AND parent_task_id IS NOT NULL;

-- name: CountPages :one
SELECT COUNT(*) FROM pages WHERE deleted_at IS NULL;

-- name: CountEmailNotificationUsers :one
SELECT COUNT(*) FROM users WHERE email_notifications;

-- name: CountActiveBlockerLinks :one
-- Only links where neither task is deleted, completed or cancelled.
SELECT COUNT(DISTINCT tb.task_id)::bigint AS blocked_tasks,
       COUNT(DISTINCT tb.blocker_task_id)::bigint AS blocking_tasks
FROM task_blockers tb
JOIN tasks bt ON bt.id = tb.task_id AND bt.deleted_at IS NULL
JOIN project_states bs ON bs.id = bt.state_id AND bs.state_type NOT IN ('completed', 'cancelled')
JOIN tasks kt ON kt.id = tb.blocker_task_id AND kt.deleted_at IS NULL
JOIN project_states ks ON ks.id = kt.state_id AND ks.state_type NOT IN ('completed', 'cancelled');

-- name: CountAssistantConversations :one
-- Only conversations with at least one user message.
SELECT COUNT(*)::bigint AS conversations,
       COUNT(DISTINCT c.user_id)::bigint AS users,
       (COUNT(*) FILTER (WHERE EXISTS (
           SELECT 1 FROM assistant_mattermost_threads mt WHERE mt.conversation_id = c.id
       )))::bigint AS mattermost_conversations
FROM assistant_conversations c
WHERE EXISTS (SELECT 1 FROM assistant_messages m WHERE m.conversation_id = c.id AND m.role = 'user');

-- name: CountAttachments :one
SELECT COUNT(*) FROM attachments;

-- name: AttachmentsTotalSize :one
SELECT COALESCE(SUM(u.size_bytes), 0)::bigint AS total_bytes
FROM attachments a
JOIN uploads u ON u.id = a.upload_id;

-- name: TasksByStateType :many
SELECT ps.state_type AS state_type, COUNT(t.id)::int AS count
FROM tasks t
JOIN project_states ps ON t.state_id = ps.id
WHERE t.deleted_at IS NULL AND t.parent_task_id IS NULL
GROUP BY ps.state_type
ORDER BY count DESC;

-- name: TasksByPriority :many
SELECT t.priority AS priority, COUNT(*)::int AS count
FROM tasks t
WHERE t.deleted_at IS NULL AND t.parent_task_id IS NULL
GROUP BY t.priority
ORDER BY t.priority ASC;

-- name: TopProjectsByTaskCount :many
SELECT p.id AS project_id, p.name, p.project_key, COUNT(t.id)::int AS task_count
FROM projects p
LEFT JOIN tasks t ON t.project_id = p.id AND t.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id, p.name, p.project_key
ORDER BY task_count DESC
LIMIT 10;

-- name: ProjectsPerWorkspace :many
SELECT w.id AS workspace_id, w.name, w.workspace_key, COUNT(p.id)::int AS project_count
FROM workspaces w
LEFT JOIN projects p ON p.workspace_id = w.id AND p.deleted_at IS NULL
WHERE w.deleted_at IS NULL
GROUP BY w.id, w.name, w.workspace_key
ORDER BY project_count DESC;

-- name: TasksCreatedPerDay :many
SELECT d::date AS day, COUNT(t.id)::int AS count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN tasks t
    ON t.created_at::date = d::date
    AND t.deleted_at IS NULL
    AND t.parent_task_id IS NULL
GROUP BY d
ORDER BY d ASC;

-- name: SubtasksCreatedPerDay :many
SELECT d::date AS day, COUNT(t.id)::int AS count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN tasks t
    ON t.created_at::date = d::date
    AND t.deleted_at IS NULL
    AND t.parent_task_id IS NOT NULL
GROUP BY d
ORDER BY d ASC;

-- name: PagesCreatedPerDay :many
SELECT d::date AS day, COUNT(pg.id)::int AS count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN pages pg
    ON pg.created_at::date = d::date
    AND pg.deleted_at IS NULL
GROUP BY d
ORDER BY d ASC;

-- name: CommentsCreatedPerDay :many
SELECT d::date AS day, COUNT(c.id)::int AS count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN comments c
    ON c.created_at::date = d::date
    AND c.deleted_at IS NULL
GROUP BY d
ORDER BY d ASC;

-- name: ActivityCreatedPerDay :many
-- All activity_log events except comment lifecycle events (state changes,
-- assignee/label edits, task lifecycle, etc.).
SELECT d::date AS day, COUNT(al.id)::int AS count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN activity_log al
    ON al.created_at::date = d::date
    AND al.activity_type NOT IN ('comment_created', 'comment_updated', 'comment_deleted')
GROUP BY d
ORDER BY d ASC;

-- name: AttachmentsCreatedPerDay :many
SELECT d::date AS day, COUNT(a.id)::int AS count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN attachments a
    ON a.created_at::date = d::date
GROUP BY d
ORDER BY d ASC;

-- name: ViewsCreatedPerDay :many
SELECT
    d::date AS day,
    (COUNT(v.id) FILTER (WHERE v.visibility = 'private'))::int AS private_count,
    (COUNT(v.id) FILTER (WHERE v.visibility = 'shared'))::int AS shared_count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN project_views v
    ON v.created_at::date = d::date
    AND v.deleted_at IS NULL
GROUP BY d
ORDER BY d ASC;

-- name: AssistantConversationsPerDay :many
-- Same "has a user message" rule as CountAssistantConversations; Mattermost ones have a thread mapping.
SELECT
    d::date AS day,
    (COUNT(c.id) FILTER (WHERE c.id IS NOT NULL AND NOT c.is_mattermost))::int AS web_count,
    (COUNT(c.id) FILTER (WHERE c.is_mattermost))::int AS mattermost_count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN (
    SELECT ac.id, ac.created_at,
           EXISTS (SELECT 1 FROM assistant_mattermost_threads mt WHERE mt.conversation_id = ac.id) AS is_mattermost
    FROM assistant_conversations ac
    WHERE EXISTS (SELECT 1 FROM assistant_messages m WHERE m.conversation_id = ac.id AND m.role = 'user')
) c ON c.created_at::date = d::date
GROUP BY d
ORDER BY d ASC;

-- name: CyclesCreatedPerDay :many
SELECT d::date AS day, COUNT(cy.id)::int AS count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN cycles cy
    ON cy.created_at::date = d::date
    AND cy.deleted_at IS NULL
GROUP BY d
ORDER BY d ASC;

-- name: ModulesCreatedPerDay :many
SELECT d::date AS day, COUNT(m.id)::int AS count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN modules m
    ON m.created_at::date = d::date
    AND m.deleted_at IS NULL
GROUP BY d
ORDER BY d ASC;

-- name: ListTaskAssignmentsForGraph :many
-- One row per (task, assignee); unassigned tasks come back once with NULL user columns.
-- The usernames filter keeps whole tasks the user created or is assigned to.
SELECT t.id AS task_id, t.task_number, t.title,
       (t.parent_task_id IS NOT NULL)::boolean AS is_subtask, t.parent_task_id,
       p.project_key, p.name AS project_name,
       w.id AS workspace_id, w.workspace_key, w.name AS workspace_name,
       ps.name AS state_name, ps.state_type, ps.color AS state_color,
       u.id AS user_id, u.username, u.email, u.first_name, u.last_name, u.avatar_url,
       c.id AS creator_id, c.username AS creator_username, c.email AS creator_email,
       c.first_name AS creator_first_name, c.last_name AS creator_last_name, c.avatar_url AS creator_avatar_url
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN workspaces w ON p.workspace_id = w.id
JOIN project_states ps ON t.state_id = ps.id
JOIN users c ON t.created_by = c.id
LEFT JOIN task_assignees ta ON ta.task_id = t.id
LEFT JOIN users u ON ta.user_id = u.id
WHERE t.deleted_at IS NULL
  AND p.deleted_at IS NULL
  AND w.deleted_at IS NULL
  AND (sqlc.narg('workspace_key')::text IS NULL OR w.workspace_key = sqlc.narg('workspace_key')::text)
  AND (COALESCE(cardinality(sqlc.arg('project_keys')::text[]), 0) = 0 OR p.project_key = ANY(sqlc.arg('project_keys')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('state_types')::text[]), 0) = 0 OR ps.state_type::text = ANY(sqlc.arg('state_types')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('usernames')::text[]), 0) = 0
       OR c.username = ANY(sqlc.arg('usernames')::text[])
       OR EXISTS (
           SELECT 1 FROM task_assignees fa JOIN users fu ON fa.user_id = fu.id
           WHERE fa.task_id = t.id AND fu.username = ANY(sqlc.arg('usernames')::text[])))
  AND (sqlc.narg('viewer_id')::uuid IS NULL OR EXISTS (
      SELECT 1 FROM project_members vm WHERE vm.project_id = p.id AND vm.user_id = sqlc.narg('viewer_id')::uuid))
ORDER BY p.project_key ASC, t.task_number ASC;

-- name: ListBlockerLinksForGraph :many
-- Blocker links where both tasks are already on the graph.
SELECT task_id, blocker_task_id
FROM task_blockers
WHERE task_id = ANY(sqlc.arg('task_ids')::uuid[])
  AND blocker_task_id = ANY(sqlc.arg('task_ids')::uuid[]);

-- name: ListGraphWorkspaceOptions :many
SELECT w.workspace_key, w.name
FROM workspaces w
WHERE w.deleted_at IS NULL
  AND (sqlc.narg('viewer_id')::uuid IS NULL OR EXISTS (
      SELECT 1 FROM projects p
      JOIN project_members vm ON vm.project_id = p.id
      WHERE p.workspace_id = w.id AND p.deleted_at IS NULL AND vm.user_id = sqlc.narg('viewer_id')::uuid))
ORDER BY w.name ASC;

-- name: ListGraphProjectOptions :many
SELECT p.project_key, p.name, w.workspace_key
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
WHERE p.deleted_at IS NULL AND w.deleted_at IS NULL
  AND (sqlc.narg('viewer_id')::uuid IS NULL OR EXISTS (
      SELECT 1 FROM project_members vm WHERE vm.project_id = p.id AND vm.user_id = sqlc.narg('viewer_id')::uuid))
ORDER BY p.project_key ASC;

-- name: ListGraphUserOptions :many
SELECT u.username, u.email, u.first_name, u.last_name
FROM users u
WHERE sqlc.narg('viewer_id')::uuid IS NULL OR EXISTS (
    SELECT 1 FROM project_members um
    JOIN project_members vm ON vm.project_id = um.project_id
    JOIN projects p ON p.id = um.project_id AND p.deleted_at IS NULL
    WHERE um.user_id = u.id AND vm.user_id = sqlc.narg('viewer_id')::uuid)
ORDER BY u.first_name ASC, u.last_name ASC;
