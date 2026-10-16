-- 'workflow': written by a workflow rule; actor_id is the user whose move triggered it.
ALTER TABLE activity_log DROP CONSTRAINT IF EXISTS activity_log_via_check;
ALTER TABLE activity_log ADD CONSTRAINT activity_log_via_check CHECK (via IN ('web', 'assistant', 'pat', 'workflow'));

-- Notifications keep the via of their latest activity so rule-made changes can show as "System".
ALTER TABLE notifications ADD COLUMN via TEXT NOT NULL DEFAULT 'web';
