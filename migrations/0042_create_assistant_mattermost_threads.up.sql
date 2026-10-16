-- Maps a Mattermost thread (root post) to the assistant conversation of one Bureaucat user.
CREATE TABLE assistant_mattermost_threads (
    root_post_id    TEXT NOT NULL,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES assistant_conversations(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (root_post_id, user_id)
);

CREATE INDEX idx_users_email_lower ON users (LOWER(email));
