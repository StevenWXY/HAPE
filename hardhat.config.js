require("@nomicfoundation/hardhat-toolbox");
const path = require("node:path");

// Hardhat runs with BNB Smart Chain Testnet's chain id. Accounts are
// ephemeral and no private key is loaded from disk.
const networks = { hardhat: { chainId: 97 } };
if (process.env.BNB_TESTNET_RPC_URL && process.env.BNB_DEPLOYER_PRIVATE_KEY) {
  networks.bscTestnet = { url: process.env.BNB_TESTNET_RPC_URL, chainId: 97, accounts: [process.env.BNB_DEPLOYER_PRIVATE_KEY] };
}
if (process.env.BNB_MAINNET_RPC_URL && process.env.BNB_DEPLOYER_PRIVATE_KEY) {
  networks.bsc = { url: process.env.BNB_MAINNET_RPC_URL, chainId: 56, accounts: [process.env.BNB_DEPLOYER_PRIVATE_KEY] };
}

const buildRoot = process.env.CLIPLI_HARDHAT_BUILD_ROOT;

module.exports = {
  solidity: {
    version: "0.8.24",
    settings: { optimizer: { enabled: true, runs: 200 }, viaIR: true },
  },
  networks,
  ...(buildRoot
    ? { paths: { artifacts: path.join(buildRoot, "artifacts"), cache: path.join(buildRoot, "cache") } }
    : {}),
};
