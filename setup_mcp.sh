#!/bin/bash

# Script to setup MCP configuration for blockchain/EVM development
# This script will backup the global config and create a project-specific one

echo "🔧 Setting up MCP configuration for blockchain development..."

# Backup global MCP config if it exists
GLOBAL_CONFIG="/root/.codeium/windsurf/mcp_config.json"
if [ -f "$GLOBAL_CONFIG" ]; then
    echo "📦 Backing up global MCP config..."
    cp "$GLOBAL_CONFIG" "$GLOBAL_CONFIG.backup.$(date +%Y%m%d_%H%M%S)"
    echo "✅ Global config backed up to: $GLOBAL_CONFIG.backup.$(date +%Y%m%d_%H%M%S)"
    
    # Rename global config to disable it
    mv "$GLOBAL_CONFIG" "$GLOBAL_CONFIG.disabled"
    echo "🚫 Global config disabled (renamed to .disabled)"
else
    echo "ℹ️  No global MCP config found"
fi

# Create project-specific MCP configuration
echo "🚀 Creating comprehensive blockchain MCP configuration..."

cat > .windsurfrc << 'EOF'
{
  "mcpServers": {
    "OpenZeppelinSolidityContracts": {
      "serverUrl": "https://mcp.openzeppelin.com/contracts/solidity/mcp"
    },
    "memory": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-memory"
      ],
      "env": {
        "MEMORY_FILE_PATH": "/root/Paxeer_Project_1/memory.json"
      }
    },
    "sequential-thinking": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-sequential-thinking"
      ],
      "env": {}
    },
    "blockchain-explorer": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-blockchain-explorer"
      ],
      "env": {
        "SUPPORTED_NETWORKS": "ethereum,sepolia,polygon,arbitrum,optimism,base,paxeer",
        "PAXEER_RPC_URL": "https://v1-api.paxeer.app/rpc",
        "PAXEER_CHAIN_ID": "80000"
      }
    },
    "solidity-analyzer": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-solidity-analyzer"
      ],
      "env": {
        "SOLC_VERSION": "0.8.27",
        "OPTIMIZER_RUNS": "200",
        "EVM_VERSION": "paris"
      }
    },
    "web3-tools": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-web3"
      ],
      "env": {
        "DEFAULT_NETWORK": "ethereum",
        "INFURA_PROJECT_ID": "",
        "ALCHEMY_API_KEY": ""
      }
    },
    "ethereum-devtools": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-ethereum"
      ],
      "env": {
        "NETWORK": "mainnet",
        "PROVIDER_URL": "https://eth-mainnet.g.alchemy.com/v2/"
      }
    },
    "defi-protocols": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-defi"
      ],
      "env": {
        "SUPPORTED_PROTOCOLS": "uniswap,aave,compound,curve,1inch",
        "DEFAULT_NETWORK": "ethereum"
      }
    },
    "nft-tools": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-nft"
      ],
      "env": {
        "OPENSEA_API_KEY": "",
        "MORALIS_API_KEY": ""
      }
    },
    "gas-tracker": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-gas-tracker"
      ],
      "env": {
        "NETWORKS": "ethereum,polygon,arbitrum,optimism,base",
        "UPDATE_INTERVAL": "30000"
      }
    },
    "contract-verifier": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-etherscan"
      ],
      "env": {
        "ETHERSCAN_API_KEY": "",
        "POLYGONSCAN_API_KEY": "",
        "ARBISCAN_API_KEY": ""
      }
    }
  }
}
EOF

echo "✅ Created comprehensive .windsurfrc with blockchain/EVM MCP servers"

# Remove the old enhanced config since we're replacing it
if [ -f ".windsurfrc-enhanced" ]; then
    mv ".windsurfrc-enhanced" ".windsurfrc-enhanced.backup"
    echo "📦 Backed up old .windsurfrc-enhanced"
fi

echo ""
echo "🎉 MCP Setup Complete!"
echo ""
echo "📋 Configured MCP Servers:"
echo "  ✅ OpenZeppelin Solidity Contracts (Remote)"
echo "  ✅ Memory Server"
echo "  ✅ Sequential Thinking"
echo "  ✅ Blockchain Explorer (Multi-chain)"
echo "  ✅ Solidity Analyzer"
echo "  ✅ Web3 Tools"
echo "  ✅ Ethereum DevTools"
echo "  ✅ DeFi Protocols"
echo "  ✅ NFT Tools"
echo "  ✅ Gas Tracker"
echo "  ✅ Contract Verifier"
echo ""
echo "🔑 API Keys needed (optional):"
echo "  - INFURA_PROJECT_ID"
echo "  - ALCHEMY_API_KEY"
echo "  - OPENSEA_API_KEY"
echo "  - MORALIS_API_KEY"
echo "  - ETHERSCAN_API_KEY"
echo "  - POLYGONSCAN_API_KEY"
echo "  - ARBISCAN_API_KEY"
echo ""
echo "🚀 Next steps:"
echo "  1. Restart Windsurf IDE"
echo "  2. Add your API keys to .windsurfrc if needed"
echo "  3. Test the MCP servers"
echo ""
echo "💡 To restore global config: mv $GLOBAL_CONFIG.disabled $GLOBAL_CONFIG"
