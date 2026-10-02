DROP INDEX IF EXISTS idx_notifications_email_pending;
ALTER TABLE notifications DROP COLUMN IF EXISTS emailed_at;
ALTER TABLE users DROP COLUMN IF EXISTS email_notifications;
