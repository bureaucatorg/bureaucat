-- Feedback: anonymous messages received from any Bureaucat instance.
-- Only the main bureaucat.org instance is expected to populate this; self-hosted
-- instances typically leave it empty and disable receiving altogether.

CREATE TABLE feedback (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message         TEXT NOT NULL,
    source_origin   VARCHAR(255),
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_feedback_created_at ON feedback(created_at DESC);
