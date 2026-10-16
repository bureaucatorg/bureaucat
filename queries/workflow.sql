-- ==================== WORKFLOW ENGINE ====================

-- name: GetProjectWorkflow :one
SELECT project_id, enabled, updated_at
FROM project_workflows
WHERE project_id = $1;

-- name: UpsertProjectWorkflow :exec
INSERT INTO project_workflows (project_id)
VALUES ($1)
ON CONFLICT (project_id) DO UPDATE SET updated_at = NOW();

-- name: SetProjectWorkflowEnabled :execrows
UPDATE project_workflows
SET enabled = sqlc.arg('enabled'), updated_at = NOW()
WHERE project_id = sqlc.arg('project_id');

-- name: ListWorkflowStates :many
SELECT state_id, project_id, is_entry, pos_x, pos_y
FROM workflow_states
WHERE project_id = $1;

-- name: GetWorkflowState :one
SELECT state_id, project_id, is_entry, pos_x, pos_y
FROM workflow_states
WHERE state_id = $1;

-- name: DeleteWorkflowStates :exec
DELETE FROM workflow_states WHERE project_id = $1;

-- name: CreateWorkflowState :exec
INSERT INTO workflow_states (state_id, project_id, is_entry, pos_x, pos_y)
VALUES (sqlc.arg('state_id'), sqlc.arg('project_id'), sqlc.arg('is_entry'), sqlc.arg('pos_x'), sqlc.arg('pos_y'));

-- name: ListWorkflowTransitions :many
SELECT id, project_id, from_state_id, to_state_id, allow_assignees, allow_creator
FROM workflow_transitions
WHERE project_id = $1;

-- name: GetWorkflowTransition :one
SELECT id, project_id, from_state_id, to_state_id, allow_assignees, allow_creator
FROM workflow_transitions
WHERE from_state_id = sqlc.arg('from_state_id') AND to_state_id = sqlc.arg('to_state_id');

-- name: DeleteWorkflowTransitions :exec
DELETE FROM workflow_transitions WHERE project_id = $1;

-- name: CreateWorkflowTransition :one
INSERT INTO workflow_transitions (project_id, from_state_id, to_state_id, allow_assignees, allow_creator)
VALUES (sqlc.arg('project_id'), sqlc.arg('from_state_id'), sqlc.arg('to_state_id'), sqlc.arg('allow_assignees'), sqlc.arg('allow_creator'))
RETURNING id;

-- name: ListWorkflowTransitionMovers :many
SELECT m.transition_id, m.user_id
FROM workflow_transition_movers m
JOIN workflow_transitions t ON t.id = m.transition_id
WHERE t.project_id = $1;

-- name: ListTransitionMoverIDs :many
SELECT user_id FROM workflow_transition_movers WHERE transition_id = $1;

-- name: AddWorkflowTransitionMover :exec
INSERT INTO workflow_transition_movers (transition_id, user_id)
VALUES (sqlc.arg('transition_id'), sqlc.arg('user_id'))
ON CONFLICT DO NOTHING;

-- name: ListWorkflowStateActions :many
SELECT a.state_id, a.action, a.user_id
FROM workflow_state_actions a
JOIN workflow_states s ON s.state_id = a.state_id
WHERE s.project_id = $1;

-- name: ListStateActions :many
SELECT action, user_id FROM workflow_state_actions WHERE state_id = $1;

-- name: AddWorkflowStateAction :exec
INSERT INTO workflow_state_actions (state_id, action, user_id)
VALUES (sqlc.arg('state_id'), sqlc.arg('action'), sqlc.arg('user_id'))
ON CONFLICT DO NOTHING;
