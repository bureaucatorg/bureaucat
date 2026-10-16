ALTER TABLE notifications DROP COLUMN IF EXISTS via;
UPDATE activity_log SET via = 'web' WHERE via = 'workflow';
ALTER TABLE activity_log DROP CONSTRAINT IF EXISTS activity_log_via_check;
ALTER TABLE activity_log ADD CONSTRAINT activity_log_via_check CHECK (via IN ('web', 'assistant', 'pat'));
