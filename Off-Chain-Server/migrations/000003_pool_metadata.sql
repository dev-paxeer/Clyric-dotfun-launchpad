BEGIN;

CREATE TABLE IF NOT EXISTS pool_metadata (
  pool_address   TEXT PRIMARY KEY,
  token_address  TEXT NOT NULL,
  name           TEXT NOT NULL CHECK (char_length(name) <= 50),
  symbol         TEXT NOT NULL CHECK (symbol ~ '^[A-Z]{3,10}$'),
  description    TEXT CHECK (char_length(description) <= 500),
  website_url    TEXT,
  twitter_url    TEXT,
  telegram_url   TEXT,
  logo_url       TEXT,
  banner_url     TEXT,
  created_by     TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pool_metadata_token ON pool_metadata(token_address);

COMMIT;
