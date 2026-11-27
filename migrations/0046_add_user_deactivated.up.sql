-- Deactivated users are fully disabled: no sign-in, sessions, PATs or Mattermost bot.
ALTER TABLE users ADD COLUMN deactivated_at TIMESTAMPTZ;
