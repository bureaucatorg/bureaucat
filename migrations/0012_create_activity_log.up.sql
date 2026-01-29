-- Activity type enum
CREATE TYPE activity_type AS ENUM (
    'task_created',
    'task_updated',
    'task_deleted',
    'assignee_added',
    'assignee_removed',
    'label_added',
    'label_removed',
    'state_changed',
    'comment_created',
    'comment_updated',
    'comment_deleted'
);

-- Activity log table (tamper-proof with hash chain)
CREATE TABLE activity_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    activity_type activity_type NOT NULL,
    actor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    field_name VARCHAR(100),           -- e.g., "title", "state", "assignees"
    old_value JSONB,                   -- Previous value (null for creates)
    new_value JSONB,                   -- New value (null for deletes)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    checksum VARCHAR(64) NOT NULL      -- SHA-256 of previous row + current data
);

-- Append-only: No UPDATE or DELETE allowed via application
CREATE INDEX idx_activity_log_task_id ON activity_log(task_id);
CREATE INDEX idx_activity_log_actor_id ON activity_log(actor_id);
CREATE INDEX idx_activity_log_created_at ON activity_log(created_at);
