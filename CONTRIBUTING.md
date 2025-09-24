# Contributing to Paxeer Launchpad

Thank you for your interest in contributing to Paxeer Launchpad! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites
- Node.js 18+ and pnpm
- Go 1.22+
- PostgreSQL 13+
- Git
- Docker (optional)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/paxeer/launchpad.git
   cd launchpad
   ```

2. **Install dependencies**
   ```bash
   pnpm dev:setup
   ```

3. **Set up environment variables**
   ```bash
   cp Smart-Contracts/.env.example Smart-Contracts/.env
   cp Off-Chain-Server/configs/config.yaml.example Off-Chain-Server/configs/config.yaml
   # Edit the files with your configuration
   ```

4. **Start development services**
   ```bash
   # Start PostgreSQL (or use Docker)
   docker run -d --name postgres -e POSTGRES_PASSWORD=pax -e POSTGRES_USER=pax -e POSTGRES_DB=paxeer -p 5432:5432 postgres:16

   # Start indexer and API
   pnpm dev:start
   ```

## 📋 Development Workflow

### Branch Naming Convention
- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring
- `test/description` - Test improvements

### Commit Message Format
We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(contracts): add creator fee collection mechanism
fix(indexer): handle websocket reconnection properly
docs(api): update endpoint documentation
test(contracts): add edge case tests for AMM
```

## 🧪 Testing Guidelines

### Smart Contracts
```bash
cd Smart-Contracts
pnpm hardhat test
pnpm hardhat coverage
```

**Test Requirements:**
- Unit tests for all public functions
- Edge case testing (zero amounts, overflows, etc.)
- Gas usage optimization tests
- Integration tests for multi-contract interactions

### Off-Chain Server
```bash
cd Off-Chain-Server
go test ./...
go test -race ./...
go test -bench=.
```

**Test Requirements:**
- Unit tests with >80% coverage
- Integration tests with real database
- Benchmark tests for performance-critical code
- Error handling and edge case tests

### End-to-End Testing
```bash
# Deploy contracts to testnet
pnpm deploy:contracts

# Test API endpoints
curl -s http://localhost:8080/health
curl -s http://localhost:8080/pools | jq .
```

## 📝 Code Style Guidelines

### Smart Contracts (Solidity)
- Follow [Solidity Style Guide](https://docs.soliditylang.org/en/latest/style-guide.html)
- Use NatSpec comments for all public functions
- Prefer explicit over implicit (e.g., `uint256` over `uint`)
- Use descriptive variable names
- Add comprehensive error messages

**Example:**
```solidity
/**
 * @notice Swaps tokens in the AMM pool
 * @param amountIn The amount of input tokens
 * @param minAmountOut Minimum acceptable output amount
 * @param usdcToToken Direction of swap (true = USDC to token)
 * @param to Recipient address
 * @return amountOut Actual output amount
 */
function swap(
    uint256 amountIn,
    uint256 minAmountOut,
    bool usdcToToken,
    address to
) external returns (uint256 amountOut) {
    require(amountIn > 0, "LaunchPool: ZERO_AMOUNT");
    require(to != address(0), "LaunchPool: ZERO_ADDRESS");
    // Implementation...
}
```

### Go Code
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Add package and function documentation
- Use descriptive error messages
- Prefer composition over inheritance

**Example:**
```go
// PriceUpdate represents a price change event in a pool
type PriceUpdate struct {
    PoolAddress string    `json:"poolAddress"`
    PriceX18    string    `json:"priceX18"`
    FloorX18    string    `json:"floorX18"`
    BlockNumber int64     `json:"blockNumber"`
    BlockTime   time.Time `json:"blockTime"`
}

// InsertPriceUpdate stores a price update event in the database
func (r *Repo) InsertPriceUpdate(ctx context.Context, update PriceUpdate) error {
    if update.PoolAddress == "" {
        return fmt.Errorf("pool address cannot be empty")
    }
    
    _, err := r.pool.Exec(ctx, `
        INSERT INTO price_updates (pool_address, price_x18, floor_x18, block_number, block_time)
        VALUES ($1, $2, $3, $4, $5)
    `, update.PoolAddress, update.PriceX18, update.FloorX18, update.BlockNumber, update.BlockTime)
    
    return err
}
```

### JavaScript/TypeScript
- Use Prettier for formatting
- Follow ESLint rules
- Use TypeScript for type safety
- Prefer async/await over promises
- Use descriptive variable names

## 🔍 Code Review Process

### Pull Request Requirements
1. **Description**: Clear description of changes and motivation
2. **Tests**: All new code must have tests
3. **Documentation**: Update relevant documentation
4. **No Breaking Changes**: Unless discussed and approved
5. **Performance**: Consider performance implications

### Review Checklist
- [ ] Code follows style guidelines
- [ ] Tests pass and provide adequate coverage
- [ ] Documentation is updated
- [ ] No security vulnerabilities introduced
- [ ] Performance impact is acceptable
- [ ] Breaking changes are documented

### Review Process
1. Create pull request with clear description
2. Automated tests run (CI/CD)
3. Code review by maintainers
4. Address feedback and update PR
5. Final approval and merge

## 🐛 Bug Reports

### Before Reporting
1. Check existing issues
2. Reproduce the bug
3. Test on latest version
4. Gather relevant information

### Bug Report Template
```markdown
## Bug Description
Brief description of the issue

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: [e.g., Ubuntu 22.04]
- Node.js version: [e.g., 18.17.0]
- Go version: [e.g., 1.22.0]
- Browser: [e.g., Chrome 120]

## Additional Context
Any other relevant information
```

## 💡 Feature Requests

### Feature Request Template
```markdown
## Feature Description
Clear description of the proposed feature

## Motivation
Why is this feature needed?

## Proposed Solution
How should this feature work?

## Alternatives Considered
Other approaches you've considered

## Additional Context
Any other relevant information
```

## 🔒 Security

### Security Policy
- Report security vulnerabilities privately
- Do not create public issues for security bugs
- Contact: security@paxeer.com

### Security Guidelines
- Never commit private keys or secrets
- Use environment variables for configuration
- Validate all inputs
- Follow secure coding practices
- Regular dependency updates

## 📚 Documentation

### Documentation Standards
- Keep documentation up-to-date
- Use clear, concise language
- Include code examples
- Add diagrams for complex concepts
- Test documentation examples

### Documentation Types
- **API Documentation**: OpenAPI/Swagger specs
- **Code Comments**: Inline documentation
- **README Files**: Setup and usage instructions
- **Architecture Docs**: High-level design documents
- **Tutorials**: Step-by-step guides

## 🏷 Release Process

### Versioning
We use [Semantic Versioning](https://semver.org/):
- `MAJOR.MINOR.PATCH`
- Major: Breaking changes
- Minor: New features (backward compatible)
- Patch: Bug fixes (backward compatible)

### Release Checklist
- [ ] All tests pass
- [ ] Documentation updated
- [ ] Version bumped
- [ ] Changelog updated
- [ ] Security review completed
- [ ] Performance testing done

## 🤝 Community Guidelines

### Code of Conduct
- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Follow project guidelines

### Communication Channels
- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and ideas
- **Discord**: Real-time community chat
- **Email**: security@paxeer.com for security issues

## 📞 Getting Help

### Resources
- [Project Documentation](README.md)
- [API Documentation](Off-Chain-Server/API_DOCUMENTATION.md)
- [Smart Contract Guide](Smart-Contracts/README.md)
- [AI Assistant Guide](AI_ASSISTANT_GUIDE.md)

### Support Channels
1. Check existing documentation
2. Search GitHub issues
3. Ask in GitHub Discussions
4. Join Discord community
5. Contact maintainers

Thank you for contributing to Paxeer Launchpad! 🚀
