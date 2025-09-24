/* eslint-env mocha */
// @post-deploy
// Validates on-chain reporting: PriceUpdate, Sync, InitialTokenSeeded, and views

const { expect } = require("chai");
const fs = require("fs");
const { ethers } = require("hardhat");

function findLog(receipt, name) {
  return receipt.logs.find((l) => l.fragment?.name === name || l.eventName === name);
}

describe("On-chain reporting", function () {
  let summary, factory;

  before("attach factory", async function () {
    summary = JSON.parse(fs.readFileSync("deployment-summary.json"));
    factory = await ethers.getContractAt("LaunchpadFactory", summary.factory);
  });

  it("emits PriceUpdate and Sync on initial seed @post-deploy", async function () {
    const TestToken = await ethers.getContractFactory("TestToken");
    const token = await TestToken.deploy(ethers.parseEther("1000000000")); // 1B
    await token.waitForDeployment();

    const SEED = 1000000000n * 10n ** 18n;
    await (await token.approve(factory.target, SEED)).wait();

    const tx = await factory.createPool(token.target);
    const rc = await tx.wait();

    const poolAddr = findLog(rc, "PoolCreated").args.pool;
    const pool = await ethers.getContractAt("LaunchPool", poolAddr);

    // Expect events from seed call
    const seeded = findLog(rc, "InitialTokenSeeded");
    const priceUp = findLog(rc, "PriceUpdate");
    const sync = findLog(rc, "Sync");

    expect(poolAddr).to.properAddress;
    expect(seeded, "InitialTokenSeeded not emitted").to.exist;
    expect(priceUp, "PriceUpdate not emitted").to.exist;
    expect(sync, "Sync not emitted").to.exist;

    // Validate views are consistent
    const [vUSDC, rUSDC, rTok, spotX18, floorX18] = await pool.getState();
    expect(vUSDC).to.be.gt(0n);
    expect(rUSDC).to.equal(0n);
    expect(rTok).to.equal(SEED);
    expect(spotX18).to.equal(floorX18);

    const spotNow = await pool.currentPriceX18();
    expect(spotNow).to.equal(spotX18);
  });
});
