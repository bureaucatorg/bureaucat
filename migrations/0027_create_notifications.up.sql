-- Per-user notifications with read/unread state and write-time coalescing.
-- Multiple activities on the same task within a short window collapse into a
-- single notification row (see queries/notifications.sql) so users are not spammed.
CREATE TABLE notifications (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipient_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id       UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    activity_type activity_type NOT NULL,              -- latest event type in the batch
    actor_id      UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT, -- most recent actor
    event_count   INT NOT NULL DEFAULT 1,
    read_at       TIMESTAMPTZ,                          -- NULL = unread
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- list/badge: a recipient's notifications, newest first
CREATE INDEX idx_notifications_recipient ON notifications(recipient_id, created_at DESC);
-- fast unread count for the bell badge
CREATE INDEX idx_notifications_unread ON notifications(recipient_id) WHERE read_at IS NULL;
-- coalescing lookup: most recent open row per (recipient, task)
CREATE INDEX idx_notifications_coalesce ON notifications(recipient_id, task_id, created_at DESC);
