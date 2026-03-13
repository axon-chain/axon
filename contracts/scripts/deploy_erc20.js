const { ethers } = require("hardhat");

async function main() {
  const [deployer] = await ethers.getSigners();
  console.log("Deploying with:", deployer.address);
  console.log("Balance:", ethers.formatEther(await ethers.provider.getBalance(deployer.address)), "AXON");

  const ERC20 = await ethers.getContractFactory("TestERC20");
  const token = await ERC20.deploy("Axon Test Token", "ATT", 18, ethers.parseEther("1000000"));
  await token.waitForDeployment();

  console.log("TestERC20 deployed to:", await token.getAddress());
  console.log("Total supply:", ethers.formatEther(await token.totalSupply()));
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
