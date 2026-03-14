SELECT m.id, m.content, m.tags, m.created_at, bm25(memories_fts) AS rank
FROM memories_fts fts
JOIN memories m ON m.id = fts.rowid
WHERE memories_fts MATCH ?
ORDER BY rank
LIMIT ?