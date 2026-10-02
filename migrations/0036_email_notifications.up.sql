-- Per-user opt-in for emailed notifications (off by default).
ALTER TABLE users ADD COLUMN email_notifications BOOLEAN NOT NULL DEFAULT false;

-- Set once the email digest worker has processed the row (sent or skipped).
ALTER TABLE notifications ADD COLUMN emailed_at TIMESTAMPTZ;

CREATE INDEX idx_notifications_email_pending ON notifications(created_at) WHERE emailed_at IS NULL;
