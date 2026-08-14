const test = require('node:test');
const assert = require('node:assert/strict');
const {
  generationCost,
  redemptionClipGrant,
  generationClipCost,
  hapwExchangeQuote,
  hapwExchangeDailySnapshot,
  dexPoolSnapshot,
  syncGenerationAccount
} = require('../server/mechanism');

test('creation credits round up consistently', () => {
  assert.equal(generationCost(15, 'standard'), 15);
  assert.equal(generationCost(15, 'pro'), 23);
  assert.equal(generationCost(30, 'pro'), 45);
  assert.throws(() => generationCost(15, 'unknown'), /invalid_generation_parameters/);
});

test('redemption grants CLIP and generation spends it by credit count', () => {
  assert.equal(redemptionClipGrant(150), 60);
  assert.equal(redemptionClipGrant(120), 48);
  assert.equal(generationClipCost(15), 3);
  assert.equal(generationClipCost(23), 5);
});

test('reserve HAPW quote includes a rounded five percent fee', () => {
  assert.deepEqual(hapwExchangeQuote(390), { price: 390, fee: 20, total: 410, feeRate: 0.05 });
  assert.deepEqual(hapwExchangeQuote(520), { price: 520, fee: 26, total: 546, feeRate: 0.05 });
});

test('reserve HAPW exchange snapshot enforces a per-day limit', () => {
  const exchanges = [
    { statusCode: 'completed', createdAt: '2026-08-12 09:00' },
    { statusCode: 'completed', createdAt: '2026-08-12 11:00' },
    { statusCode: 'completed', createdAt: '2026-08-11 19:00' },
    { statusCode: 'pending', createdAt: '2026-08-12 12:00' }
  ];
  assert.deepEqual(hapwExchangeDailySnapshot(exchanges, 2, '2026-08-12T13:00:00Z'), {
    date: '2026-08-12', timezone: 'UTC', resetsAt: '2026-08-13T00:00:00.000Z', dailyLimit: 2, usedToday: 2, remainingToday: 0, reached: true
  });
});

test('DEX snapshot exposes both reserves, rate, and total notional liquidity', () => {
  assert.deepEqual(dexPoolSnapshot({ clipReserve: 250000, usdtReserve: 25000, updatedAt: 'demo' }), {
    clipReserve: 250000,
    usdtReserve: 25000,
    clipPerUsdt: 10,
    usdtPerClip: 0.1,
    totalLiquidityUsdt: 50000,
    updatedAt: 'demo'
  });
});

test('generation account is derived from redemption and generation records', () => {
  const account = {};
  const redemptions = [
    { status: '有效', creditsGranted: 120, creditsRemaining: 97, clipGranted: 48 },
    { status: '有效', creditsGranted: 150, creditsRemaining: 105, clipGranted: 60 }
  ];
  const generations = [
    { creditsUsed: 23, clipCost: 5 },
    { creditsUsed: 45, clipCost: 9 }
  ];
  assert.deepEqual(syncGenerationAccount(account, redemptions, generations), {
    balance: 202,
    lifetimeGranted: 270,
    lifetimeUsed: 68,
    lifetimeClipGranted: 108,
    lifetimeClipSpent: 14
  });
});
