/* eslint-env mocha */
// @post-deploy
const { expect } = require("chai");
const fs = require("fs");
const { ethers } = require("hardhat");

describe("Post-deployment validation", function () {
  let summary, factory, router, usdc;

  before("load deployment summary", async function () {
    summary = JSON.parse(fs.readFileSync("deployment-summary.json"));
    factory = await ethers.getContractAt("LaunchpadFactory", summary.factory);
    router = await ethers.getContractAt("LaunchpadRouter", summary.router);
    usdc = await ethers.getContractAt("IERC20", summary.usdc);
  });

  it("factory and router addresses are set", async function () {
    expect(factory.target).to.properAddress;
    expect(router.target).to.properAddress;
  });

  it("factory stores USDC address correctly", async function () {
    expect(await factory.usdc()).to.equal(summary.usdc);
  });

  it("router points to factory", async function () {
    expect(await router.factory()).to.equal(factory.target);
  });

  it("can create a pool and perform a swap @post-deploy", async function () {
    const [deployer] = await ethers.getSigners();

    // deploy mock project token
    const TestToken = await ethers.getContractFactory("TestToken");
    const token = await TestToken.deploy(ethers.parseEther("1000000000")); // 1B
    await token.waitForDeployment();

    // approve factory to pull tokens
    const VIRTUAL_TOKEN = 1000000000n * 10n ** 18n;
    await (await token.approve(factory.target, VIRTUAL_TOKEN)).wait();

    // create pool
    const tx = await factory.createPool(token.target);
    const receipt = await tx.wait();
    const event = receipt.logs.find((l) => l.fragment?.name === "PoolCreated");
    const poolAddr = event.args.pool;

    // basic invariant
    expect(await factory.getPool(token.target)).to.equal(poolAddr);

    // attach pool
    const pool = await ethers.getContractAt("LaunchPool", poolAddr);

    // Check whether swap is feasible on this network (needs real USDC)
    const [reserveUSDC] = await pool.getRealReserves();
    const usdcDeployerBal = await usdc.balanceOf(deployer.address);
    // Always validate quote > 0 using virtual liquidity (no real USDC required)
    const USDC_ONE = 10n ** 18n; // Paxeer USDC = 18 decimals
    const quote = await pool.quoteUSDCToToken(USDC_ONE);
    expect(quote).to.be.gt(0n);
    if (reserveUSDC === 0n || usdcDeployerBal === 0n) {
      // Can't perform a live swap without real USDC; end test after invariant checks
      return;
    }

    // perform small swap via router
    await usdc.approve(router.target, USDC_ONE);
    const swapTx = await router.swapExactUSDCForTokens(token.target, USDC_ONE, 0, deployer.address);
    await swapTx.wait();
    // If we reach here, swap succeeded
  });
});
