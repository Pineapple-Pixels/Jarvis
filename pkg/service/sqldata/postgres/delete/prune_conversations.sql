DELETE FROM conversations
WHERE session_id IN (
    SELECT session_id
    FROM conversations
    GROUP BY session_id
    HAVING MAX(created_at) < NOW() - CAST($1 || ' days' AS INTERVAL)
);
