-- Extensible per-message context from the channel adapter, e.g. {"page": {...}} for the web page the user is on.
ALTER TABLE assistant_messages ADD COLUMN meta JSONB NOT NULL DEFAULT '{}';
