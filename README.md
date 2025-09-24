# Clyric Launchpad

A decentralized token launch platform with single-sided AMM pools and virtual USDC bootstrapping.

## 🚀 Overview

Clyric Launchpad enables fair token launches through an innovative single-sided AMM design:

- **Virtual USDC Bootstrap**: Each pool starts with 10,000 virtual USDC
- **Creator Token Seeding**: 1 billion tokens seeded by creator
- **Floor Price Protection**: Sells cannot go below initial floor price
- **Progressive Price Discovery**: Buys can push price up without limits
- **Creator Fee Collection**: 75% of fees go to token creators

## 📁 Project Structure

```
Paxeer_Project_1/
├── Smart-Contracts/          # Solidity contracts and deployment
│   ├── contracts/            # Smart contract source code
│   ├── scripts/              # Deployment and utility scripts
│   ├── test/                 # Contract tests
│   ├── abis/                 # Event-only ABIs for indexer
│   └── abi-dist/             # Full ABIs for frontend (generated)
├── Off-Chain-Server/         # Go indexer and REST API
│   ├── cmd/                  # Main applications (indexer, api)
│   ├── internal/             # Internal packages
│   ├── migrations/           # Database migrations
│   ├── configs/              # Configuration files
│   └── abiassets/            # Embedded ABI files
└── Frontend/                 # React/Next.js frontend (to be created)
```

## 🛠 Technology Stack

### Smart Contracts
- **Solidity 0.8.24**: Smart contract language
- **Hardhat**: Development framework
- **OpenZeppelin**: Security-audited contract libraries
- **Paxeer Network**: Custom EVM chain (Chain ID: 80000)

### Off-Chain Infrastructure
- **Go 1.22**: High-performance indexer and API
- **PostgreSQL**: Time-series data storage
- **REST API**: JSON endpoints for frontend integration
- **WebSocket**: Real-time blockchain event streaming

### Frontend (Planned)
- **React/Next.js**: Modern web framework
- **ethers.js**: Ethereum interaction library
- **TypeScript**: Type-safe development
- **TailwindCSS**: Utility-first styling

## 🏗 Architecture

### Smart Contract Layer
1. **LaunchpadFactory**: Creates new token launch pools
2. **LaunchPool**: Individual AMM pools with virtual USDC
3. **LaunchpadRouter**: Convenient multi-step operations
4. **LaunchPoolOracle**: Time-weighted average price (TWAP)

### Off-Chain Layer
1. **Indexer**: Monitors blockchain, processes events
2. **API Server**: Serves structured data via REST endpoints
3. **Database**: Optimized schema for time-series queries

### Data Flow
```
Blockchain Events → Indexer → PostgreSQL → REST API → Frontend
                      ↓
                 WebSocket Subscriptions
```

## 🚀 Quick Start

### Prerequisites
- Node.js 18+
- pnpm
- Go 1.22+
- PostgreSQL 13+
- Docker (optional)

### 1. Smart Contracts

```bash
cd Smart-Contracts
pnpm install
pnpm hardhat compile
pnpm hardhat test

# Deploy to Paxeer Network
pnpm hardhat run scripts/deploy.js --network paxeer-network

# Export ABIs for frontend
pnpm hardhat run scripts/export_abis.js
```

### 2. Off-Chain Server

```bash
cd Off-Chain-Server

# Build binaries
go mod tidy
go build -o bin/indexer ./cmd/indexer
go build -o bin/api ./cmd/api

# Configure environment
export PAXEER_RPC_HTTP="https://v1-api.paxeer.app/rpc"
export PAXEER_RPC_WS="wss://v1-api.paxeer.app/rpc"
export PAXEER_FACTORY="0xFB4E790C9f047c96a53eFf08b9F58E96E6730c6a"
export PAXEER_DB_DSN="postgres://user:pass@localhost:5432/paxeer?sslmode=disable"

# Start indexer (backfill + live)
./bin/indexer -config configs/config.yaml

# Start API server
PAXEER_API_ADDR=":8080" ./bin/api -config configs/config.yaml
```

### 3. Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up -d

# Or build individual image
docker build -t paxeer-launchpad .
docker run -p 8080:8080 paxeer-launchpad
```

## 📊 Deployed Contracts (Paxeer Network)

| Contract | Address | Description |
|----------|---------|-------------|
| LaunchpadFactory | `0xFB4E790C9f047c96a53eFf08b9F58E96E6730c6a` | Creates new pools |
| LaunchpadRouter | `0x534DfB04a1A15924daB357694647e4f957543e8F` | Multi-step operations |
| USDC Token | `0x61Be934234717c57585d5f558360aFA59F8adB56` | Virtual USDC for pools |

**Network Details:**
- Chain ID: 80000
- RPC: https://v1-api.paxeer.app/rpc
- WebSocket: wss://v1-api.paxeer.app/rpc

## 🔌 API Endpoints

Base URL: `http://localhost:8080` (development)

### Core Endpoints
- `GET /health` - Service health check
- `GET /pools` - List all pools with current state
- `GET /pools/{address}/state` - Individual pool state
- `GET /pools/{address}/price-updates` - Price history
- `GET /pools/{address}/swaps` - Trade history
- `GET /pools/{address}/candles` - OHLC price data

### Example Response
```json
{
  "pool": "0xeeceb441803f722a23DaBe79AF2749cA2FB89D27",
  "token": "0xb75482c25d5cA9E293e8df82cF366d3c03F860C6",
  "oracle": "0x0B82DB609B3748f9bfC609D071844eae72B82367",
  "createdBlock": 169142,
  "reserveUSDC": "494975498712813715",
  "reserveToken": "999950504900014898525046020",
  "spotX18": "10000989975497",
  "floorX18": "10000000000000"
}
```

## 🧪 Testing

### Smart Contracts
```bash
cd Smart-Contracts
pnpm hardhat test
pnpm hardhat coverage
```

### Off-Chain Server
```bash
cd Off-Chain-Server
go test ./...
go test -race ./...
```

### Integration Tests
```bash
# Test deployed contracts
cd Smart-Contracts
pnpm hardhat test test/postDeploy.test.js --network paxeer-network

# Test API endpoints
curl -s http://localhost:8080/health
curl -s http://localhost:8080/pools | jq .
```

## 📈 Monitoring & Operations

### Health Checks
```bash
# API health
curl -f http://localhost:8080/health

# Indexer progress
tail -f Off-Chain-Server/logs/indexer.log | grep "scan"

# Database status
psql $PAXEER_DB_DSN -c "SELECT COUNT(*) FROM pools;"
```

### Key Metrics
- **Indexer Lag**: Blocks behind current head
- **API Response Time**: P95 latency for endpoints
- **Database Size**: Growth rate of event tables
- **Error Rate**: Failed transactions and API errors

## 🔧 Configuration

### Environment Variables
```bash
# Blockchain
PAXEER_RPC_HTTP=https://v1-api.paxeer.app/rpc
PAXEER_RPC_WS=wss://v1-api.paxeer.app/rpc
PAXEER_FACTORY=0xFB4E790C9f047c96a53eFf08b9F58E96E6730c6a

# Database
PAXEER_DB_DSN=postgres://user:pass@localhost:5432/paxeer?sslmode=disable

# Indexer
PAXEER_START_BLOCK=1
PAXEER_CONFIRMATIONS=2
PAXEER_BATCH_SIZE=5000

# API
PAXEER_API_ADDR=:8080
```

### Configuration Files
- `Smart-Contracts/hardhat.config.js` - Contract deployment
- `Off-Chain-Server/configs/config.yaml` - Indexer/API settings
- `docker-compose.yml` - Container orchestration

## 🛡 Security Considerations

### Smart Contracts
- OpenZeppelin security patterns
- Reentrancy protection
- Integer overflow protection
- Access control mechanisms

### Off-Chain Infrastructure
- Input validation on all endpoints
- SQL injection prevention
- Rate limiting (recommended for production)
- HTTPS termination (production)

### Operational Security
- Environment variable management
- Database access controls
- Network security groups
- Regular dependency updates

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines
- Follow existing code style
- Add tests for new features
- Update documentation
- Ensure all tests pass

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: See individual README files in each directory
- **Issues**: GitHub Issues for bug reports and feature requests
- **Discussions**: GitHub Discussions for questions and ideas

## 🗺 Roadmap

### Phase 1: Core Platform ✅
- [x] Smart contract development
- [x] Off-chain indexer and API
- [x] Basic deployment infrastructure

### Phase 2: Frontend Development 🚧
- [ ] React/Next.js web application
- [ ] Wallet integration (MetaMask, WalletConnect)
- [ ] Pool creation and management UI
- [ ] Trading interface with charts

### Phase 3: Advanced Features 📋
- [ ] Advanced charting and analytics
- [ ] Mobile application
- [ ] Governance token and DAO
- [ ] Cross-chain bridge integration

### Phase 4: Ecosystem Growth 🌱
- [ ] Third-party integrations
- [ ] API partnerships
- [ ] Developer tools and SDKs
- [ ] Community incentive programs

---

**Built with ❤️ by the Paxeer Labs team**