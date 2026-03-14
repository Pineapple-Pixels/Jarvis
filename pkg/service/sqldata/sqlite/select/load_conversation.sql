SELECT role, content, created_at FROM conversations
WHERE session_id = ? ORDER BY created_at DESC LIMIT ?