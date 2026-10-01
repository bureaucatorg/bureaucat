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
SELECT t.id AS task_id, t.task_number, t.title,
       (t.parent_task_id IS NOT NULL)::boolean AS is_subtask,
       p.project_key, p.name AS project_name,
       w.id AS workspace_id, w.workspace_key, w.name AS workspace_name,
       ps.name AS state_name, ps.state_type, ps.color AS state_color,
       u.id AS user_id, u.username, u.email, u.first_name, u.last_name, u.avatar_url
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN workspaces w ON p.workspace_id = w.id
JOIN project_states ps ON t.state_id = ps.id
JOIN task_assignees ta ON ta.task_id = t.id
JOIN users u ON ta.user_id = u.id
WHERE t.deleted_at IS NULL
  AND p.deleted_at IS NULL
  AND w.deleted_at IS NULL
  AND (sqlc.narg('workspace_key')::text IS NULL OR w.workspace_key = sqlc.narg('workspace_key')::text)
  AND (COALESCE(cardinality(sqlc.arg('project_keys')::text[]), 0) = 0 OR p.project_key = ANY(sqlc.arg('project_keys')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('state_types')::text[]), 0) = 0 OR ps.state_type::text = ANY(sqlc.arg('state_types')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('usernames')::text[]), 0) = 0 OR u.username = ANY(sqlc.arg('usernames')::text[]))
ORDER BY p.project_key ASC, t.task_number ASC;

-- name: ListGraphWorkspaceOptions :many
SELECT workspace_key, name
FROM workspaces
WHERE deleted_at IS NULL
ORDER BY name ASC;

-- name: ListGraphProjectOptions :many
SELECT p.project_key, p.name, w.workspace_key
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
WHERE p.deleted_at IS NULL AND w.deleted_at IS NULL
ORDER BY p.project_key ASC;

-- name: ListGraphUserOptions :many
SELECT username, email, first_name, last_name
FROM users
ORDER BY first_name ASC, last_name ASC;
