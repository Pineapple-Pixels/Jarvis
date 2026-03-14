SELECT role, content, created_at FROM conversations
WHERE session_id = $1 ORDER BY created_at DESC LIMIT $2