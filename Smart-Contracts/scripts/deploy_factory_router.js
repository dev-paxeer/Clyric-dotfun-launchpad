/* eslint-disable no-console */
// Deploy LaunchpadFactory and LaunchpadRouter to Paxeer network
// Then run test suite against the live contracts

const { run, ethers, network } = require("hardhat");

async function main() {
  const USDC_ADDRESS = "0x61Be934234717c57585d5f558360aFA59F8adB56";

  const [deployer] = await ethers.getSigners();
  console.log("Deployer:", deployer.address);
  console.log("Network:", network.name);

  // 1. Deploy LaunchpadFactory
  const LaunchpadFactory = await ethers.getContractFactory("LaunchpadFactory");
  const factory = await LaunchpadFactory.deploy(USDC_ADDRESS, deployer.address);
  await factory.waitForDeployment();
  console.log("LaunchpadFactory deployed at:", factory.target);

  // 2. Deploy LaunchpadRouter
  const LaunchpadRouter = await ethers.getContractFactory("LaunchpadRouter");
  const router = await LaunchpadRouter.deploy(factory.target, USDC_ADDRESS);
  await router.waitForDeployment();
  console.log("LaunchpadRouter deployed at:", router.target);

  // 3. Save deployment summary
  const fs = require("fs");
  const summary = {
    network: network.name,
    factory: factory.target,
    router: router.target,
    usdc: USDC_ADDRESS,
    deployedAt: new Date().toISOString(),
  };
  fs.writeFileSync("deployment-summary.json", JSON.stringify(summary, null, 2));
  console.log("Saved deployment-summary.json\n");

  // 4. Run Hardhat test suite (uses --grep @post-deploy to selectively run)
  console.log("\nRunning post-deployment tests ...\n");
  await run("test", { grep: "@post-deploy" }); // only tests tagged with @post-deploy

  console.log("All done ✅");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
