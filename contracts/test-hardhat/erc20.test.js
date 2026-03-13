const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("TestERC20", function () {
  let token, owner, addr1;

  beforeEach(async function () {
    [owner, addr1] = await ethers.getSigners();
    const ERC20 = await ethers.getContractFactory("TestERC20");
    token = await ERC20.deploy("Test Token", "TT", 18, ethers.parseEther("1000"));
    await token.waitForDeployment();
  });

  it("should have correct name and symbol", async function () {
    expect(await token.name()).to.equal("Test Token");
    expect(await token.symbol()).to.equal("TT");
  });

  it("should have 18 decimals", async function () {
    expect(await token.decimals()).to.equal(18n);
  });

  it("should assign total supply to deployer", async function () {
    const balance = await token.balanceOf(owner.address);
    expect(balance).to.equal(ethers.parseEther("1000"));
  });

  it("should transfer tokens", async function () {
    await token.transfer(addr1.address, ethers.parseEther("100"));
    expect(await token.balanceOf(addr1.address)).to.equal(ethers.parseEther("100"));
    expect(await token.balanceOf(owner.address)).to.equal(ethers.parseEther("900"));
  });

  it("should emit Transfer event", async function () {
    await expect(token.transfer(addr1.address, ethers.parseEther("50")))
      .to.emit(token, "Transfer")
      .withArgs(owner.address, addr1.address, ethers.parseEther("50"));
  });

  it("should handle approve and transferFrom", async function () {
    await token.approve(addr1.address, ethers.parseEther("200"));
    expect(await token.allowance(owner.address, addr1.address)).to.equal(ethers.parseEther("200"));

    await token.connect(addr1).transferFrom(owner.address, addr1.address, ethers.parseEther("150"));
    expect(await token.balanceOf(addr1.address)).to.equal(ethers.parseEther("150"));
  });
});

describe("Precompile: IAgentRegistry (0x...0801)", function () {
  const REGISTRY = "0x0000000000000000000000000000000000000801";
  const abi = [
    "function isAgent(address) view returns (bool)",
    "function getAgent(address) view returns (string, string[], string, uint64, bool)",
  ];

  it("should respond to isAgent query", async function () {
    const registry = new ethers.Contract(REGISTRY, abi, ethers.provider);
    const result = await registry.isAgent("0x0000000000000000000000000000000000000001");
    expect(typeof result).to.equal("boolean");
  });

  it("should return agent info struct", async function () {
    const registry = new ethers.Contract(REGISTRY, abi, ethers.provider);
    const result = await registry.getAgent("0x0000000000000000000000000000000000000001");
    expect(result).to.have.lengthOf(5);
  });
});

describe("Precompile: IAgentReputation (0x...0802)", function () {
  const REPUTATION = "0x0000000000000000000000000000000000000802";
  const abi = [
    "function getReputation(address) view returns (uint64)",
    "function meetsReputation(address, uint64) view returns (bool)",
    "function getReputations(address[]) view returns (uint64[])",
  ];

  it("should return reputation score", async function () {
    const rep = new ethers.Contract(REPUTATION, abi, ethers.provider);
    const score = await rep.getReputation("0x0000000000000000000000000000000000000001");
    expect(score).to.be.a("bigint");
  });

  it("should check meetsReputation", async function () {
    const rep = new ethers.Contract(REPUTATION, abi, ethers.provider);
    const meets = await rep.meetsReputation("0x0000000000000000000000000000000000000001", 0);
    expect(typeof meets).to.equal("boolean");
  });

  it("should batch query reputations", async function () {
    const rep = new ethers.Contract(REPUTATION, abi, ethers.provider);
    const scores = await rep.getReputations([
      "0x0000000000000000000000000000000000000001",
      "0x0000000000000000000000000000000000000002",
    ]);
    expect(scores).to.have.lengthOf(2);
  });
});

describe("Precompile: IAgentWallet (0x...0803)", function () {
  const WALLET = "0x0000000000000000000000000000000000000803";
  const abi = [
    "function getWalletInfo(address) view returns (uint256, uint256, uint256, bool, address, address, address)",
    "function getTrust(address, address) view returns (uint8, uint256, uint256, uint256, uint256)",
  ];

  it("should return wallet info with 7 fields", async function () {
    const w = new ethers.Contract(WALLET, abi, ethers.provider);
    const info = await w.getWalletInfo("0x0000000000000000000000000000000000000001");
    expect(info).to.have.lengthOf(7);
  });

  it("should return trust info with 5 fields", async function () {
    const w = new ethers.Contract(WALLET, abi, ethers.provider);
    const trust = await w.getTrust(
      "0x0000000000000000000000000000000000000001",
      "0x0000000000000000000000000000000000000002"
    );
    expect(trust).to.have.lengthOf(5);
  });
});
