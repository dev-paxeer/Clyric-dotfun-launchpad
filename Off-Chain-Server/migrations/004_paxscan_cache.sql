-- Migration to add Paxscan cache table
CREATE TABLE IF NOT EXISTS paxscan_cache (
    token_address VARCHAR(42) PRIMARY KEY,
    name TEXT,
    symbol VARCHAR(20),
    holders_count INTEGER,
    total_supply TEXT,
    decimals VARCHAR(3),
    icon_url TEXT,
    cached_at TIMESTAMP DEFAULT NOW()
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_paxscan_cache_cached_at ON paxscan_cache(cached_at);
