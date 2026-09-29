-- Users following a task receive its in-app notifications (see ListTaskParticipants).
CREATE TABLE task_followers (
    task_id     UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (task_id, user_id)
);

CREATE INDEX idx_task_followers_user_id ON task_followers(user_id);
