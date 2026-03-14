-- 005_habits

CREATE TABLE IF NOT EXISTS habits (
    id        SERIAL PRIMARY KEY,
    name      TEXT        NOT NULL,
    logged_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_habits_name
    ON habits(name);

CREATE INDEX IF NOT EXISTS idx_habits_logged_at
    ON habits(logged_at DESC);
