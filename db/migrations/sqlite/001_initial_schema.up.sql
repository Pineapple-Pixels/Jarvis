-- 001_initial_schema
-- Core tables for the asistente personal assistant.

-------------------------------------------------------
-- MEMORIES
-- Stores notes, facts, and any user-saved content
-- with vector embeddings for semantic search.
-------------------------------------------------------

CREATE TABLE IF NOT EXISTS memories (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    content    TEXT    NOT NULL,
    tags       TEXT    DEFAULT '[]',   -- JSON array of strings
    embedding  TEXT    DEFAULT '[]',   -- JSON array of float64 (vector)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_memories_created
    ON memories(created_at DESC);

-------------------------------------------------------
-- MEMORIES FTS5
-- Full-text search virtual table kept in sync with
-- memories via triggers. Enables keyword search
-- alongside vector similarity.
-------------------------------------------------------

CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
    content,
    tags,
    content=memories,
    content_rowid=id
);

-- Keep FTS in sync on INSERT
CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON memories
BEGIN
    INSERT INTO memories_fts(rowid, content, tags)
    VALUES (new.id, new.content, new.tags);
END;

-- Keep FTS in sync on DELETE
CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON memories
BEGIN
    INSERT INTO memories_fts(memories_fts, rowid, content, tags)
    VALUES ('delete', old.id, old.content, old.tags);
END;

-- Keep FTS in sync on UPDATE
CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON memories
BEGIN
    INSERT INTO memories_fts(memories_fts, rowid, content, tags)
    VALUES ('delete', old.id, old.content, old.tags);
    INSERT INTO memories_fts(rowid, content, tags)
    VALUES (new.id, new.content, new.tags);
END;

-------------------------------------------------------
-- CONVERSATIONS
-- Stores chat history per session. Used by the
-- Context Engine for assembling prompts and compaction.
-------------------------------------------------------

CREATE TABLE IF NOT EXISTS conversations (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    role       TEXT NOT NULL,          -- 'user' | 'assistant' | 'system'
    content    TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_conversations_session
    ON conversations(session_id, created_at);
