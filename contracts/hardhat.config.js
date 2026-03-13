require("@nomicfoundation/hardhat-toolbox");

const axonNetwork = {
  url: process.env.AXON_RPC_URL,
};

if (process.env.AXON_CHAIN_ID) {
  axonNetwork.chainId = Number(process.env.AXON_CHAIN_ID);
}

if (process.env.AXON_MNEMONIC) {
  axonNetwork.accounts = {
    mnemonic: process.env.AXON_MNEMONIC,
  };
}

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: "0.8.20",
  networks: {
    axon: axonNetwork,
  },
  paths: {
    sources: "./",
    tests: "./test-hardhat",
    cache: "./cache",
    artifacts: "./artifacts",
  },
};
