# Paxeer Off-Chain Server - API Documentation

## Overview

The Paxeer Off-Chain Server provides a REST API for accessing indexed blockchain data from the Paxeer Launchpad AMM pools. It continuously monitors the blockchain for events and maintains a PostgreSQL database with structured data for efficient querying.

## Base Configuration

### Production Deployment
```bash
# Environment Variables
PAXEER_RPC_HTTP=https://v1-api.paxeer.app/rpc
PAXEER_RPC_WS=wss://v1-api.paxeer.app/rpc
PAXEER_FACTORY=0xFB4E790C9f047c96a53eFf08b9F58E96E6730c6a
PAXEER_DB_DSN=postgres://user:pass@localhost:5432/paxeer?sslmode=disable
PAXEER_API_ADDR=:8080
PAXEER_START_BLOCK=1
PAXEER_CONFIRMATIONS=2
PAXEER_BATCH_SIZE=5000
```

### Service Architecture
- **Indexer Process**: Monitors blockchain, processes events, updates database
- **API Process**: Serves HTTP REST endpoints from database
- **Database**: PostgreSQL with optimized schema for time-series data

---

## REST API Endpoints

### Base URL
```
Production: https://api.paxeer.com
Development: http://localhost:8080
```

### Authentication
Currently no authentication required. Consider implementing API keys or rate limiting for production.

---

## Endpoint Reference

### 1. Health Check

**Endpoint**: `GET /health`

**Description**: Service health status

**Response**:
```json
{
  "ok": true
}
```

**Status Codes**:
- `200`: Service healthy
- `500`: Service unhealthy

---

### 2. List All Pools

**Endpoint**: `GET /pools`

**Description**: Returns all indexed pools with latest state snapshots

**Response**:
```json
[
  {
    "pool": "0xeeceb441803f722a23DaBe79AF2749cA2FB89D27",
    "token": "0xb75482c25d5cA9E293e8df82cF366d3c03F860C6",
    "oracle": "0x0B82DB609B3748f9bfC609D071844eae72B82367",
    "createdBlock": 169142,
    "createdTx": "0x6a437956f98c36a8757c57469fd294088e63bfb24a3b0ba2daf788fa4b930196",
    "createdTime": "2025-09-24T06:23:14Z",
    "reserveUSDC": "494975498712813715",
    "reserveToken": "999950504900014898525046020",
    "spotX18": "10000989975497",
    "floorX18": "10000000000000"
  }
]
```

**Field Descriptions**:
- `pool`: Pool contract address
- `token`: Token contract address
- `oracle`: Oracle contract address
- `createdBlock`: Block number when pool was created
- `createdTx`: Transaction hash of pool creation
- `createdTime`: UTC timestamp of pool creation
- `reserveUSDC`: Current USDC reserves (18 decimals, as string)
- `reserveToken`: Current token reserves (18 decimals, as string)
- `spotX18`: Current spot price (18 decimals, as string)
- `floorX18`: Floor price (18 decimals, as string)

**Notes**:
- Results ordered by `createdBlock` DESC (newest first)
- All numeric values returned as strings to preserve precision
- Empty result returns `[]`, not `null`

---

### 3. Pool State

**Endpoint**: `GET /pools/{poolAddress}/state`

**Description**: Returns current state for a specific pool

**Parameters**:
- `poolAddress`: Ethereum address of the pool (case-insensitive)

**Response**: Same format as individual pool object from `/pools`

**Status Codes**:
- `200`: Pool found
- `404`: Pool not found
- `500`: Database error

**Example**:
```bash
curl -s http://localhost:8080/pools/0xeeceb441803f722a23DaBe79AF2749cA2FB89D27/state
```

---

### 4. Price Updates History

**Endpoint**: `GET /pools/{poolAddress}/price-updates`

**Description**: Returns historical price updates for a pool

**Query Parameters**:
- `fromBlock` (optional): Start from this block number (default: 0)
- `limit` (optional): Maximum number of records (default: 200, max: 1000)

**Response**:
```json
[
  {
    "priceX18": "10000989975497",
    "floorX18": "10000000000000",
    "blockNumber": 169150,
    "txHash": "0x6a437956f98c36a8757c57469fd294088e63bfb24a3b0ba2daf788fa4b930196",
    "logIndex": 0,
    "blockTime": "2025-09-24T06:23:20Z"
  }
]
```

**Field Descriptions**:
- `priceX18`: Spot price at this update (18 decimals)
- `floorX18`: Floor price at this update (18 decimals)
- `blockNumber`: Block number of the update
- `txHash`: Transaction hash containing the update
- `logIndex`: Log index within the transaction
- `blockTime`: UTC timestamp of the block

**Notes**:
- Results ordered by `blockNumber` DESC, `logIndex` DESC (newest first)
- Use for building price charts and tracking price movements
- Price updates occur on every swap that changes the price

**Example**:
```bash
curl -s "http://localhost:8080/pools/0xeeceb441803f722a23DaBe79AF2749cA2FB89D27/price-updates?fromBlock=169000&limit=50"
```

---

### 5. Swap History

**Endpoint**: `GET /pools/{poolAddress}/swaps`

**Description**: Returns swap transaction history for a pool

**Query Parameters**:
- `limit` (optional): Maximum number of records (default: 100, max: 1000)

**Response**:
```json
[
  {
    "sender": "0x742d35Cc6634C0532925a3b8D4C9db96F728b2A4",
    "usdcToToken": true,
    "amountIn": "1000000000000000000",
    "amountOut": "99900000000000000000",
    "recipient": "0x742d35Cc6634C0532925a3b8D4C9db96F728b2A4",
    "blockNumber": 169145,
    "txHash": "0x8f2a7c9d1e3b4f5a6c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b",
    "logIndex": 1,
    "blockTime": "2025-09-24T06:23:16Z"
  }
]
```

**Field Descriptions**:
- `sender`: Address that initiated the swap
- `usdcToToken`: `true` for buy (USDC → Token), `false` for sell (Token → USDC)
- `amountIn`: Input amount (18 decimals)
- `amountOut`: Output amount (18 decimals)
- `recipient`: Address that received the output tokens
- `blockNumber`: Block number of the swap
- `txHash`: Transaction hash
- `logIndex`: Log index within the transaction
- `blockTime`: UTC timestamp of the block

**Notes**:
- Results ordered by `blockNumber` DESC, `logIndex` DESC (newest first)
- Use for trade history, volume calculations, and activity feeds
- All amounts in 18-decimal format regardless of actual token decimals

**Example**:
```bash
curl -s "http://localhost:8080/pools/0xeeceb441803f722a23DaBe79AF2749cA2FB89D27/swaps?limit=25"
```

---

### 6. Price Candles (OHLC)

**Endpoint**: `GET /pools/{poolAddress}/candles`

**Description**: Returns OHLC (Open, High, Low, Close) price data aggregated by time intervals

**Query Parameters**:
- `interval` (optional): Time bucket size - `5m`, `15m`, `1h`, `4h`, `1d` (default: `5m`)
- `limit` (optional): Maximum number of candles (default: 200, max: 1000)

**Response**:
```json
[
  {
    "bucketTime": "2025-09-24T06:20:00Z",
    "open": "10000000000000",
    "high": "10001000000000",
    "low": "10000000000000",
    "close": "10000989975497"
  }
]
```

**Field Descriptions**:
- `bucketTime`: Start time of the time bucket (UTC)
- `open`: First price in the time bucket (18 decimals)
- `high`: Highest price in the time bucket (18 decimals)
- `low`: Lowest price in the time bucket (18 decimals)
- `close`: Last price in the time bucket (18 decimals)

**Supported Intervals**:
- `5m`: 5 minutes
- `15m`: 15 minutes
- `1h`: 1 hour
- `4h`: 4 hours
- `1d`: 1 day

**Notes**:
- Results ordered by `bucketTime` ASC (oldest first) for charting
- Buckets with no price updates are omitted
- Use for price charts and technical analysis
- OHLC calculated from `PriceUpdate` events within each time bucket

**Example**:
```bash
curl -s "http://localhost:8080/pools/0xeeceb441803f722a23DaBe79AF2749cA2FB89D27/candles?interval=1h&limit=100"
```

---

## Data Models

### Database Schema Overview

The indexer maintains the following key tables:

#### pools
- Pool metadata and current state snapshots
- Updated on `PoolCreated` and state-changing events

#### price_updates
- Historical price changes from `PriceUpdate` events
- Used for price charts and OHLC calculations

#### reserves
- Reserve changes from `Sync` events
- Tracks liquidity changes over time

#### swaps
- All swap transactions from `Swap` events
- Used for trade history and volume analysis

#### liquidity_events
- Add/remove liquidity from `AddLiquidity`/`RemoveLiquidity` events
- Tracks LP activity

#### oracle_updates
- Oracle price updates from `OracleUpdate` events
- Used for TWAP calculations

#### creator_fees
- Creator fee collections from `CollectCreatorFees` events
- Tracks creator earnings

---

## Integration Examples

### JavaScript/TypeScript

```typescript
interface Pool {
  pool: string;
  token: string;
  oracle: string;
  createdBlock: number;
  createdTx: string;
  createdTime: string;
  reserveUSDC: string;
  reserveToken: string;
  spotX18: string;
  floorX18: string;
}

interface PriceUpdate {
  priceX18: string;
  floorX18: string;
  blockNumber: number;
  txHash: string;
  logIndex: number;
  blockTime: string;
}

interface Swap {
  sender: string;
  usdcToToken: boolean;
  amountIn: string;
  amountOut: string;
  recipient: string;
  blockNumber: number;
  txHash: string;
  logIndex: number;
  blockTime: string;
}

interface Candle {
  bucketTime: string;
  open: string;
  high: string;
  low: string;
  close: string;
}

class PaxeerAPI {
  constructor(private baseURL: string = 'http://localhost:8080') {}

  async getPools(): Promise<Pool[]> {
    const response = await fetch(`${this.baseURL}/pools`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
  }

  async getPoolState(poolAddress: string): Promise<Pool> {
    const response = await fetch(`${this.baseURL}/pools/${poolAddress}/state`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
  }

  async getPriceUpdates(
    poolAddress: string, 
    fromBlock: number = 0, 
    limit: number = 200
  ): Promise<PriceUpdate[]> {
    const url = `${this.baseURL}/pools/${poolAddress}/price-updates?fromBlock=${fromBlock}&limit=${limit}`;
    const response = await fetch(url);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
  }

  async getSwaps(poolAddress: string, limit: number = 100): Promise<Swap[]> {
    const url = `${this.baseURL}/pools/${poolAddress}/swaps?limit=${limit}`;
    const response = await fetch(url);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
  }

  async getCandles(
    poolAddress: string, 
    interval: string = '5m', 
    limit: number = 200
  ): Promise<Candle[]> {
    const url = `${this.baseURL}/pools/${poolAddress}/candles?interval=${interval}&limit=${limit}`;
    const response = await fetch(url);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
  }
}

// Usage example
const api = new PaxeerAPI('https://api.paxeer.com');

// Get all pools
const pools = await api.getPools();
console.log(`Found ${pools.length} pools`);

// Get price history for first pool
if (pools.length > 0) {
  const priceHistory = await api.getPriceUpdates(pools[0].pool, 0, 100);
  console.log(`${priceHistory.length} price updates`);
  
  // Get 1-hour candles
  const candles = await api.getCandles(pools[0].pool, '1h', 50);
  console.log(`${candles.length} hourly candles`);
}
```

### Python

```python
import requests
from typing import List, Dict, Optional
from dataclasses import dataclass
from datetime import datetime

@dataclass
class Pool:
    pool: str
    token: str
    oracle: str
    created_block: int
    created_tx: str
    created_time: datetime
    reserve_usdc: str
    reserve_token: str
    spot_x18: str
    floor_x18: str

class PaxeerAPI:
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.session = requests.Session()
    
    def get_pools(self) -> List[Dict]:
        response = self.session.get(f"{self.base_url}/pools")
        response.raise_for_status()
        return response.json()
    
    def get_pool_state(self, pool_address: str) -> Dict:
        response = self.session.get(f"{self.base_url}/pools/{pool_address}/state")
        response.raise_for_status()
        return response.json()
    
    def get_price_updates(self, pool_address: str, from_block: int = 0, limit: int = 200) -> List[Dict]:
        params = {"fromBlock": from_block, "limit": limit}
        response = self.session.get(
            f"{self.base_url}/pools/{pool_address}/price-updates", 
            params=params
        )
        response.raise_for_status()
        return response.json()
    
    def get_swaps(self, pool_address: str, limit: int = 100) -> List[Dict]:
        params = {"limit": limit}
        response = self.session.get(
            f"{self.base_url}/pools/{pool_address}/swaps", 
            params=params
        )
        response.raise_for_status()
        return response.json()
    
    def get_candles(self, pool_address: str, interval: str = "5m", limit: int = 200) -> List[Dict]:
        params = {"interval": interval, "limit": limit}
        response = self.session.get(
            f"{self.base_url}/pools/{pool_address}/candles", 
            params=params
        )
        response.raise_for_status()
        return response.json()

# Usage example
api = PaxeerAPI("https://api.paxeer.com")

# Get all pools
pools = api.get_pools()
print(f"Found {len(pools)} pools")

# Analyze first pool
if pools:
    pool = pools[0]
    pool_address = pool["pool"]
    
    # Get recent swaps
    swaps = api.get_swaps(pool_address, limit=50)
    
    # Calculate volume
    total_volume_usdc = sum(
        float(swap["amountIn"]) for swap in swaps 
        if swap["usdcToToken"]
    )
    
    print(f"Recent USDC volume: {total_volume_usdc / 1e18:.2f}")
    
    # Get hourly candles for chart
    candles = api.get_candles(pool_address, "1h", 24)
    for candle in candles[-5:]:  # Last 5 hours
        print(f"{candle['bucketTime']}: {float(candle['close']) / 1e18:.8f}")
```

---

## Performance & Optimization

### Caching Strategy
- **Pool list**: Cache for 30 seconds (pools don't change frequently)
- **Pool state**: Cache for 10 seconds (updates on every block)
- **Historical data**: Cache for 5 minutes (immutable once confirmed)

### Rate Limiting
Consider implementing rate limiting in production:
- 100 requests per minute per IP for general endpoints
- 1000 requests per minute per IP for authenticated users

### Database Optimization
- Indexes on `pool_address`, `block_number`, `block_time`
- Partitioning by time for large datasets
- Regular VACUUM and ANALYZE operations

### Monitoring
- Track API response times
- Monitor database connection pool usage
- Alert on indexer lag (block processing delay)

---

## Error Handling

### HTTP Status Codes
- `200`: Success
- `400`: Bad request (invalid parameters)
- `404`: Resource not found (pool doesn't exist)
- `429`: Rate limit exceeded
- `500`: Internal server error
- `503`: Service unavailable (maintenance)

### Error Response Format
```json
{
  "error": "Pool not found",
  "code": "POOL_NOT_FOUND",
  "timestamp": "2025-09-24T08:00:00Z"
}
```

### Client Error Handling
```typescript
async function safeApiCall<T>(apiCall: () => Promise<T>): Promise<T> {
  try {
    return await apiCall();
  } catch (error) {
    if (error instanceof Response) {
      const errorData = await error.json();
      throw new Error(`API Error: ${errorData.error}`);
    }
    throw error;
  }
}

// Usage
const pools = await safeApiCall(() => api.getPools());
```

---

## Deployment & Operations

### Health Monitoring
```bash
# Check API health
curl -f http://localhost:8080/health || exit 1

# Check indexer is processing blocks
tail -n 10 /var/log/paxeer/indexer.log | grep -q "scan"
```

### Backup Strategy
- Daily PostgreSQL dumps
- Retain 30 days of backups
- Test restore procedures monthly

### Scaling Considerations
- Use read replicas for API queries
- Separate indexer and API databases if needed
- Consider CDN for static responses (pool lists)

### Security
- Use HTTPS in production
- Implement API authentication for write operations
- Regular security updates for dependencies
- Network-level access controls

This documentation provides comprehensive guidance for integrating with the Paxeer Off-Chain Server API and understanding its data models and operational characteristics.
