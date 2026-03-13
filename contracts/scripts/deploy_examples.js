const hre = require("hardhat");

function readUintEnv(name, fallback) {
  const raw = process.env[name];
  if (!raw) {
    return fallback;
  }

  const value = Number(raw);
  if (!Number.isFinite(value) || value < 0 || !Number.isInteger(value)) {
    throw new Error(`Invalid integer for ${name}: ${raw}`);
  }

  return value;
}

async function deployContract(name, args) {
  const factory = await hre.ethers.getContractFactory(name);
  const contract = await factory.deploy(...args);
  await contract.waitForDeployment();
  const address = await contract.getAddress();
  console.log(`${name}: ${address}`);
  return address;
}

async function main() {
  const config = {
    daoMinReputation: readUintEnv("AXON_DAO_MIN_REPUTATION", 60),
    daoVotingPeriod: readUintEnv("AXON_DAO_VOTING_PERIOD", 7200),
    vaultMinReputation: readUintEnv("AXON_VAULT_MIN_REPUTATION", 50),
  };

  console.log("Deploying Axon example contracts with config:");
  console.log(JSON.stringify(config, null, 2));

  await deployContract("AgentDAO", [config.daoMinReputation, config.daoVotingPeriod]);
  await deployContract("AgentMarketplace", []);
  await deployContract("ReputationVault", [config.vaultMinReputation]);
  await deployContract("TrustChannelExample", []);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
