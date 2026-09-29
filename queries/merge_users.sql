-- Admin "merge user": additively grants the target everything the source has.
-- The source user is never modified and activity_log is never touched.

-- name: MergeCopyWorkspaceMembers :execrows
INSERT INTO workspace_members (workspace_id, user_id)
SELECT workspace_id, sqlc.arg('target_id')::uuid
FROM workspace_members
WHERE user_id = sqlc.arg('source_id')::uuid
ON CONFLICT (workspace_id, user_id) DO NOTHING;

-- name: MergeCopyProjectMembers :execrows
INSERT INTO project_members (project_id, user_id, role)
SELECT project_id, sqlc.arg('target_id')::uuid, role
FROM project_members
WHERE user_id = sqlc.arg('source_id')::uuid
ON CONFLICT (project_id, user_id) DO UPDATE
SET role = EXCLUDED.role
WHERE project_members.role < EXCLUDED.role;

-- name: MergeCopyTaskAssignees :execrows
INSERT INTO task_assignees (task_id, user_id, assigned_by)
SELECT task_id, sqlc.arg('target_id')::uuid, assigned_by
FROM task_assignees
WHERE user_id = sqlc.arg('source_id')::uuid
ON CONFLICT (task_id, user_id) DO NOTHING;

-- name: MergeCopyModuleMembers :execrows
INSERT INTO module_members (module_id, user_id, added_by)
SELECT module_id, sqlc.arg('target_id')::uuid, added_by
FROM module_members
WHERE user_id = sqlc.arg('source_id')::uuid
ON CONFLICT (module_id, user_id) DO NOTHING;

-- name: MergeFollowSourceTasks :execrows
INSERT INTO task_followers (task_id, user_id)
SELECT task_id, sqlc.arg('target_id')::uuid FROM (
  SELECT id AS task_id FROM tasks
  WHERE created_by = sqlc.arg('source_id')::uuid AND deleted_at IS NULL
  UNION
  SELECT task_id FROM task_followers
  WHERE user_id = sqlc.arg('source_id')::uuid
  UNION
  SELECT task_id FROM comments
  WHERE created_by = sqlc.arg('source_id')::uuid AND deleted_at IS NULL
) src
ON CONFLICT (task_id, user_id) DO NOTHING;

-- name: MergeCopyPrivateViews :execrows
INSERT INTO project_views (
    project_id, slug, name, description, visibility, owner_id,
    filter_tree, group_by, sort_by, sort_dir, default_tab, position
)
SELECT project_id,
       LEFT(slug, 55) || '-' || LEFT(REPLACE(uuid_generate_v4()::text, '-', ''), 8),
       name, description, visibility, sqlc.arg('target_id')::uuid,
       filter_tree, group_by, sort_by, sort_dir, default_tab, position
FROM project_views
WHERE owner_id = sqlc.arg('source_id')::uuid
  AND visibility = 'private'
  AND deleted_at IS NULL;
