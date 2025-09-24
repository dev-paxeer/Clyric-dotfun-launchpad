/* eslint-disable no-console */
const { ethers, network } = require("hardhat");
const fs = require("fs");

async function main() {
  const summary = JSON.parse(fs.readFileSync("deployment-summary.json"));
  const [deployer] = await ethers.getSigners();

  const factory = await ethers.getContractAt("LaunchpadFactory", summary.factory);
  const router = await ethers.getContractAt("LaunchpadRouter", summary.router);
  const usdc = await ethers.getContractAt("IERC20", summary.usdc);

  console.log("Network:", network.name);
  console.log("Deployer:", deployer.address);
  console.log("Factory:", factory.target);
  console.log("Router:", router.target);
  console.log("USDC:", usdc.target);

  // Deploy a fresh test token and create a pool
  const TestToken = await ethers.getContractFactory("TestToken");
  const token = await TestToken.deploy(ethers.parseEther("1000000000"));
  await token.waitForDeployment();
  console.log("TestToken:", token.target);

  const VIRTUAL_TOKEN = 1000000000n * 10n ** 18n;
  await (await token.approve(factory.target, VIRTUAL_TOKEN)).wait();
  const tx = await factory.createPool(token.target);
  const rc = await tx.wait();
  const evt = rc.logs.find((l) => l.fragment?.name === "PoolCreated");
  const poolAddr = evt.args.pool;
  console.log("Pool:", poolAddr);

  const pool = await ethers.getContractAt("LaunchPool", poolAddr);
  const [rUSDC0, rTok0] = await pool.getRealReserves();
  console.log("Pool reserves (USDC,Token):", rUSDC0.toString(), rTok0.toString());

  const balDeployUSDC = await usdc.balanceOf(deployer.address);
  const allowRouter = await usdc.allowance(deployer.address, router.target);
  console.log("Deployer USDC balance:", balDeployUSDC.toString());
  console.log("Deployer->Router USDC allowance:", allowRouter.toString());

  const amount = 10n ** 18n; // 1 USDC (18 decimals)

  try {
    console.log("\nTrying router path...");
    if (allowRouter < amount) {
      await (await usdc.approve(router.target, amount)).wait();
    }
    const tx1 = await router.swapExactUSDCForTokens(token.target, amount, 0, deployer.address);
    const rc1 = await tx1.wait();
    console.log("Router swap succeeded. gasUsed=", rc1.gasUsed.toString());
  } catch (e) {
    console.error("Router swap failed:", e.message || e);
  }

  try {
    console.log("\nTrying direct pre-transfer path...");
    await (await usdc.transfer(poolAddr, amount)).wait();
    const tx2 = await pool.swapExactUSDCForTokens(amount, 0, deployer.address);
    const rc2 = await tx2.wait();
    console.log("Direct swap succeeded. gasUsed=", rc2.gasUsed.toString());
  } catch (e) {
    console.error("Direct swap failed:", e.message || e);
  }

  const [rUSDC1, rTok1] = await pool.getRealReserves();
  console.log("Pool reserves after (USDC,Token):", rUSDC1.toString(), rTok1.toString());
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
