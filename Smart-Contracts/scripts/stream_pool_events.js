/* eslint-disable no-console */
// Streams on-chain price and state updates for a LaunchPool in real time.
// Usage:
//   - TOKEN_ADDRESS=0xToken pnpm hardhat run scripts/stream_pool_events.js --network paxeer
//   - Or specify POOL_ADDRESS directly:
//       POOL_ADDRESS=0xPool pnpm hardhat run scripts/stream_pool_events.js --network paxeer
//   - If neither is provided, the script will watch all pools created by the Factory from logs.

const { ethers, network } = require("hardhat");
const fs = require("fs");

function fmtX18(x) {
  return x.toString(); // scaled by 1e18; convert in UI as needed
}

async function getSummary() {
  const raw = fs.readFileSync("deployment-summary.json");
  return JSON.parse(raw.toString());
}

async function getPoolsFromFactory(factory) {
  const filter = factory.filters.PoolCreated();
  const logs = await factory.queryFilter(filter, 0, "latest");
  const pools = logs.map((l) => ({ token: l.args.token, pool: l.args.pool, oracle: l.args.oracle }));
  return pools;
}

async function attachPool(poolAddr) {
  const pool = await ethers.getContractAt("LaunchPool", poolAddr);
  const oracleAddr = await pool.oracle();
  const oracle = await ethers.getContractAt("LaunchPoolOracle", oracleAddr);
  return { pool, oracle };
}

async function printSnapshot(pool, tag) {
  const [vUSDC, rUSDC, rTok, spotX18, floorX18, pendingFeesUSDC] = await pool.getState();
  console.log(`\n[${tag}]`);
  console.log("vUSDC:", vUSDC.toString());
  console.log("rUSDC:", rUSDC.toString());
  console.log("rToken:", rTok.toString());
  console.log("spotX18:", fmtX18(spotX18));
  console.log("floorX18:", fmtX18(floorX18));
  console.log("pendingCreatorFeesUSDC:", pendingFeesUSDC.toString());
}

async function main() {
  const [signer] = await ethers.getSigners();
  const summary = await getSummary();
  const factory = await ethers.getContractAt("LaunchpadFactory", summary.factory);

  console.log("Network:", network.name);
  console.log("Observer:", signer.address);
  console.log("Factory:", factory.target);

  const tokenEnv = process.env.TOKEN_ADDRESS?.trim();
  const poolEnv = process.env.POOL_ADDRESS?.trim();

  let targets = [];
  if (poolEnv && poolEnv !== "") {
    targets.push({ token: ethers.ZeroAddress, pool: poolEnv });
  } else if (tokenEnv && tokenEnv !== "") {
    const poolAddr = await factory.getPool(tokenEnv);
    if (poolAddr === ethers.ZeroAddress) {
      console.error("No pool found for token", tokenEnv);
      process.exit(1);
    }
    targets.push({ token: tokenEnv, pool: poolAddr });
  } else {
    targets = await getPoolsFromFactory(factory);
    if (targets.length === 0) {
      console.error("No pools found via Factory events. Provide TOKEN_ADDRESS or POOL_ADDRESS.");
      process.exit(1);
    }
  }

  console.log("Watching pools:", targets.map((t) => t.pool));

  for (const t of targets) {
    const { pool, oracle } = await attachPool(t.pool);
    await printSnapshot(pool, `Initial ${t.pool}`);

    // Pool events
    pool.on("PriceUpdate", async (priceX18, floorX18, ev) => {
      console.log("[PriceUpdate] pool=", pool.target, "priceX18=", fmtX18(priceX18), "floorX18=", fmtX18(floorX18), "block=", ev.blockNumber);
    });

    pool.on("Sync", async (rUSDC, rToken, ev) => {
      console.log("[Sync] pool=", pool.target, "rUSDC=", rUSDC.toString(), "rToken=", rToken.toString(), "block=", ev.blockNumber);
    });

    pool.on("Swap", async (sender, amountIn, amountOut, usdcToToken, to, ev) => {
      console.log("[Swap] pool=", pool.target, "sender=", sender, "usdcToToken=", usdcToToken, "in=", amountIn.toString(), "out=", amountOut.toString(), "to=", to, "block=", ev.blockNumber);
    });

    pool.on("AddLiquidity", async (provider, amountUSDC, amountToken, lpMinted, ev) => {
      console.log("[AddLiquidity] pool=", pool.target, "provider=", provider, "USDC=", amountUSDC.toString(), "Token=", amountToken.toString(), "LP=", lpMinted.toString(), "block=", ev.blockNumber);
    });

    pool.on("RemoveLiquidity", async (provider, lpBurned, amountUSDC, amountToken, ev) => {
      console.log("[RemoveLiquidity] pool=", pool.target, "provider=", provider, "LP=", lpBurned.toString(), "USDC=", amountUSDC.toString(), "Token=", amountToken.toString(), "block=", ev.blockNumber);
    });

    pool.on("CollectCreatorFees", async (amountUSDC, ev) => {
      console.log("[CollectCreatorFees] pool=", pool.target, "amountUSDC=", amountUSDC.toString(), "block=", ev.blockNumber);
    });

    // Oracle events
    oracle.on("OracleUpdate", (priceCumulative, timestamp) => {
      console.log("[OracleUpdate] oracle=", oracle.target, "priceCumulative=", priceCumulative.toString(), "timestamp=", timestamp);
    });
  }

  console.log("Event stream started. Press Ctrl+C to exit.");
  // Keep process alive
  await new Promise(() => {});
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
