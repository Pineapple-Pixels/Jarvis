-- 003_finance_cache
-- Local cache of expenses for querying without hitting Google Sheets.
-- Source of truth remains Sheets; this enables offline queries and summaries.

CREATE TABLE IF NOT EXISTS expenses (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    date        TEXT    NOT NULL,               -- YYYY-MM-DD
    description TEXT    NOT NULL,
    category    TEXT    NOT NULL,
    amount      REAL    NOT NULL DEFAULT 0,     -- ARS
    amount_usd  REAL    NOT NULL DEFAULT 0,     -- USD
    paid_by     TEXT    NOT NULL DEFAULT '',
    synced      INTEGER NOT NULL DEFAULT 1,     -- 1=written to Sheets, 0=pending
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_expenses_date
    ON expenses(date DESC);

CREATE INDEX IF NOT EXISTS idx_expenses_category
    ON expenses(category);

CREATE INDEX IF NOT EXISTS idx_expenses_paid_by
    ON expenses(paid_by);
