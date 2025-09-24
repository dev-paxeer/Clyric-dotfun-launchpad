/* eslint-disable no-console */
// Exports clean ABIs for frontend/indexer consumption.
// Usage:
//   pnpm hardhat run scripts/export_abis.js
// Output:
//   ./abi-dist/<Contract>.json

const fs = require("fs");
const path = require("path");

const contracts = [
  { file: "LaunchpadFactory.sol", name: "LaunchpadFactory" },
  { file: "LaunchpadRouter.sol", name: "LaunchpadRouter" },
  { file: "LaunchPool.sol", name: "LaunchPool" },
  { file: "LaunchPoolOracle.sol", name: "LaunchPoolOracle" },
];

function readArtifact(contract) {
  const artifactPath = path.join(
    __dirname,
    "..",
    "artifacts",
    "contracts",
    contract.file,
    `${contract.name}.json`
  );
  if (!fs.existsSync(artifactPath)) {
    throw new Error(`Artifact not found: ${artifactPath}. Did you run 'pnpm hardhat compile'?`);
  }
  const raw = fs.readFileSync(artifactPath, "utf-8");
  return JSON.parse(raw);
}

function main() {
  const outDir = path.join(__dirname, "..", "abi-dist");
  fs.mkdirSync(outDir, { recursive: true });

  for (const c of contracts) {
    const artifact = readArtifact(c);
    const cleaned = {
      abi: artifact.abi,
      bytecode: artifact.bytecode,
      deployedBytecode: artifact.deployedBytecode,
      contractName: c.name,
    };
    const outPath = path.join(outDir, `${c.name}.json`);
    fs.writeFileSync(outPath, JSON.stringify(cleaned, null, 2));
    console.log("Wrote:", outPath);
  }
}

main();
