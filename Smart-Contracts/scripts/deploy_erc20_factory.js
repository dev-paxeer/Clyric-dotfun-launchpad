/* eslint-disable no-console */
// Deploy ERC20Factory to Paxeer network

const { ethers, network } = require("hardhat");
const fs = require("fs");

async function main() {
  const [deployer] = await ethers.getSigners();
  console.log("Deployer:", deployer.address);
  console.log("Network:", network.name);

  const ERC20Factory = await ethers.getContractFactory("ERC20Factory");
  const factory = await ERC20Factory.deploy();
  await factory.waitForDeployment();
  console.log("ERC20Factory deployed at:", factory.target);

  // Update deployment-summary.json if present
  try {
    const path = "deployment-summary.json";
    let summary = {};
    if (fs.existsSync(path)) summary = JSON.parse(fs.readFileSync(path).toString());
    summary.erc20Factory = factory.target;
    summary.updatedAt = new Date().toISOString();
    fs.writeFileSync(path, JSON.stringify(summary, null, 2));
    console.log("Updated deployment-summary.json");
  } catch (e) {
    console.warn("Could not update deployment-summary.json:", e.message);
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
