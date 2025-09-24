# Paxeer Off-Chain Server (Indexer + API)

This service ingests on-chain data from the Paxeer Launchpad AMM pools and exposes a simple REST API for your frontend. It performs:
- Backfill of historical logs from the LaunchpadFactory, LaunchPool, and LaunchPoolOracle.
- Live subscription with confirmation-depth safety.
- Persistence to PostgreSQL with convenient tables for pools, price updates, reserves, swaps, LP events, oracle updates, and creator fee claims.

It targets the on-chain design you deployed:
- Single-sided virtual USDC (18 decimals) bootstrap.
- Floor price enforced on sells.
- Rich on-chain events: `PriceUpdate`, `Sync`, `Swap`, `AddLiquidity`, `RemoveLiquidity`, `CollectCreatorFees`, `InitialTokenSeeded`.
- Oracle `OracleUpdate(priceCumulative, timestamp)` for TWAP.

---

## Quick Start

Prereqs:
- Go 1.22+
- PostgreSQL 13+
- Access to Paxeer RPC (HTTP + WebSocket)

1) Configure

- Copy and edit `configs/config.yaml` or use environment overrides:
```
PAXEER_RPC_HTTP=https://your-node
PAXEER_RPC_WS=wss://your-node/ws
PAXEER_FACTORY=0x...  # LaunchpadFactory address
PAXEER_DB_DSN=postgres://user:pass@localhost:5432/paxeer?sslmode=disable
PAXEER_START_BLOCK=0   # backfill start block
PAXEER_CONFIRMATIONS=2
PAXEER_BATCH_SIZE=5000
```

2) Build
```
# From Off-Chain-Server/
go build ./cmd/indexer

go build ./cmd/api
```

3) Run migrations and indexer
```
./indexer -config configs/config.yaml
```
This will:
- Run DB migrations from `migrations/`.
- Backfill from `startBlock` to the current safe head.
- Continue live with confirmation lag.

4) Run API server (separate terminal)
```
PAXEER_API_ADDR=:8080 ./api -config configs/config.yaml
```

---

## Configuration

File: `configs/config.yaml` (ENV vars override):
- rpc.ws: WebSocket endpoint (live subscriptions)
- rpc.http: HTTP endpoint (backfill queries)
- contracts.factory: LaunchpadFactory address
- indexer.startBlock: starting block (0 = genesis)
- indexer.confirmations: reorg safety margin
- indexer.batchSize: backfill range size per call
- postgres.dsn: DSN for PostgreSQL

---

## Database Schema (key tables)

- pools
  - pool_address (PK), token_address, oracle_address
  - created_block, created_tx, created_time
  - latest snapshot: reserve_usdc, reserve_token, spot_x18, floor_x18

- price_updates
  - pool_address (FK), price_x18, floor_x18, block_number, tx_hash, log_index, block_time

- reserves
  - pool_address (FK), reserve_usdc, reserve_token, block_number, tx_hash, log_index, block_time

- swaps
  - pool_address (FK), sender, usdc_to_token, amount_in, amount_out, recipient, block_number, tx_hash, log_index, block_time

- liquidity_events
  - pool_address (FK), event_type (add|remove), provider, amount_usdc, amount_token, lp_amount, block_number, tx_hash, log_index, block_time

- oracle_updates
  - pool_address (FK), price_cumulative, oracle_timestamp, block_number, tx_hash, log_index, block_time

- creator_fees
  - pool_address (FK), amount_usdc, block_number, tx_hash, log_index, block_time

---

## REST API

- GET `/health`
  - Returns `{ "ok": true }` if healthy.

- GET `/pools`
  - Lists pools with latest snapshot.

- GET `/pools/{pool}/state`
  - Returns current stored snapshot for the pool.

- GET `/pools/{pool}/price-updates?fromBlock=&limit=`
  - Returns recent price updates (spot and floor), newest first.

- GET `/pools/{pool}/swaps?limit=`
  - Returns recent swaps.

- GET `/pools/{pool}/candles?interval=5m|1h|1d&limit=`
  - Builds OHLC from price_updates by time bucket.

Example responses are simple JSON lists of records using NUMERIC as strings, suitable for direct BigInt/Decimal parsing in the frontend.

---

## On-Chain Event Model for Frontend

Contracts emit the following events:
- LaunchpadFactory
  - `PoolCreated(token indexed, pool, oracle)`
- LaunchPool
  - `PriceUpdate(priceX18, floorX18)` spot & floor (scaled 1e18)
  - `Sync(reserveUSDC, reserveToken)` latest real reserves
  - `Swap(sender indexed, amountIn, amountOut, usdcToToken, to indexed)`
  - `AddLiquidity(provider indexed, amountUSDC, amountToken, lpMinted)`
  - `RemoveLiquidity(provider indexed, lpBurned, amountUSDC, amountToken)`
  - `CollectCreatorFees(amountUSDC)`
  - `InitialTokenSeeded(amount)`
- LaunchPoolOracle
  - `OracleUpdate(priceCumulative, timestamp)` for TWAP

Recommended frontend strategy:
- Subscribe to `PriceUpdate` for instant spot/floor.
- Subscribe to `Sync` for reserve changes and liquidity UI.
- Use `OracleUpdate` stream to compute TWAP: `TWAP = (cum[t1]-cum[t0])/(t1-t0)`.
- Fallback views: `currentPriceX18()`, `getState()` for on-demand snapshot.

---

## Building ABIs for Frontend

From `Smart-Contracts/`:
```
pnpm hardhat compile
pnpm hardhat run scripts/export_abis.js
```
Outputs: `Smart-Contracts/abi-dist/*.json` including `abi`, `bytecode`, `deployedBytecode`, `contractName`.

For lightweight listeners (like the Go indexer), we also include event-only ABIs in `Off-Chain-Server/abis/*.json`.

---

## Production Hardening Notes

- Confirmations: Currently configurable (default 2). Increase on unstable chains.
- Reorg handling: The indexer uses safe-head scanning. For deeper reorgs, consider:
  - Marking records as `confirmed=false` until beyond confirmation depth.
  - Detecting reorgs by comparing canonical hashes per block and reconciling.
- Persistence of progress: Provide `PAXEER_START_BLOCK` on restart, or extend `indexer_state` to store last scanned block. (Easy to add.)
- Metrics: Add Prometheus counters on processed logs, API latencies, DB errors.
- Backpressure: Tune `batchSize` to your node’s capacity.

---

## Local Development

- Start Postgres (example Docker):
```
docker run --rm -e POSTGRES_PASSWORD=pass -e POSTGRES_USER=user -e POSTGRES_DB=paxeer -p 5432:5432 postgres:16
```
- Export env and run indexer & API as above.

---

## Security & Reliability

- The indexer is read-only on-chain and writes to your DB only. Provide least-privilege DB credentials.
- Avoid exposing the DB directly; run the API behind a reverse proxy (nginx) with HTTPS.
- Validate inputs on API endpoints (we restrict to hex-like pool addresses and capped limits).

---

## Directory Layout

- `cmd/indexer`: main for the indexer process
- `cmd/api`: REST server
- `internal/indexer`: ABI loader, log decoding, backfill + live subscribe, repository layer
- `internal/db`: PG connection and migration runner
- `internal/api`: HTTP handlers
- `internal/config`: YAML + ENV loader
- `abis/`: event ABIs (embedded into the binary)
- `migrations/`: SQL migrations

---

## Contact & Contributions

- Open issues or requests in your team repo.
- Suggested enhancements:
  - Store last processed block and support resume.
  - Add /pools/{pool}/oracle-updates endpoint.
  - JWT or IP-based access control for API.
