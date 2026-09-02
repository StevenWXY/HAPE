const { ethers, network } = require("hardhat");

// Deployment helper for BNB Smart Chain. Pass real addresses and a funded
// deployer only from a secure CI secret store; the defaults are local-only.
async function main() {
  const [deployer] = await ethers.getSigners();
  const liveNetwork = network.name === "bscTestnet" || network.name === "bsc";
  const requiredLiveConfig = [
    "CLIP_TREASURY",
    "CLIP_TIMELOCK_PROPOSER",
    "CLIP_TIMELOCK_EXECUTOR",
    "CLIP_GUARDIAN",
    "CLIP_METRICS_ORACLE",
    "CLIP_CAMPAIGN_OPERATOR",
    "CLIP_INITIAL_SUPPLY",
    "CLIP_HARD_CAP",
  ];
  if (liveNetwork) {
    const missing = requiredLiveConfig.filter((name) => !process.env[name]);
    if (missing.length > 0) throw new Error(`Missing required live BNB deployment variables: ${missing.join(", ")}`);
  }
  const proposer = process.env.CLIP_TIMELOCK_PROPOSER || deployer.address;
  const executor = process.env.CLIP_TIMELOCK_EXECUTOR || deployer.address;
  const guardian = process.env.CLIP_GUARDIAN || proposer;
  const metricsOracle = process.env.CLIP_METRICS_ORACLE || proposer;
  const campaignOperator = process.env.CLIP_CAMPAIGN_OPERATOR || proposer;
  const treasury = process.env.CLIP_TREASURY || deployer.address;
  const timelockDelay = Number(process.env.CLIP_TIMELOCK_DELAY || 172800);
  const initialSupply = ethers.parseEther(process.env.CLIP_INITIAL_SUPPLY || "900000000");
  const hardCap = ethers.parseEther(process.env.CLIP_HARD_CAP || "1000000000");
  if (hardCap < initialSupply) throw new Error("CLIP_HARD_CAP must be >= CLIP_INITIAL_SUPPLY");

  const Timelock = await ethers.getContractFactory("ClipliTimelock");
  const timelock = await Timelock.deploy(timelockDelay, [proposer], [executor], ethers.ZeroAddress);
  await timelock.waitForDeployment();

  const Token = await ethers.getContractFactory("ClipToken");
  // Deployer is temporary admin only so it can wire roles atomically below.
  const token = await Token.deploy(deployer.address, treasury, initialSupply, hardCap);
  await token.waitForDeployment();

  const Minter = await ethers.getContractFactory("AdaptiveMinter");
  const minter = await Minter.deploy(deployer.address, await token.getAddress(), hardCap, 7 * 24 * 60 * 60, 2 * 24 * 60 * 60, 100, 500, 25, 500);
  await minter.waitForDeployment();
  const minterRole = await token.MINTER_ROLE();
  await (await token.grantRole(minterRole, await minter.getAddress())).wait();
  await (await minter.grantRole(await minter.METRICS_ROLE(), metricsOracle)).wait();
  await (await minter.grantRole(await minter.GUARDIAN_ROLE(), guardian)).wait();
  await (await minter.grantRole(await minter.DEFAULT_ADMIN_ROLE(), await timelock.getAddress())).wait();
  await (await minter.grantRole(await minter.GUARDIAN_ROLE(), await timelock.getAddress())).wait();
  await (await minter.renounceRole(await minter.DEFAULT_ADMIN_ROLE(), deployer.address)).wait();
  if (guardian.toLowerCase() !== deployer.address.toLowerCase()) await (await minter.renounceRole(await minter.GUARDIAN_ROLE(), deployer.address)).wait();

  const Airdrop = await ethers.getContractFactory("ClipAirdrop");
  const airdrop = await Airdrop.deploy(deployer.address, await token.getAddress());
  await airdrop.waitForDeployment();
  await (await airdrop.grantRole(await airdrop.CAMPAIGN_ROLE(), campaignOperator)).wait();
  await (await airdrop.grantRole(await airdrop.DEFAULT_ADMIN_ROLE(), await timelock.getAddress())).wait();
  await (await airdrop.renounceRole(await airdrop.DEFAULT_ADMIN_ROLE(), deployer.address)).wait();
  if (campaignOperator.toLowerCase() !== deployer.address.toLowerCase()) await (await airdrop.renounceRole(await airdrop.CAMPAIGN_ROLE(), deployer.address)).wait();

  await (await token.grantRole(await token.PAUSER_ROLE(), guardian)).wait();
  await (await token.grantRole(await token.PAUSER_ROLE(), await timelock.getAddress())).wait();
  await (await token.grantRole(await token.DEFAULT_ADMIN_ROLE(), await timelock.getAddress())).wait();
  await (await token.renounceRole(await token.DEFAULT_ADMIN_ROLE(), deployer.address)).wait();
  if (guardian.toLowerCase() !== deployer.address.toLowerCase()) await (await token.renounceRole(await token.PAUSER_ROLE(), deployer.address)).wait();

  console.log(JSON.stringify({
    network: (await ethers.provider.getNetwork()).chainId.toString(),
    deployer: deployer.address,
    timelock: await timelock.getAddress(),
    proposer,
    executor,
    guardian,
    metricsOracle,
    campaignOperator,
    treasury,
    clipToken: await token.getAddress(),
    adaptiveMinter: await minter.getAddress(),
    airdrop: await airdrop.getAddress(),
    initialSupply: initialSupply.toString(),
    hardCap: hardCap.toString(),
    note: "BNB EVM deployment. Verify bytecode, roles, timelock and audit before funding or mainnet use.",
  }, null, 2));
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
