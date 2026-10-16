-- Files the user attached to a chat message: [{upload_id, filename, mime_type, size_bytes}].
ALTER TABLE assistant_messages ADD COLUMN attachments JSONB NOT NULL DEFAULT '[]';
