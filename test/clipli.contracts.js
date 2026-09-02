const { expect } = require("chai");
const { ethers } = require("hardhat");

const e18 = (n) => ethers.parseEther(String(n));
const day = 24 * 60 * 60;

async function deployCore() {
  const [admin, treasury, recipient, holder, other] = await ethers.getSigners();
  const Token = await ethers.getContractFactory("ClipToken");
  const token = await Token.deploy(admin.address, treasury.address, e18(900_000_000), e18(1_000_000_000));
  await token.waitForDeployment();
  const Minter = await ethers.getContractFactory("AdaptiveMinter");
  const minter = await Minter.deploy(admin.address, await token.getAddress(), e18(1_000_000_000), day, 3600, 100, 500, 25, 50);
  await minter.waitForDeployment();
  await token.grantRole(await token.MINTER_ROLE(), await minter.getAddress());
  await minter.grantRole(await minter.METRICS_ROLE(), admin.address);
  return { admin, treasury, recipient, holder, other, token, minter };
}

describe("Clipli BNB token contracts", function () {
  it("mints 900 million genesis CLIP with a 1 billion hard cap", async function () {
    const { admin, treasury, token } = await deployCore();
    expect(await token.totalSupply()).to.equal(e18(900_000_000));
    expect(await token.balanceOf(treasury.address)).to.equal(e18(900_000_000));
    await expect(token.connect(treasury).mint(treasury.address, 1)).to.be.reverted;
    await token.connect(admin).pause();
    await expect(token.transfer(admin.address, 1)).to.be.reverted;
    await token.connect(admin).unpause();
  });

  it("requires growth, delay and annual cap for adaptive issuance", async function () {
    const { admin, recipient, other, token, minter } = await deployCore();
    await expect(minter.connect(other).recordMetrics(100, 100, 1000)).to.be.reverted;
    await ethers.provider.send("evm_increaseTime", [day + 1]);
    await ethers.provider.send("evm_mine");
    await minter.recordMetrics(100, 100, 1000);
    await ethers.provider.send("evm_increaseTime", [day + 1]);
    await ethers.provider.send("evm_mine");
    await minter.recordMetrics(200, 150, 2000);
    const epoch = await minter.lastRecordedEpoch();
    const quote = await minter.emissionFor(epoch);
    expect(quote[0]).to.be.gt(0);
    await expect(minter.queueEmission(epoch, recipient.address)).to.emit(minter, "EmissionQueued");
    await expect(minter.queueEmission(epoch, other.address)).to.be.reverted;
    const proposalId = ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(["address", "uint256", "address", "uint256", "uint256"], [await minter.getAddress(), epoch, recipient.address, quote[0], 0]));
    await expect(minter.executeEmission(proposalId)).to.be.reverted;

    await expect(minter.cancelEmission(proposalId)).to.emit(minter, "EmissionCancelled");
    await expect(minter.queueEmission(epoch, admin.address)).to.emit(minter, "EmissionQueued");
    const secondProposalId = ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(["address", "uint256", "address", "uint256", "uint256"], [await minter.getAddress(), epoch, admin.address, quote[0], 1]));
    await ethers.provider.send("evm_increaseTime", [3601]);
    await ethers.provider.send("evm_mine");
    await expect(minter.executeEmission(secondProposalId)).to.emit(minter, "EmissionExecuted");
    expect(await token.balanceOf(admin.address)).to.equal(quote[0]);
    await ethers.provider.send("evm_increaseTime", [day + 1]);
    await ethers.provider.send("evm_mine");
    await minter.recordMetrics(ethers.MaxUint256, ethers.MaxUint256, ethers.MaxUint256);
    const saturated = await minter.emissionFor(3);
    expect(saturated[1]).to.equal(10_000);
    await minter.setIssuancePaused(true);
    await expect(minter.queueEmission(epoch, recipient.address)).to.be.reverted;
    await minter.connect(admin).setIssuancePaused(false);
    await expect(minter.connect(other).setIssuancePaused(true)).to.be.reverted;
  });

  it("locks and claims an asset-specific Merkle airdrop once", async function () {
    const { admin, treasury, holder, other, token } = await deployCore();
    const Airdrop = await ethers.getContractFactory("ClipAirdrop");
    const airdrop = await Airdrop.deploy(admin.address, await token.getAddress());
    await airdrop.waitForDeployment();
    const amount = e18(60);
    const inner = ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(["uint256", "address", "uint256"], [0, holder.address, amount]));
    const leaf = ethers.keccak256(ethers.concat([inner]));
    await token.connect(treasury).transfer(admin.address, amount);
    await token.connect(admin).approve(await airdrop.getAddress(), amount);
    const start = (await ethers.provider.getBlock("latest")).timestamp;
    const end = start + 3600;
    const campaign = ethers.keccak256(ethers.toUtf8Bytes("HAPW-2048-epoch-1"));
    const assetRef = ethers.keccak256(ethers.toUtf8Bytes("haiwen:HWF-HAPW-2048"));
    await airdrop.createCampaign(campaign, assetRef, leaf, amount, start, end);
    await expect(airdrop.claim(campaign, 0, holder.address, amount, [])).to.emit(airdrop, "Claimed");
    expect(await token.balanceOf(holder.address)).to.equal(amount);
    await expect(airdrop.claim(campaign, 0, holder.address, amount, [])).to.be.reverted;
    await expect(airdrop.claim(campaign, 1, other.address, amount, [])).to.be.reverted;
    await expect(airdrop.createCampaign(campaign, assetRef, leaf, amount, start, end)).to.be.reverted;
    await expect(airdrop.connect(other).pause()).to.be.reverted;
  });

  it("cannot sweep another campaign's escrow when one campaign closes", async function () {
    const { admin, treasury, holder, other, token } = await deployCore();
    const Airdrop = await ethers.getContractFactory("ClipAirdrop");
    const airdrop = await Airdrop.deploy(admin.address, await token.getAddress());
    await airdrop.waitForDeployment();
    const amount = e18(60);
    const inner = ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(["uint256", "address", "uint256"], [0, holder.address, amount]));
    const leaf = ethers.keccak256(ethers.concat([inner]));
    const now = (await ethers.provider.getBlock("latest")).timestamp;
    const first = ethers.keccak256(ethers.toUtf8Bytes("campaign-first"));
    const second = ethers.keccak256(ethers.toUtf8Bytes("campaign-second"));
    const assetRef = ethers.keccak256(ethers.toUtf8Bytes("haiwen:HWF-HAPW-2048"));
    await token.connect(treasury).transfer(admin.address, amount * 2n);
    await token.connect(admin).approve(await airdrop.getAddress(), amount * 2n);
    await airdrop.createCampaign(first, assetRef, leaf, amount, now, now + 100);
    await airdrop.createCampaign(second, assetRef, leaf, amount, now, now + 5000);
    await ethers.provider.send("evm_increaseTime", [101]);
    await ethers.provider.send("evm_mine");
    const before = await token.balanceOf(other.address);
    await airdrop.closeCampaign(first, other.address);
    expect((await token.balanceOf(other.address)) - before).to.equal(amount);
    expect(await token.balanceOf(await airdrop.getAddress())).to.equal(amount);
    await expect(airdrop.closeCampaign(first, other.address)).to.be.reverted;
  });

  it("enforces the BNB AMM invariant, deadline and slippage", async function () {
    const { treasury, holder, other, token } = await deployCore();
    const USDT = await ethers.getContractFactory("MockUSDT");
    const usdt = await USDT.deploy();
    await usdt.waitForDeployment();
    const Pair = await ethers.getContractFactory("ClipliDexPair");
    const pair = await Pair.deploy(await token.getAddress(), await usdt.getAddress());
    await pair.waitForDeployment();
    const clipLiquidity = e18(100_000);
    const usdtLiquidity = e18(10_000);
    await usdt.mint(treasury.address, usdtLiquidity * 2n);
    await token.connect(treasury).approve(await pair.getAddress(), clipLiquidity);
    await usdt.connect(treasury).approve(await pair.getAddress(), usdtLiquidity * 2n);
    const deadline = (await ethers.provider.getBlock("latest")).timestamp + 3600;
    const token0 = await pair.token0();
    const amount0 = token0.toLowerCase() === (await token.getAddress()).toLowerCase() ? clipLiquidity : usdtLiquidity;
    const amount1 = token0.toLowerCase() === (await token.getAddress()).toLowerCase() ? usdtLiquidity : clipLiquidity;
    await pair.connect(treasury).addLiquidity(amount0, amount1, 1, deadline);
    await usdt.mint(holder.address, e18(100));
    await usdt.connect(holder).approve(await pair.getAddress(), e18(100));
    const before = await token.balanceOf(holder.address);
    await pair.connect(holder).swapExactIn(await usdt.getAddress(), e18(100), e18(900), holder.address, deadline);
    expect(await token.balanceOf(holder.address)).to.be.gt(before);
    await expect(pair.connect(holder).swapExactIn(await usdt.getAddress(), e18(1), e18(100), holder.address, deadline)).to.be.reverted;
    await ethers.provider.send("evm_increaseTime", [3601]);
    await ethers.provider.send("evm_mine");
    await expect(pair.connect(holder).swapExactIn(await usdt.getAddress(), e18(1), 0, holder.address, deadline)).to.be.reverted;
    await expect(pair.connect(other).removeLiquidity(1, 0, 0, deadline + 3600)).to.be.reverted;
  });

  it("supports the reverse BNB-pair direction with deadline protection", async function () {
    const { treasury, holder, token } = await deployCore();
    const USDT = await ethers.getContractFactory("MockUSDT");
    const usdt = await USDT.deploy();
    await usdt.waitForDeployment();
    const Pair = await ethers.getContractFactory("ClipliDexPair");
    const pair = await Pair.deploy(await token.getAddress(), await usdt.getAddress());
    await pair.waitForDeployment();
    await usdt.mint(treasury.address, e18(10_000));
    await token.connect(treasury).approve(await pair.getAddress(), e18(100_000));
    await usdt.connect(treasury).approve(await pair.getAddress(), e18(10_000));
    const deadline = (await ethers.provider.getBlock("latest")).timestamp + 3600;
    const token0 = await pair.token0();
    const amount0 = token0.toLowerCase() === (await token.getAddress()).toLowerCase() ? e18(100_000) : e18(10_000);
    const amount1 = token0.toLowerCase() === (await token.getAddress()).toLowerCase() ? e18(10_000) : e18(100_000);
    await pair.connect(treasury).addLiquidity(amount0, amount1, 1, deadline);
    await token.connect(treasury).transfer(holder.address, e18(100));
    await token.connect(holder).approve(await pair.getAddress(), e18(100));
    const before = await usdt.balanceOf(holder.address);
    await pair.connect(holder).swapExactIn(await token.getAddress(), e18(100), e18(9), holder.address, deadline);
    expect(await usdt.balanceOf(holder.address)).to.be.gt(before);
  });
});
