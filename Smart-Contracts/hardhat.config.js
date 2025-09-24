require("@nomicfoundation/hardhat-toolbox");
require("@openzeppelin/hardhat-upgrades");
require("dotenv/config"); // Import and configure dotenv

// Retrieve the private key and API keys from the .env file
const privateKey = process.env.PRIVATE_KEY;
const etherscanApiKey = process.env.ETHERSCAN_API_KEY;
const basescanApiKey = process.env.BASESCAN_API_KEY;

// Check if the private key is set
if (!privateKey) {
  console.warn("🚨 WARNING: PRIVATE_KEY is not set in the .env file. Deployments will not be possible.");
}

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    compilers: [
      {
        version: "0.8.20",
        settings: {
          optimizer: {
            enabled: true,
            runs: 200,
          },
          viaIR: true, // Enable IR-based code generator to fix "Stack too deep" errors
        },
      },
      {
        version: "0.8.21",
        settings: {
          optimizer: {
            enabled: true,
            runs: 200,
          },
          viaIR: true,
        },
      },
      {
        version: "0.8.27",
        settings: {
          optimizer: {
            enabled: true,
            runs: 200,
          },
          viaIR: true,
        },
      }
    ]
  },
  networks: {
    'paxeer-network': {
      url: 'https://v1-api.paxeer.app/rpc',
      chainId: 80000,
      accounts: privateKey ? [privateKey] : [],
      gasPrice: 20000000000, // 20 gwei
      gas: 8000000,
    },
    paxeer: {
      url: 'https://v1-api.paxeer.app/rpc',
      chainId: 80000,
      accounts: privateKey ? [privateKey] : [],
      gasPrice: 20000000000, // 20 gwei
      gas: 8000000,
    },
  },
  etherscan: {
    apiKey: {
      'paxeer-network': 'empty'
    },
    customChains: [
      {
        network: "paxeer-network",
        chainId: 80000,
        urls: {
          apiURL: "https://paxscan.paxeer.app:443/api/v2",
          browserURL: "https://paxscan.paxeer.app:443"
        }
      }
    ]
  },
  sourcify: {
    enabled: false,
  }
};