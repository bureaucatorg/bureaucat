-- 'mcp': done through the MCP server with a personal access token.
ALTER TABLE activity_log DROP CONSTRAINT IF EXISTS activity_log_via_check;
ALTER TABLE activity_log ADD CONSTRAINT activity_log_via_check CHECK (via IN ('web', 'assistant', 'pat', 'workflow', 'mcp'));
