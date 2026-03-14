-- 005_habits
-- Habit tracking: log daily habits and query streaks.

CREATE TABLE IF NOT EXISTS habits (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    name      TEXT     NOT NULL,
    logged_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_habits_name
    ON habits(name);

CREATE INDEX IF NOT EXISTS idx_habits_logged_at
    ON habits(logged_at DESC);
