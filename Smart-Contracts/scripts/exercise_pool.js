/* eslint-disable no-console */
// Live exercise script: uses deployed Factory/Router (deployment-summary.json)
// Optionally creates a fresh pool if TOKEN_ADDRESS env is not provided
// Performs quotes and back-and-forth swaps and prints oracle readings

const { ethers, network } = require("hardhat");
const fs = require("fs");

const USDC_ONE = 10n ** 18n; // 1 USDC (18 decimals)

async function getSummary() {
  const raw = fs.readFileSync("deployment-summary.json");
  return JSON.parse(raw.toString());
}

async function attachCore(summary) {
  const factory = await ethers.getContractAt("LaunchpadFactory", summary.factory);
  const router = await ethers.getContractAt("LaunchpadRouter", summary.router);
  const usdc = await ethers.getContractAt("IERC20", summary.usdc);
  return { factory, router, usdc };
}

async function ensurePool({ factory, router, usdc }, tokenAddr) {
  if (tokenAddr) {
    const poolAddr = await factory.getPool(tokenAddr);
    if (poolAddr !== ethers.ZeroAddress) {
      return { tokenAddr, poolAddr };
    }
    // If pool missing but token provided, require creator to approve before calling createPool
    const VIRTUAL_TOKEN = 1000000000n * 10n ** 18n;
    const token = await ethers.getContractAt("IERC20", tokenAddr);
    const [creator] = await ethers.getSigners();
    const allowance = await token.allowance(creator.address, factory.target);
    if (allowance < VIRTUAL_TOKEN) {
      console.log("Approving factory to pull 1B tokens from creator...");
      await (await token.approve(factory.target, VIRTUAL_TOKEN)).wait();
    }
    console.log("Creating pool for existing token...");
    const tx = await factory.createPool(tokenAddr);
    const rc = await tx.wait();
    const evt = rc.logs.find((l) => l.fragment?.name === "PoolCreated");
    const pool = evt.args.pool;
    return { tokenAddr, poolAddr: pool };
  }

  // No token provided -> deploy TestToken and create pool
  const TestToken = await ethers.getContractFactory("TestToken");
  const token = await TestToken.deploy(ethers.parseEther("1000000000"));
  await token.waitForDeployment();
  const VIRTUAL_TOKEN = 1000000000n * 10n ** 18n;
  console.log("Deployed TestToken:", token.target);
  await (await token.approve(factory.target, VIRTUAL_TOKEN)).wait();
  console.log("Creating pool for TestToken...");
  const tx = await factory.createPool(token.target);
  const rc = await tx.wait();
  const evt = rc.logs.find((l) => l.fragment?.name === "PoolCreated");
  const poolAddr = evt.args.pool;
  return { tokenAddr: token.target, poolAddr };
}

function formatUSDC(x) {
  return `${x} (18d)`;
}
function formatToken(x) {
  return `${x} (18d)`;
}

async function printState(pool, title) {
  const vUSDC = await pool.virtualReserveUSDC();
  const vTok = await pool.virtualReserveToken();
  const [rUSDC, rTok] = await pool.getRealReserves();
  const floor = await pool.FLOOR_PRICE_X18();
  // spot price USDC per token scaled 1e18
  const spotX18 = ((vUSDC + rUSDC) * 10n ** 18n) / (vTok + rTok);
  console.log(`\n[${title}]`);
  console.log("Virtual USDC:", vUSDC.toString(), " Virtual Token:", vTok.toString());
  console.log("Real    USDC:", rUSDC.toString(), " Real    Token:", rTok.toString());
  console.log("Spot price (USDC/token * 1e18):", spotX18.toString());
  console.log("Floor     (USDC/token * 1e18):", floor.toString());

  // Oracle
  const oracleAddr = await pool.oracle();
  const oracle = await ethers.getContractAt("LaunchPoolOracle", oracleAddr);
  const obs = await oracle.lastObservation();
  console.log("Oracle lastObservation:", {
    priceCumulative: obs.priceCumulative.toString(),
    timestamp: obs.timestamp,
  });
}

async function main() {
  const summary = await getSummary();
  const { factory, router, usdc } = await attachCore(summary);
  const [signer] = await ethers.getSigners();

  console.log("Network:", network.name);
  console.log("Deployer:", signer.address);
  console.log("Factory:", factory.target);
  console.log("Router:", router.target);
  console.log("USDC:", usdc.target);

  const ENV_TOKEN = process.env.TOKEN_ADDRESS?.trim();
  const { tokenAddr, poolAddr } = await ensurePool({ factory, router, usdc }, ENV_TOKEN);
  console.log("Token:", tokenAddr);
  console.log("Pool:", poolAddr);

  const pool = await ethers.getContractAt("LaunchPool", poolAddr);

  await printState(pool, "Initial");

  // Quotes
  const qBuy = await pool.quoteUSDCToToken(USDC_ONE);
  const qSell = await pool.quoteTokenToUSDC(10n ** 18n);
  console.log("Quote 1 USDC -> token:", qBuy.toString());
  console.log("Quote 1 token -> USDC:", qSell.toString());

  // Attempt buy via router (1 USDC)
  const usdcBal = await usdc.balanceOf(signer.address);
  if (usdcBal >= USDC_ONE) {
    const allow = await usdc.allowance(signer.address, router.target);
    if (allow < USDC_ONE) {
      console.log("Approving router to spend 1 USDC...");
      await (await usdc.approve(router.target, USDC_ONE)).wait();
    }
    console.log("Buying tokens for 1 USDC...");
    const tx1 = await router.swapExactUSDCForTokens(tokenAddr, USDC_ONE, 0, signer.address);
    const rc1 = await tx1.wait();
    console.log("Buy swap gasUsed:", rc1.gasUsed.toString());
  } else {
    console.log("Skipping buy: not enough USDC balance.");
  }

  await printState(pool, "After Buy");

  // Attempt sell via router with 50% of tokens bought (approx using balance)
  const token = await ethers.getContractAt("IERC20", tokenAddr);
  const tokenBal = await token.balanceOf(signer.address);
  if (tokenBal > 0n) {
    const sellAmt = tokenBal / 2n;
    const allowT = await token.allowance(signer.address, router.target);
    if (allowT < sellAmt) {
      console.log("Approving router to spend tokens...");
      await (await token.approve(router.target, sellAmt)).wait();
    }
    console.log("Selling", formatToken(sellAmt));
    const tx2 = await router.swapExactTokensForUSDC(tokenAddr, sellAmt, 0, signer.address);
    const rc2 = await tx2.wait();
    console.log("Sell swap gasUsed:", rc2.gasUsed.toString());
  } else {
    console.log("Skipping sell: no token balance.");
  }

  await printState(pool, "After Sell");

  console.log("Done.");
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
