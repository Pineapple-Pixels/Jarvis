DELETE FROM conversations
WHERE session_id IN (
    SELECT DISTINCT session_id
    FROM conversations
    GROUP BY session_id
    HAVING MAX(created_at) < datetime('now', ? || ' days')
);
