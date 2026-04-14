-- Make password_hash nullable for SSO users who don't have passwords
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- Add SSO provider columns
ALTER TABLE users ADD COLUMN auth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN provider_user_id VARCHAR(255);

-- Unique constraint: one account per provider+subject pair
CREATE UNIQUE INDEX idx_users_provider_identity
  ON users (auth_provider, provider_user_id)
  WHERE auth_provider IS NOT NULL AND provider_user_id IS NOT NULL;
