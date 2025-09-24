# Paxeer Launchpad - Frontend Integration Guide

## Overview

The Paxeer Launchpad is a single-sided AMM system for token launches with virtual USDC bootstrapping. This guide covers smart contract integration, event handling, and API usage for frontend developers.

## Deployed Contracts (Paxeer Network)

```json
{
  "network": "paxeer",
  "chainId": 80000,
  "rpc": "https://v1-api.paxeer.app/rpc",
  "factory": "0xFB4E790C9f047c96a53eFf08b9F58E96E6730c6a",
  "router": "0x534DfB04a1A15924daB357694647e4f957543e8F",
  "usdc": "0x61Be934234717c57585d5f558360aFA59F8adB56"
}
```

## Core Concepts

### Virtual USDC Bootstrap
- Each pool starts with **10,000 virtual USDC** (10,000 × 10^18 wei)
- Creator seeds **1 billion tokens** (1,000,000,000 × 10^18 wei)
- Initial floor price: `floor = virtualUSDC / seededTokens = 0.00001 USDC per token`

### Price Mechanics
- **Floor enforcement**: Sells (token → USDC) cannot go below floor price
- **No ceiling**: Buys (USDC → token) can push price up indefinitely
- **Current price**: Calculated as `reserveUSDC / reserveToken` (constant product)

### Fee Structure
- **1% total fee** on all swaps
- **75% to creator** (deferred collection via `collectCreatorFees()`)
- **25% to treasury** (immediate)

---

## Smart Contract Integration

### 1. LaunchpadFactory

**Purpose**: Creates new token launch pools

**Key Functions**:
```solidity
function createPool(address token) external returns (address pool, address oracle)
```

**Events**:
```solidity
event PoolCreated(address indexed token, address pool, address oracle);
```

**Usage**:
```javascript
// Create a new pool for your token
const tx = await factory.createPool(tokenAddress);
const receipt = await tx.wait();
const event = receipt.events.find(e => e.event === 'PoolCreated');
const { token, pool, oracle } = event.args;
```

### 2. LaunchPool (Individual AMM Pool)

**Key Read Functions**:
```solidity
// Get current state
function getState() external view returns (
    uint256 reserveUSDC,
    uint256 reserveToken,
    uint256 totalSupply,
    uint256 creatorFees
);

// Get current spot price (18 decimals)
function currentPriceX18() external view returns (uint256);

// Get floor price (18 decimals)  
function floorPriceX18() external view returns (uint256);

// Quote swap amounts
function getAmountOut(uint256 amountIn, bool usdcToToken) external view returns (uint256);
function getAmountIn(uint256 amountOut, bool usdcToToken) external view returns (uint256);
```

**Key Write Functions**:
```solidity
// Swap tokens
function swap(uint256 amountIn, uint256 minAmountOut, bool usdcToToken, address to) external;

// Add liquidity (after initial seeding)
function addLiquidity(uint256 amountUSDC, uint256 amountToken, address to) external;

// Remove liquidity
function removeLiquidity(uint256 lpAmount, address to) external;

// Collect creator fees (creator only)
function collectCreatorFees() external;
```

**Critical Events**:
```solidity
event PriceUpdate(uint256 priceX18, uint256 floorX18);
event Sync(uint256 reserveUSDC, uint256 reserveToken);
event Swap(address indexed sender, uint256 amountIn, uint256 amountOut, bool usdcToToken, address indexed to);
event AddLiquidity(address indexed provider, uint256 amountUSDC, uint256 amountToken, uint256 lpMinted);
event RemoveLiquidity(address indexed provider, uint256 lpBurned, uint256 amountUSDC, uint256 amountToken);
event CollectCreatorFees(uint256 amountUSDC);
```

### 3. LaunchpadRouter

**Purpose**: Convenient multi-step operations with slippage protection

**Key Functions**:
```solidity
// Swap with deadline and slippage protection
function swapExactTokensForTokens(
    uint256 amountIn,
    uint256 amountOutMin,
    address[] calldata path, // [tokenIn, tokenOut]
    address to,
    uint256 deadline
) external;

// Add liquidity with slippage protection
function addLiquidity(
    address pool,
    uint256 amountUSDCDesired,
    uint256 amountTokenDesired,
    uint256 amountUSDCMin,
    uint256 amountTokenMin,
    address to,
    uint256 deadline
) external;
```

### 4. LaunchPoolOracle

**Purpose**: Time-weighted average price (TWAP) calculation

**Key Functions**:
```solidity
function update() external; // Updates price cumulative
function consult(uint256 timeWindow) external view returns (uint256 twap);
```

**Events**:
```solidity
event OracleUpdate(uint256 priceCumulative, uint32 timestamp);
```

---

## Frontend Implementation Examples

### 1. Connect to Paxeer Network

```javascript
// Add Paxeer network to MetaMask
const paxeerNetwork = {
  chainId: '0x13880', // 80000 in hex
  chainName: 'Paxeer Network',
  rpcUrls: ['https://v1-api.paxeer.app/rpc'],
  nativeCurrency: {
    name: 'PAX',
    symbol: 'PAX',
    decimals: 18
  }
};

await window.ethereum.request({
  method: 'wallet_addEthereumChain',
  params: [paxeerNetwork]
});
```

### 2. Pool State Display

```javascript
import { ethers } from 'ethers';

// Get pool state for UI display
async function getPoolState(poolAddress) {
  const pool = new ethers.Contract(poolAddress, LaunchPoolABI, provider);
  
  const [reserveUSDC, reserveToken, totalSupply, creatorFees] = await pool.getState();
  const currentPrice = await pool.currentPriceX18();
  const floorPrice = await pool.floorPriceX18();
  
  return {
    reserveUSDC: ethers.utils.formatUnits(reserveUSDC, 18),
    reserveToken: ethers.utils.formatUnits(reserveToken, 18),
    totalSupply: ethers.utils.formatUnits(totalSupply, 18),
    creatorFees: ethers.utils.formatUnits(creatorFees, 18),
    currentPrice: ethers.utils.formatUnits(currentPrice, 18),
    floorPrice: ethers.utils.formatUnits(floorPrice, 18),
    marketCap: parseFloat(ethers.utils.formatUnits(reserveToken, 18)) * parseFloat(ethers.utils.formatUnits(currentPrice, 18))
  };
}
```

### 3. Swap with Slippage Protection

```javascript
async function executeSwap(poolAddress, amountIn, usdcToToken, slippageBps = 100) {
  const pool = new ethers.Contract(poolAddress, LaunchPoolABI, signer);
  
  // Get quote
  const amountOut = await pool.getAmountOut(
    ethers.utils.parseUnits(amountIn, 18),
    usdcToToken
  );
  
  // Apply slippage (100 bps = 1%)
  const minAmountOut = amountOut.mul(10000 - slippageBps).div(10000);
  
  // Execute swap
  const tx = await pool.swap(
    ethers.utils.parseUnits(amountIn, 18),
    minAmountOut,
    usdcToToken,
    userAddress
  );
  
  return tx.wait();
}
```

### 4. Real-time Price Updates

```javascript
// Subscribe to price updates
function subscribeToPoolEvents(poolAddress) {
  const pool = new ethers.Contract(poolAddress, LaunchPoolABI, provider);
  
  // Price updates
  pool.on('PriceUpdate', (priceX18, floorX18, event) => {
    const price = ethers.utils.formatUnits(priceX18, 18);
    const floor = ethers.utils.formatUnits(floorX18, 18);
    
    updatePriceDisplay(price, floor);
  });
  
  // Reserve changes (liquidity events)
  pool.on('Sync', (reserveUSDC, reserveToken, event) => {
    const usdc = ethers.utils.formatUnits(reserveUSDC, 18);
    const token = ethers.utils.formatUnits(reserveToken, 18);
    
    updateLiquidityDisplay(usdc, token);
  });
  
  // Swap events for trade feed
  pool.on('Swap', (sender, amountIn, amountOut, usdcToToken, to, event) => {
    const trade = {
      sender,
      amountIn: ethers.utils.formatUnits(amountIn, 18),
      amountOut: ethers.utils.formatUnits(amountOut, 18),
      type: usdcToToken ? 'buy' : 'sell',
      timestamp: Date.now(),
      txHash: event.transactionHash
    };
    
    addToTradeHistory(trade);
  });
}
```

### 5. TWAP Calculation

```javascript
// Calculate TWAP from oracle events
async function calculateTWAP(oracleAddress, timeWindowSeconds = 3600) {
  const oracle = new ethers.Contract(oracleAddress, LaunchPoolOracleABI, provider);
  
  // Get current cumulative price
  const currentBlock = await provider.getBlockNumber();
  const currentTime = (await provider.getBlock(currentBlock)).timestamp;
  
  // Get historical cumulative price
  const historicalTime = currentTime - timeWindowSeconds;
  const events = await oracle.queryFilter(
    oracle.filters.OracleUpdate(),
    currentBlock - 1000, // Adjust range as needed
    currentBlock
  );
  
  // Find closest historical point
  const historicalEvent = events.find(e => 
    e.args.timestamp <= historicalTime
  );
  
  if (!historicalEvent) return null;
  
  // Calculate TWAP
  const priceDiff = currentCumulative.sub(historicalEvent.args.priceCumulative);
  const timeDiff = currentTime - historicalEvent.args.timestamp;
  
  const twap = priceDiff.div(timeDiff);
  return ethers.utils.formatUnits(twap, 18);
}
```

---

## Off-Chain API Integration

### Base URL
```
http://localhost:8080  # Local development
https://api.paxeer.com  # Production (replace with your domain)
```

### Endpoints

#### 1. Health Check
```
GET /health
```
**Response**:
```json
{"ok": true}
```

#### 2. List All Pools
```
GET /pools
```
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

#### 3. Pool State
```
GET /pools/{poolAddress}/state
```
**Response**: Same as individual pool object above

#### 4. Price History
```
GET /pools/{poolAddress}/price-updates?fromBlock=0&limit=200
```
**Response**:
```json
[
  {
    "priceX18": "10000989975497",
    "floorX18": "10000000000000",
    "blockNumber": 169150,
    "txHash": "0x...",
    "logIndex": 0,
    "blockTime": "2025-09-24T06:23:20Z"
  }
]
```

#### 5. Trade History
```
GET /pools/{poolAddress}/swaps?limit=100
```
**Response**:
```json
[
  {
    "sender": "0x...",
    "usdcToToken": true,
    "amountIn": "1000000000000000000",
    "amountOut": "99900000000000000000",
    "recipient": "0x...",
    "blockNumber": 169145,
    "txHash": "0x...",
    "logIndex": 0,
    "blockTime": "2025-09-24T06:23:16Z"
  }
]
```

#### 6. Price Candles (OHLC)
```
GET /pools/{poolAddress}/candles?interval=5m&limit=200
```
**Intervals**: `5m`, `15m`, `1h`, `4h`, `1d`

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

### Frontend API Usage

```javascript
class PaxeerAPI {
  constructor(baseURL = 'http://localhost:8080') {
    this.baseURL = baseURL;
  }
  
  async getPools() {
    const response = await fetch(`${this.baseURL}/pools`);
    return response.json();
  }
  
  async getPoolState(poolAddress) {
    const response = await fetch(`${this.baseURL}/pools/${poolAddress}/state`);
    return response.json();
  }
  
  async getPriceHistory(poolAddress, fromBlock = 0, limit = 200) {
    const response = await fetch(
      `${this.baseURL}/pools/${poolAddress}/price-updates?fromBlock=${fromBlock}&limit=${limit}`
    );
    return response.json();
  }
  
  async getTradeHistory(poolAddress, limit = 100) {
    const response = await fetch(`${this.baseURL}/pools/${poolAddress}/swaps?limit=${limit}`);
    return response.json();
  }
  
  async getCandles(poolAddress, interval = '5m', limit = 200) {
    const response = await fetch(
      `${this.baseURL}/pools/${poolAddress}/candles?interval=${interval}&limit=${limit}`
    );
    return response.json();
  }
}

// Usage
const api = new PaxeerAPI();
const pools = await api.getPools();
const candles = await api.getCandles(pools[0].pool, '1h', 100);
```

---

## Data Types & Precision

### Decimal Handling
- **All token amounts**: 18 decimals (use `ethers.utils.formatUnits(value, 18)`)
- **Prices**: 18 decimals (priceX18, floorX18, spotX18)
- **USDC amounts**: 18 decimals (virtual USDC, not 6 like real USDC)

### BigNumber Arithmetic
```javascript
import { BigNumber, ethers } from 'ethers';

// Safe price calculations
function calculateMarketCap(reserveToken, priceX18) {
  const supply = BigNumber.from(reserveToken);
  const price = BigNumber.from(priceX18);
  
  // marketCap = supply * price / 1e18
  const marketCap = supply.mul(price).div(ethers.constants.WeiPerEther);
  return ethers.utils.formatUnits(marketCap, 18);
}

// Price impact calculation
function calculatePriceImpact(amountIn, reserveIn, reserveOut) {
  const amountInBN = BigNumber.from(amountIn);
  const reserveInBN = BigNumber.from(reserveIn);
  const reserveOutBN = BigNumber.from(reserveOut);
  
  // Current price
  const currentPrice = reserveInBN.mul(ethers.constants.WeiPerEther).div(reserveOutBN);
  
  // Price after swap (constant product: x * y = k)
  const newReserveIn = reserveInBN.add(amountInBN);
  const newReserveOut = reserveInBN.mul(reserveOutBN).div(newReserveIn);
  const newPrice = newReserveIn.mul(ethers.constants.WeiPerEther).div(newReserveOut);
  
  // Impact percentage
  const impact = newPrice.sub(currentPrice).mul(10000).div(currentPrice);
  return impact.toNumber() / 100; // Convert to percentage
}
```

---

## Error Handling

### Common Contract Errors
```javascript
// Handle common revert reasons
try {
  const tx = await pool.swap(amountIn, minAmountOut, usdcToToken, to);
  await tx.wait();
} catch (error) {
  if (error.message.includes('INSUFFICIENT_OUTPUT_AMOUNT')) {
    throw new Error('Slippage too high - try increasing slippage tolerance');
  } else if (error.message.includes('BELOW_FLOOR')) {
    throw new Error('Cannot sell below floor price');
  } else if (error.message.includes('INSUFFICIENT_LIQUIDITY')) {
    throw new Error('Not enough liquidity for this trade size');
  } else {
    throw new Error(`Transaction failed: ${error.message}`);
  }
}
```

### API Error Handling
```javascript
async function safeApiCall(apiCall) {
  try {
    return await apiCall();
  } catch (error) {
    if (error.status === 404) {
      throw new Error('Pool not found');
    } else if (error.status >= 500) {
      throw new Error('Server error - please try again later');
    } else {
      throw new Error('API request failed');
    }
  }
}
```

---

## Performance Optimization

### 1. Batch Contract Calls
```javascript
import { Contract } from '@ethersproject/contracts';

// Use multicall for batch reads
async function getMultiplePoolStates(poolAddresses) {
  const multicall = new Contract(MULTICALL_ADDRESS, MulticallABI, provider);
  
  const calls = poolAddresses.map(addr => ({
    target: addr,
    callData: LaunchPoolInterface.encodeFunctionData('getState')
  }));
  
  const results = await multicall.aggregate(calls);
  
  return results.returnData.map(data => 
    LaunchPoolInterface.decodeFunctionResult('getState', data)
  );
}
```

### 2. Efficient Event Filtering
```javascript
// Use specific event filters to reduce data
const filter = pool.filters.Swap(userAddress); // Only user's swaps
const events = await pool.queryFilter(filter, fromBlock, toBlock);
```

### 3. API Caching
```javascript
class CachedPaxeerAPI extends PaxeerAPI {
  constructor(baseURL, cacheTTL = 30000) {
    super(baseURL);
    this.cache = new Map();
    this.cacheTTL = cacheTTL;
  }
  
  async getCachedData(key, fetcher) {
    const cached = this.cache.get(key);
    if (cached && Date.now() - cached.timestamp < this.cacheTTL) {
      return cached.data;
    }
    
    const data = await fetcher();
    this.cache.set(key, { data, timestamp: Date.now() });
    return data;
  }
  
  async getPools() {
    return this.getCachedData('pools', () => super.getPools());
  }
}
```

---

## Security Considerations

### 1. Input Validation
```javascript
function validateSwapInputs(amountIn, slippageBps) {
  if (!amountIn || parseFloat(amountIn) <= 0) {
    throw new Error('Invalid amount');
  }
  if (slippageBps < 0 || slippageBps > 5000) { // Max 50% slippage
    throw new Error('Invalid slippage tolerance');
  }
}
```

### 2. Transaction Deadlines
```javascript
// Always use deadlines for time-sensitive operations
const deadline = Math.floor(Date.now() / 1000) + 1200; // 20 minutes

await router.swapExactTokensForTokens(
  amountIn,
  amountOutMin,
  [tokenA, tokenB],
  to,
  deadline
);
```

### 3. Address Validation
```javascript
import { ethers } from 'ethers';

function validateAddress(address) {
  if (!ethers.utils.isAddress(address)) {
    throw new Error('Invalid Ethereum address');
  }
  return ethers.utils.getAddress(address); // Checksum format
}
```

This comprehensive guide should provide your frontend developers with everything they need to integrate with the Paxeer Launchpad system effectively.
