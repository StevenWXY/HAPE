const { expect } = require("chai");

const GENESIS = 900_000_000;
const HARD_CAP = 1_000_000_000;
const EPOCHS_PER_YEAR = 52;
const EPOCH_CAP_BPS = 25;
const ANNUAL_CAP_BPS = 500;
const MIN_GROWTH_BPS = 500;

function simulateSupply(years, growthBps) {
  let supply = GENESIS;
  const annual = [];
  for (let year = 0; year < years; year += 1) {
    let issued = 0;
    for (let epoch = 0; epoch < EPOCHS_PER_YEAR; epoch += 1) {
      if (growthBps < MIN_GROWTH_BPS || supply >= HARD_CAP) continue;
      const score = Math.min(growthBps, 10_000);
      const proposed = Math.floor((GENESIS * score * EPOCH_CAP_BPS) / 100_000_000);
      const epochCap = Math.floor((GENESIS * EPOCH_CAP_BPS) / 10_000);
      const annualCap = Math.floor((GENESIS * ANNUAL_CAP_BPS) / 10_000);
      const amount = Math.min(proposed, epochCap, annualCap - issued, HARD_CAP - supply);
      supply += amount;
      issued += amount;
    }
    annual.push(issued);
  }
  return { supply, annual };
}

function amountOut(amountIn, reserveIn, reserveOut) {
  const withFee = amountIn * 997;
  return (withFee * reserveOut) / (reserveIn * 1000 + withFee);
}

describe("CLIP economic stability model", function () {
  it("issues nothing below the 5% weighted-growth threshold", function () {
    const model = simulateSupply(10, 499);
    expect(model.supply).to.equal(GENESIS);
    expect(model.annual.every((amount) => amount === 0)).to.equal(true);
  });

  it("never exceeds the 1 billion hard cap under maximum growth", function () {
    const model = simulateSupply(10, 10_000);
    expect(model.supply).to.equal(HARD_CAP);
    expect(model.annual.every((amount) => amount <= 45_000_000)).to.equal(true);
    expect(model.annual.reduce((sum, amount) => sum + amount, 0)).to.equal(HARD_CAP - GENESIS);
  });

  it("keeps sustainable growth gradual over ten years", function () {
    const model = simulateSupply(10, 700);
    expect(model.supply).to.be.greaterThan(GENESIS);
    expect(model.supply).to.be.lessThan(HARD_CAP);
    expect(model.annual.every((amount) => amount <= 45_000_000)).to.equal(true);
  });

  it("shows rising AMM price impact for large DEX trades", function () {
    const reserveUSDT = 10_000_000;
    const reserveCLIP = 100_000_000;
    const spotCLIPPerUSDT = reserveCLIP / reserveUSDT;
    const fractions = [0.001, 0.01, 0.1, 0.5];
    const impacts = fractions.map((fraction) => {
      const input = reserveUSDT * fraction;
      const output = amountOut(input, reserveUSDT, reserveCLIP);
      const ideal = input * spotCLIPPerUSDT;
      expect(output).to.be.greaterThan(0);
      expect(output).to.be.lessThan(ideal);
      return 1 - output / ideal;
    });
    for (let index = 1; index < impacts.length; index += 1) {
      expect(impacts[index]).to.be.greaterThan(impacts[index - 1]);
    }
    expect(impacts[0]).to.be.lessThan(0.005);
    expect(impacts[3]).to.be.greaterThan(0.33);
  });
});
