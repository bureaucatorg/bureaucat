-- Optional reference to the comment that triggered the notification, so the bell
-- can deep-link to the highlighted comment (#comment-<id>) instead of just the task.
-- ON DELETE SET NULL: if the comment is ever hard-deleted, fall back to the task link.
ALTER TABLE notifications
    ADD COLUMN comment_id UUID REFERENCES comments(id) ON DELETE SET NULL;
