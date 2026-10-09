-- task_id is blocked by blocker_task_id. Both tasks belong to the same project.
CREATE TABLE task_blockers (
    task_id         UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    blocker_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (task_id, blocker_task_id),
    CHECK (task_id <> blocker_task_id)
);

CREATE INDEX idx_task_blockers_blocker_task_id ON task_blockers(blocker_task_id);

ALTER TYPE activity_type ADD VALUE IF NOT EXISTS 'blocker_added';
ALTER TYPE activity_type ADD VALUE IF NOT EXISTS 'blocker_removed';
