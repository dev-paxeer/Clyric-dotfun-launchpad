/* eslint-disable no-console */
// End-to-end: (optionally) deploy LaunchpadFactory & ERC20Factory, then create a token and pool
// Usage:
//   npx hardhat run scripts/e2e_create_token_and_pool.js --network paxeer
// Env (optional):
//   NAME=MyToken SYMBOL=MYT USDC=0x... FACTORY=0x... ERC20_FACTORY=0x...

const { ethers, network } = require("hardhat");
const fs = require("fs");

const ONE_BILLION_18 = ethers.parseEther("1000000000");

async function loadSummary() {
  try {
    const raw = fs.readFileSync("deployment-summary.json");
    return JSON.parse(raw.toString());
  } catch {
    return {};
  }
}

async function saveSummary(summary) {
  fs.writeFileSync("deployment-summary.json", JSON.stringify(summary, null, 2));
}

async function ensureLaunchpadFactory(usdcAddr, summary, deployerAddr) {
  if (process.env.FACTORY) {
    summary.factory = process.env.FACTORY;
    return summary.factory;
  }
  if (summary.factory) return summary.factory;

  const LaunchpadFactory = await ethers.getContractFactory("LaunchpadFactory");
  const factory = await LaunchpadFactory.deploy(usdcAddr, deployerAddr);
  await factory.waitForDeployment();
  summary.factory = factory.target;
  console.log("LaunchpadFactory deployed:", summary.factory);
  return summary.factory;
}

async function ensureERC20Factory(summary) {
  if (process.env.ERC20_FACTORY) {
    summary.erc20Factory = process.env.ERC20_FACTORY;
    return summary.erc20Factory;
  }
  if (summary.erc20Factory) return summary.erc20Factory;

  const ERC20Factory = await ethers.getContractFactory("ERC20Factory");
  const f = await ERC20Factory.deploy();
  await f.waitForDeployment();
  summary.erc20Factory = f.target;
  console.log("ERC20Factory deployed:", summary.erc20Factory);
  return summary.erc20Factory;
}

async function main() {
  const [signer] = await ethers.getSigners();
  console.log("Network:", network.name);
  console.log("Deployer:", signer.address);

  const summary = await loadSummary();
  const USDC = process.env.USDC || summary.usdc || "0x61Be934234717c57585d5f558360aFA59F8adB56";
  if (!USDC) throw new Error("USDC address required (set USDC env or in deployment-summary.json)");

  // Ensure factories
  const factoryAddr = await ensureLaunchpadFactory(USDC, summary, signer.address);
  const erc20FactoryAddr = await ensureERC20Factory(summary);
  summary.usdc = USDC;
  await saveSummary(summary);

  // Create token
  const name = (process.env.NAME || "MyToken").trim();
  const symbol = (process.env.SYMBOL || ("MYT" + Math.floor(Math.random()*900+100))).toUpperCase();

  const erc20Factory = await ethers.getContractAt("ERC20Factory", erc20FactoryAddr);
  console.log("[1/4] Creating token:", name, symbol);
  const txCreate = await erc20Factory.createToken(name, symbol, ONE_BILLION_18);
  console.log("  tx:", txCreate.hash);
  const rcCreate = await txCreate.wait();
  const logCreated = rcCreate.logs.find((l) => l.fragment?.name === "TokenCreated");
  const tokenAddr = logCreated?.args?.token || logCreated?.args?.[0];
  if (!tokenAddr) throw new Error("Token address not found in receipt");
  console.log("  token:", tokenAddr);

  // Approve LaunchpadFactory
  const erc = await ethers.getContractAt("IERC20", tokenAddr);
  console.log("[2/4] Approving factory for seed amount (1B):", factoryAddr);
  const txApprove = await erc.approve(factoryAddr, ONE_BILLION_18);
  console.log("  tx:", txApprove.hash);
  await txApprove.wait();

  // Create pool
  const factory = await ethers.getContractAt("LaunchpadFactory", factoryAddr);
  console.log("[3/4] Creating pool for token...");
  const txPool = await factory.createPool(tokenAddr);
  console.log("  tx:", txPool.hash);
  const rcPool = await txPool.wait();
  const logPool = rcPool.logs.find((l) => l.fragment?.name === "PoolCreated");
  const poolAddr = logPool?.args?.pool || logPool?.args?.[1];
  if (!poolAddr) throw new Error("Pool address not found in receipt");
  console.log("  pool:", poolAddr);

  // Print state
  const pool = await ethers.getContractAt("LaunchPool", poolAddr);
  const vUSDC = await pool.virtualReserveUSDC();
  const [rUSDC, rToken] = await pool.getRealReserves();
  console.log("[4/4] Pool state:", { vUSDC: vUSDC.toString(), rUSDC: rUSDC.toString(), rToken: rToken.toString() });

  console.log("Done ✅", { token: tokenAddr, pool: poolAddr });
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
