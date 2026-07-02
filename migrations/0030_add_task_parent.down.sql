DROP INDEX IF EXISTS idx_tasks_parent_task_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS parent_task_id;
