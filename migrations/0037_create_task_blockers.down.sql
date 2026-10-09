-- Postgres cannot drop enum values, so blocker_added/blocker_removed remain.
DROP TABLE IF EXISTS task_blockers;
