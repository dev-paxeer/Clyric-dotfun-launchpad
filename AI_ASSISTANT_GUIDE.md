# AI Assistant Guide - Paxeer Launchpad Project

## 🤖 Project Context for AI Assistants

This document provides comprehensive context about the Paxeer Launchpad project to help AI assistants understand the codebase structure, architecture, and development patterns.

## 📋 Project Overview

**Paxeer Launchpad** is a decentralized token launch platform featuring:
- Single-sided AMM pools with virtual USDC bootstrapping
- Floor price protection for token creators
- Progressive price discovery mechanism
- Creator fee collection system (75% to creators, 25% to treasury)

## 🏗 Architecture Summary

### Three-Tier Architecture
1. **Smart Contracts (Solidity)**: On-chain AMM logic and token management
2. **Off-Chain Server (Go)**: Blockchain indexer and REST API
3. **Frontend (React/Next.js)**: User interface (to be developed)

### Key Design Patterns
- **Factory Pattern**: LaunchpadFactory creates standardized pools
- **AMM Pattern**: Constant product formula with virtual USDC
- **Event-Driven**: Comprehensive event emission for off-chain indexing
- **Microservices**: Separate indexer and API processes

## 📁 Detailed Directory Structure

```
Paxeer_Project_1/
├── Smart-Contracts/                 # Ethereum smart contracts
│   ├── contracts/
│   │   ├── LaunchpadFactory.sol     # Main factory contract
│   │   ├── LaunchPool.sol           # Individual AMM pool
│   │   ├── LaunchpadRouter.sol      # Multi-step operations
│   │   ├── LaunchPoolOracle.sol     # TWAP price oracle
│   │   ├── MockUSDC.sol             # Test USDC token
│   │   └── interfaces/              # Contract interfaces
│   ├── scripts/
│   │   ├── deploy.js                # Deployment script
│   │   ├── exercise_pool.js         # Pool interaction demo
│   │   ├── stream_pool_events.js    # Live event monitoring
│   │   └── export_abis.js           # ABI export utility
│   ├── test/                        # Contract tests
│   ├── abis/                        # Event-only ABIs for indexer
│   ├── abi-dist/                    # Full ABIs for frontend
│   ├── hardhat.config.js            # Hardhat configuration
│   ├── package.json                 # Node.js dependencies
│   └── deployment-summary.json      # Deployed contract addresses
├── Off-Chain-Server/                # Go backend services
│   ├── cmd/
│   │   ├── indexer/main.go          # Blockchain indexer service
│   │   └── api/main.go              # REST API server
│   ├── internal/
│   │   ├── config/                  # Configuration management
│   │   ├── db/                      # Database connection & migrations
│   │   ├── indexer/                 # Event processing logic
│   │   └── api/                     # HTTP handlers
│   ├── migrations/                  # SQL database schema
│   ├── configs/config.yaml          # Default configuration
│   ├── abiassets/                   # Embedded ABI files
│   ├── go.mod                       # Go module definition
│   └── README.md                    # Server documentation
├── Frontend/                        # (To be created)
├── .gitignore                       # Git ignore patterns
├── .gitattributes                   # Git file attributes
├── .npmrc                           # NPM configuration
├── Dockerfile                       # Container build instructions
├── .dockerignore                    # Docker ignore patterns
├── LICENSE                          # MIT license
├── README.md                        # Main project documentation
└── AI_ASSISTANT_GUIDE.md           # This file
```

## 🔧 Technology Stack Details

### Smart Contracts
- **Language**: Solidity 0.8.24
- **Framework**: Hardhat with TypeScript support
- **Libraries**: OpenZeppelin (security, utilities)
- **Network**: Paxeer Network (Chain ID: 80000)
- **Package Manager**: pnpm

### Off-Chain Server
- **Language**: Go 1.22
- **Database**: PostgreSQL 13+
- **HTTP Framework**: Standard library with custom routing
- **Blockchain Client**: go-ethereum (geth)
- **Configuration**: YAML + environment variables

### Development Tools
- **Version Control**: Git with conventional commits
- **Containerization**: Docker with multi-stage builds
- **Package Management**: pnpm (Node.js), Go modules
- **Testing**: Hardhat (contracts), Go testing (backend)

## 🎯 Core Concepts

### Virtual USDC Bootstrap
```
Initial Pool State:
- Virtual USDC: 10,000 * 10^18 wei
- Creator Tokens: 1,000,000,000 * 10^18 wei
- Floor Price: 10,000 / 1,000,000,000 = 0.00001 USDC per token
```

### AMM Mechanics
- **Constant Product**: `x * y = k` where x=USDC, y=tokens
- **Floor Enforcement**: Sells cannot reduce price below initial floor
- **Fee Structure**: 1% total (0.75% creator, 0.25% treasury)
- **Price Updates**: Emitted on every swap for real-time tracking

### Event-Driven Architecture
All state changes emit events for off-chain indexing:
- `PoolCreated`: New pool deployment
- `PriceUpdate`: Price and floor changes
- `Sync`: Reserve updates
- `Swap`: Trade execution
- `AddLiquidity`/`RemoveLiquidity`: LP operations
- `CollectCreatorFees`: Fee collection
- `OracleUpdate`: TWAP updates

## 🔌 API Design Patterns

### RESTful Endpoints
```
GET /health                           # Service status
GET /pools                           # List all pools
GET /pools/{address}/state           # Pool snapshot
GET /pools/{address}/price-updates   # Price history
GET /pools/{address}/swaps          # Trade history
GET /pools/{address}/candles        # OHLC data
```

### Data Format Standards
- **Addresses**: Lowercase hex strings (0x...)
- **Numbers**: String representation for precision (BigInt compatibility)
- **Timestamps**: ISO 8601 UTC format
- **Decimals**: All amounts in 18-decimal format

### Error Handling
- HTTP status codes for different error types
- Structured JSON error responses
- Graceful degradation for missing data

## 🗄 Database Schema

### Key Tables
```sql
-- Pool metadata and current state
pools (
    pool_address PRIMARY KEY,
    token_address,
    oracle_address,
    created_block,
    created_tx,
    created_time,
    reserve_usdc,     -- Current reserves
    reserve_token,    -- Current reserves
    spot_x18,         -- Current spot price
    floor_x18         -- Floor price
);

-- Historical price updates
price_updates (
    pool_address,
    price_x18,
    floor_x18,
    block_number,
    tx_hash,
    log_index,
    block_time,
    confirmed
);

-- Trade history
swaps (
    pool_address,
    sender,
    usdc_to_token,    -- Direction: true=buy, false=sell
    amount_in,
    amount_out,
    recipient,
    block_number,
    tx_hash,
    log_index,
    block_time,
    confirmed
);
```

## 🚀 Development Workflows

### Smart Contract Development
1. Write contracts in `contracts/`
2. Add tests in `test/`
3. Deploy with `scripts/deploy.js`
4. Export ABIs with `scripts/export_abis.js`
5. Verify deployment with post-deploy tests

### Backend Development
1. Modify Go code in `internal/`
2. Update database schema in `migrations/`
3. Build with `go build`
4. Test with `go test`
5. Deploy with Docker or binary

### Integration Testing
1. Deploy contracts to testnet
2. Start indexer with deployed addresses
3. Verify API endpoints return expected data
4. Test real-time event processing

## 🔍 Common Development Tasks

### Adding New Contract Events
1. Add event to Solidity contract
2. Update ABI files in `abis/`
3. Add event signature to `internal/indexer/abi.go`
4. Implement handler in `internal/indexer/indexer.go`
5. Add database table/columns if needed
6. Update API endpoints if exposing data

### Adding New API Endpoints
1. Add route in `internal/api/server.go`
2. Implement handler function
3. Add database queries if needed
4. Update documentation
5. Add integration tests

### Database Migrations
1. Create new SQL file in `migrations/`
2. Use sequential numbering (e.g., `000002_add_table.sql`)
3. Test migration on development database
4. Update schema documentation

## 🐛 Debugging Guidelines

### Smart Contract Issues
- Use `console.log` in Hardhat tests
- Check transaction receipts for events
- Verify gas usage and limits
- Test edge cases (zero amounts, etc.)

### Indexer Issues
- Check logs for event processing errors
- Verify ABI signatures match contract events
- Monitor database for missing/duplicate data
- Test with different block ranges

### API Issues
- Check HTTP status codes and error messages
- Verify database queries return expected results
- Test with different query parameters
- Monitor response times and memory usage

## 📊 Performance Considerations

### Smart Contracts
- Gas optimization for frequently called functions
- Batch operations where possible
- Efficient storage layout
- Event emission for off-chain processing

### Off-Chain Server
- Database indexing on frequently queried columns
- Connection pooling for database access
- Batch processing for blockchain events
- Caching for frequently requested data

### API Design
- Pagination for large result sets
- Efficient SQL queries with proper indexes
- Response compression
- Rate limiting for production

## 🔒 Security Patterns

### Smart Contracts
- Reentrancy guards on state-changing functions
- Input validation and bounds checking
- Access control for administrative functions
- Safe math operations (Solidity 0.8+)

### Off-Chain Infrastructure
- Input sanitization for API endpoints
- SQL injection prevention with parameterized queries
- Environment variable management for secrets
- HTTPS termination in production

## 🧪 Testing Strategies

### Unit Tests
- Individual contract function testing
- Go package testing with mocks
- Edge case validation
- Error condition handling

### Integration Tests
- End-to-end contract interactions
- API endpoint testing with real database
- Event processing verification
- Performance testing under load

### Deployment Testing
- Testnet deployment validation
- Production deployment verification
- Rollback procedures
- Monitoring and alerting setup

## 📈 Monitoring & Observability

### Key Metrics
- **Indexer Lag**: Blocks behind current head
- **API Latency**: Response time percentiles
- **Database Performance**: Query execution times
- **Error Rates**: Failed transactions and API calls

### Logging Patterns
- Structured logging with consistent fields
- Log levels: DEBUG, INFO, WARN, ERROR
- Request tracing for API calls
- Event processing status logging

### Health Checks
- API health endpoint (`/health`)
- Database connectivity checks
- Blockchain node connectivity
- Service dependency validation

## 🔄 Deployment Patterns

### Development Environment
- Local blockchain (Hardhat Network)
- Local PostgreSQL instance
- Hot reloading for rapid development
- Comprehensive logging for debugging

### Production Environment
- Containerized services with Docker
- Load balancing for API servers
- Database replication for high availability
- Monitoring and alerting infrastructure

This guide should provide AI assistants with comprehensive context to effectively help with development, debugging, and enhancement of the Paxeer Launchpad project.
