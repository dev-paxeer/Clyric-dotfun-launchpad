-- Profiles, Comments, Auth (sessions + nonces)

BEGIN;

CREATE TABLE IF NOT EXISTS profiles (
  address        TEXT PRIMARY KEY,
  username       TEXT CHECK (char_length(username) <= 30),
  bio            TEXT CHECK (char_length(bio) <= 160),
  avatar_url     TEXT,
  banner_url     TEXT,
  website_url    TEXT,
  twitter_url    TEXT,
  telegram_url   TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS comments (
  id             BIGSERIAL PRIMARY KEY,
  pool_address   TEXT NOT NULL,
  author_address TEXT NOT NULL,
  message        TEXT NOT NULL CHECK (char_length(message) > 0 AND char_length(message) <= 280),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_comments_pool_created_at ON comments(pool_address, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_comments_author ON comments(author_address);

CREATE TABLE IF NOT EXISTS sessions (
  token         TEXT PRIMARY KEY,
  address       TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_address ON sessions(address);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS auth_nonces (
  nonce        TEXT PRIMARY KEY,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  used         BOOLEAN NOT NULL DEFAULT FALSE
);

COMMIT;
