-- Initial schema for Paxeer indexer

CREATE TABLE IF NOT EXISTS schema_migrations (
  id SERIAL PRIMARY KEY,
  filename TEXT UNIQUE NOT NULL,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pools (
  pool_address TEXT PRIMARY KEY,
  token_address TEXT NOT NULL,
  oracle_address TEXT NOT NULL,
  created_block BIGINT NOT NULL,
  created_tx TEXT NOT NULL,
  created_time TIMESTAMPTZ,
  -- latest snapshot
  reserve_usdc NUMERIC,
  reserve_token NUMERIC,
  spot_x18 NUMERIC,
  floor_x18 NUMERIC
);

CREATE TABLE IF NOT EXISTS price_updates (
  id BIGSERIAL PRIMARY KEY,
  pool_address TEXT NOT NULL REFERENCES pools(pool_address) ON DELETE CASCADE,
  price_x18 NUMERIC NOT NULL,
  floor_x18 NUMERIC NOT NULL,
  block_number BIGINT NOT NULL,
  tx_hash TEXT NOT NULL,
  log_index INT NOT NULL,
  block_time TIMESTAMPTZ,
  confirmed BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_price_updates_pool_block ON price_updates(pool_address, block_number);

CREATE TABLE IF NOT EXISTS reserves (
  id BIGSERIAL PRIMARY KEY,
  pool_address TEXT NOT NULL REFERENCES pools(pool_address) ON DELETE CASCADE,
  reserve_usdc NUMERIC NOT NULL,
  reserve_token NUMERIC NOT NULL,
  block_number BIGINT NOT NULL,
  tx_hash TEXT NOT NULL,
  log_index INT NOT NULL,
  block_time TIMESTAMPTZ,
  confirmed BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_reserves_pool_block ON reserves(pool_address, block_number);

CREATE TABLE IF NOT EXISTS swaps (
  id BIGSERIAL PRIMARY KEY,
  pool_address TEXT NOT NULL REFERENCES pools(pool_address) ON DELETE CASCADE,
  sender TEXT NOT NULL,
  usdc_to_token BOOLEAN NOT NULL,
  amount_in NUMERIC NOT NULL,
  amount_out NUMERIC NOT NULL,
  recipient TEXT NOT NULL,
  block_number BIGINT NOT NULL,
  tx_hash TEXT NOT NULL,
  log_index INT NOT NULL,
  block_time TIMESTAMPTZ,
  confirmed BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_swaps_pool_block ON swaps(pool_address, block_number);

CREATE TABLE IF NOT EXISTS liquidity_events (
  id BIGSERIAL PRIMARY KEY,
  pool_address TEXT NOT NULL REFERENCES pools(pool_address) ON DELETE CASCADE,
  event_type TEXT NOT NULL CHECK (event_type IN ('add','remove')),
  provider TEXT NOT NULL,
  amount_usdc NUMERIC NOT NULL,
  amount_token NUMERIC NOT NULL,
  lp_amount NUMERIC,
  block_number BIGINT NOT NULL,
  tx_hash TEXT NOT NULL,
  log_index INT NOT NULL,
  block_time TIMESTAMPTZ,
  confirmed BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_liquidity_pool_block ON liquidity_events(pool_address, block_number);

CREATE TABLE IF NOT EXISTS oracle_updates (
  id BIGSERIAL PRIMARY KEY,
  pool_address TEXT NOT NULL REFERENCES pools(pool_address) ON DELETE CASCADE,
  price_cumulative NUMERIC NOT NULL,
  oracle_timestamp BIGINT NOT NULL,
  block_number BIGINT NOT NULL,
  tx_hash TEXT NOT NULL,
  log_index INT NOT NULL,
  block_time TIMESTAMPTZ,
  confirmed BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_oracle_pool_block ON oracle_updates(pool_address, block_number);

CREATE TABLE IF NOT EXISTS creator_fees (
  id BIGSERIAL PRIMARY KEY,
  pool_address TEXT NOT NULL REFERENCES pools(pool_address) ON DELETE CASCADE,
  amount_usdc NUMERIC NOT NULL,
  block_number BIGINT NOT NULL,
  tx_hash TEXT NOT NULL,
  log_index INT NOT NULL,
  block_time TIMESTAMPTZ,
  confirmed BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_creator_fees_pool_block ON creator_fees(pool_address, block_number);

CREATE TABLE IF NOT EXISTS indexer_state (
  id SMALLINT PRIMARY KEY DEFAULT 1,
  last_backfilled_block BIGINT NOT NULL DEFAULT 0,
  last_seen_head BIGINT NOT NULL DEFAULT 0
);
