SELECT id, content, tags, created_at,
    ts_rank(search_vector, plainto_tsquery('spanish', $1)) AS rank
FROM memories
WHERE search_vector @@ plainto_tsquery('spanish', $1)
ORDER BY rank DESC
LIMIT $2