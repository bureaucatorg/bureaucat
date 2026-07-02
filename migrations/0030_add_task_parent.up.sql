-- Add self-referential parent for subtasks (one level of nesting).
-- Soft-deleted parents cascade to children explicitly in the handler; this
-- CASCADE only fires on a hard delete of the parent row.
ALTER TABLE tasks
    ADD COLUMN parent_task_id UUID REFERENCES tasks(id) ON DELETE CASCADE;

CREATE INDEX idx_tasks_parent_task_id ON tasks(parent_task_id) WHERE parent_task_id IS NOT NULL;
