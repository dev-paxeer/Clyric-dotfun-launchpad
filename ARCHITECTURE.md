# Paxeer Launchpad - System Architecture

## 🏗 High-Level Architecture

The Paxeer Launchpad is built as a three-tier decentralized application with the following components:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │  Off-Chain      │    │  Smart          │
│   (React/Next)  │◄──►│  Server (Go)    │◄──►│  Contracts      │
│                 │    │                 │    │  (Solidity)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        │                       │                       │
        ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Browser   │    │   PostgreSQL    │    │  Paxeer Network │
│   (MetaMask)    │    │   Database      │    │  (EVM Chain)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🔗 Component Interactions

### Data Flow
1. **User Actions** → Frontend → Smart Contracts (via MetaMask)
2. **Blockchain Events** → Indexer → Database → API → Frontend
3. **Real-time Updates** → WebSocket/Polling → Frontend

### Service Dependencies
- **Frontend** depends on API Server and Smart Contracts
- **API Server** depends on Database
- **Indexer** depends on Database and Blockchain RPC
- **Smart Contracts** are autonomous on blockchain

## 📦 Smart Contract Layer

### Contract Hierarchy
```
LaunchpadFactory (Singleton)
├── Creates → LaunchPool (Multiple instances)
│   ├── Uses → MockUSDC (ERC20 token)
│   └── Creates → LaunchPoolOracle (Price oracle)
└── Works with → LaunchpadRouter (Helper contract)
```

### Key Contracts

#### LaunchpadFactory
- **Purpose**: Factory pattern for creating standardized pools
- **Key Functions**: `createPool(address token)`
- **Events**: `PoolCreated(token, pool, oracle)`
- **Storage**: Mapping of tokens to pools

#### LaunchPool
- **Purpose**: Individual AMM pool with virtual USDC
- **Key Functions**: `swap()`, `addLiquidity()`, `removeLiquidity()`
- **Events**: `Swap`, `PriceUpdate`, `Sync`, `AddLiquidity`, `RemoveLiquidity`
- **Storage**: Reserves, LP tokens, creator fees

#### LaunchPoolOracle
- **Purpose**: Time-weighted average price calculation
- **Key Functions**: `update()`, `consult()`
- **Events**: `OracleUpdate(priceCumulative, timestamp)`
- **Storage**: Price cumulative, last update timestamp

#### LaunchpadRouter
- **Purpose**: Multi-step operations with slippage protection
- **Key Functions**: `swapExactTokensForTokens()`, `addLiquidity()`
- **Features**: Deadline protection, slippage limits

### AMM Mechanics

#### Virtual USDC Bootstrap
```
Initial State:
┌─────────────────┐
│ Virtual USDC    │ = 10,000 * 10^18
│ Creator Tokens  │ = 1,000,000,000 * 10^18
│ Floor Price     │ = 0.00001 USDC per token
└─────────────────┘
```

#### Constant Product Formula
```
x * y = k (where x = USDC, y = tokens)

Price = x / y
Floor enforcement: Price >= initial_floor
```

#### Fee Distribution
```
Total Fee: 1% of swap amount
├── Creator Fee: 75% (deferred collection)
└── Treasury Fee: 25% (immediate)
```

## 🔧 Off-Chain Server Architecture

### Service Structure
```
Off-Chain-Server/
├── cmd/
│   ├── indexer/     # Blockchain event processor
│   └── api/         # REST API server
├── internal/
│   ├── config/      # Configuration management
│   ├── db/          # Database layer
│   ├── indexer/     # Event processing logic
│   └── api/         # HTTP handlers
└── migrations/      # Database schema
```

### Indexer Service

#### Event Processing Pipeline
```
Blockchain → RPC Client → Event Filter → ABI Decoder → Database
     ↓              ↓            ↓           ↓           ↓
WebSocket/HTTP → FilterLogs → UnpackLog → ProcessEvent → Insert
```

#### Processing Modes
1. **Backfill Mode**: Historical event processing
2. **Live Mode**: Real-time event subscription
3. **Polling Mode**: HTTP fallback when WebSocket unavailable

#### Confirmation Handling
```
Block N → Block N+1 → Block N+2 → Confirmed (with 2 confirmations)
   ↓         ↓           ↓
Pending → Pending → Confirmed → Database Insert
```

### API Service

#### Endpoint Categories
1. **Health**: Service status and diagnostics
2. **Pools**: Pool metadata and current state
3. **History**: Historical price and trade data
4. **Analytics**: Aggregated data (candles, volume)

#### Response Format
```json
{
  "data": [...],           // Main response data
  "meta": {               // Metadata
    "total": 100,
    "page": 1,
    "limit": 50
  },
  "timestamp": "2025-09-24T08:00:00Z"
}
```

## 🗄 Database Schema

### Entity Relationship Diagram
```
┌─────────────┐    ┌─────────────────┐    ┌─────────────────┐
│    pools    │◄──►│  price_updates  │    │     swaps       │
│             │    │                 │    │                 │
│ pool_address│    │ pool_address    │    │ pool_address    │
│ token_addr  │    │ price_x18       │    │ sender          │
│ oracle_addr │    │ floor_x18       │    │ amount_in       │
│ reserves    │    │ block_number    │    │ amount_out      │
│ spot_price  │    │ block_time      │    │ usdc_to_token   │
└─────────────┘    └─────────────────┘    └─────────────────┘
       │                    │                       │
       │                    │                       │
       ▼                    ▼                       ▼
┌─────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ liquidity_  │    │ oracle_updates  │    │ creator_fees    │
│ events      │    │                 │    │                 │
│             │    │ pool_address    │    │ pool_address    │
│ event_type  │    │ price_cumulative│    │ amount_usdc     │
│ provider    │    │ oracle_timestamp│    │ block_number    │
│ amounts     │    │ block_number    │    │ block_time      │
└─────────────┘    └─────────────────┘    └─────────────────┘
```

### Indexing Strategy
```sql
-- Primary indexes
CREATE INDEX idx_pools_created_block ON pools(created_block DESC);
CREATE INDEX idx_price_updates_pool_block ON price_updates(pool_address, block_number DESC);
CREATE INDEX idx_swaps_pool_block ON swaps(pool_address, block_number DESC);

-- Composite indexes for API queries
CREATE INDEX idx_price_updates_time ON price_updates(pool_address, block_time DESC);
CREATE INDEX idx_swaps_sender ON swaps(sender, block_number DESC);

-- Partial indexes for active pools
CREATE INDEX idx_active_pools ON pools(pool_address) WHERE reserve_usdc > 0;
```

## 🌐 Frontend Architecture (Planned)

### Technology Stack
- **Framework**: Next.js 14 with App Router
- **Styling**: TailwindCSS + Headless UI
- **State Management**: Zustand + React Query
- **Blockchain**: ethers.js + wagmi
- **Charts**: Recharts + TradingView

### Component Structure
```
Frontend/
├── app/                 # Next.js App Router
│   ├── (dashboard)/     # Dashboard layout
│   ├── pools/           # Pool pages
│   └── api/             # API routes
├── components/
│   ├── ui/              # Reusable UI components
│   ├── charts/          # Chart components
│   └── web3/            # Blockchain components
├── hooks/               # Custom React hooks
├── lib/                 # Utilities and configurations
└── types/               # TypeScript type definitions
```

### State Management
```typescript
// Global state stores
interface AppState {
  pools: PoolStore;        // Pool data and operations
  wallet: WalletStore;     // Wallet connection state
  ui: UIStore;             // UI state (modals, themes)
  settings: SettingsStore; // User preferences
}

// Real-time data flow
WebSocket → React Query → Zustand Store → Components
```

## 🔄 Data Synchronization

### Real-time Updates
```
Blockchain Event → Indexer → Database → WebSocket → Frontend
                     ↓
                 REST API ← Polling ← Frontend (fallback)
```

### Caching Strategy
```
Level 1: Browser Cache (Frontend)
Level 2: Redis Cache (API Server)
Level 3: Database Query Cache
Level 4: Blockchain RPC Cache
```

### Consistency Model
- **Eventual Consistency**: Between blockchain and database
- **Strong Consistency**: Within database transactions
- **Optimistic Updates**: Frontend UI updates

## 🚀 Deployment Architecture

### Container Strategy
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │    Indexer      │    │      API        │
│   (Next.js)     │    │   (Go Binary)   │    │   (Go Binary)   │
│                 │    │                 │    │                 │
│ nginx:alpine    │    │ alpine:latest   │    │ alpine:latest   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        └───────────────────────┼───────────────────────┘
                                │
                    ┌─────────────────┐
                    │   PostgreSQL    │
                    │   (Official)    │
                    └─────────────────┘
```

### Kubernetes Deployment
```yaml
# Simplified K8s structure
apiVersion: v1
kind: Namespace
metadata:
  name: paxeer-launchpad
---
# Database (StatefulSet)
# Indexer (Deployment)
# API (Deployment + Service)
# Frontend (Deployment + Service + Ingress)
```

### Scaling Strategy
- **Horizontal**: Multiple API server replicas
- **Vertical**: Database and indexer resource scaling
- **Caching**: Redis for frequently accessed data
- **CDN**: Static asset distribution

## 🔒 Security Architecture

### Smart Contract Security
```
┌─────────────────┐
│ Access Control  │ → onlyOwner, role-based permissions
├─────────────────┤
│ Reentrancy      │ → ReentrancyGuard on state changes
├─────────────────┤
│ Integer Safety  │ → Solidity 0.8+ overflow protection
├─────────────────┤
│ Input Validation│ → require() statements, bounds checking
└─────────────────┘
```

### Off-Chain Security
```
┌─────────────────┐
│ Input Sanitization │ → SQL injection prevention
├─────────────────┤
│ Rate Limiting   │ → API request throttling
├─────────────────┤
│ Authentication  │ → JWT tokens (optional)
├─────────────────┤
│ HTTPS/TLS       │ → Encrypted communication
└─────────────────┘
```

### Infrastructure Security
```
┌─────────────────┐
│ Network         │ → VPC, security groups, firewalls
├─────────────────┤
│ Secrets         │ → Environment variables, key management
├─────────────────┤
│ Monitoring      │ → Intrusion detection, audit logs
├─────────────────┤
│ Updates         │ → Regular dependency updates
└─────────────────┘
```

## 📊 Monitoring & Observability

### Metrics Collection
```
Application Metrics → Prometheus → Grafana Dashboard
     ↓
System Metrics → Node Exporter → Alertmanager
     ↓
Custom Metrics → Go metrics → HTTP /metrics endpoint
```

### Key Performance Indicators
- **Indexer Lag**: Blocks behind current head
- **API Latency**: P95 response times
- **Database Performance**: Query execution times
- **Error Rates**: Failed requests and transactions
- **Business Metrics**: Pool creation rate, trading volume

### Alerting Strategy
```
Critical: Service down, database unavailable
Warning: High latency, indexer lag > 10 blocks
Info: New pool created, high trading volume
```

## 🔄 Development Workflow

### CI/CD Pipeline
```
Code Push → GitHub Actions → Tests → Build → Deploy
    ↓              ↓           ↓       ↓       ↓
Lint Check → Unit Tests → Integration → Docker → K8s
    ↓              ↓           ↓       ↓       ↓
Security → Contract Tests → API Tests → Registry → Staging
```

### Environment Strategy
```
Development → Staging → Production
     ↓           ↓          ↓
Local DB → Test DB → Prod DB
     ↓           ↓          ↓
Testnet → Testnet → Mainnet
```

This architecture provides a robust, scalable foundation for the Paxeer Launchpad platform while maintaining security, performance, and developer experience.
