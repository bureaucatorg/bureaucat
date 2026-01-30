-- Settings table for application-wide configuration
CREATE TABLE settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default branding settings
INSERT INTO settings (key, value) VALUES
    ('branding', '{"enabled": false, "app_name": "Bureaucat"}');
