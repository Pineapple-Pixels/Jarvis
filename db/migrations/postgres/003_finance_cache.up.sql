-- 003_finance_cache

CREATE TABLE IF NOT EXISTS expenses (
    id          SERIAL PRIMARY KEY,
    date        DATE    NOT NULL,
    description TEXT    NOT NULL,
    category    TEXT    NOT NULL,
    amount      NUMERIC(12,2) NOT NULL DEFAULT 0,
    amount_usd  NUMERIC(12,2) NOT NULL DEFAULT 0,
    paid_by     TEXT    NOT NULL DEFAULT '',
    synced      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_expenses_date
    ON expenses(date DESC);

CREATE INDEX IF NOT EXISTS idx_expenses_category
    ON expenses(category);

CREATE INDEX IF NOT EXISTS idx_expenses_paid_by
    ON expenses(paid_by);

-- Composite index for monthly summaries: WHERE date >= X AND date < Y GROUP BY category
CREATE INDEX IF NOT EXISTS idx_expenses_date_category
    ON expenses(date, category);
