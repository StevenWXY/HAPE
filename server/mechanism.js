const GENERATION_FACTORS = Object.freeze({ standard: 1, pro: 1.5 });
const CLIP_GRANT_PER_CREDIT = 0.4;
const CLIP_COST_PER_CREDIT = 0.2;
const HAPW_EXCHANGE_FEE_RATE = 0.05;
const HAPW_EXCHANGE_DAILY_LIMIT = 2;

function generationCost(duration, quality) {
  const seconds = Number(duration);
  const factor = GENERATION_FACTORS[quality];
  if (!Number.isFinite(seconds) || seconds <= 0 || !factor) throw new RangeError('invalid_generation_parameters');
  return Math.ceil(seconds * factor);
}

function redemptionClipGrant(credits) {
  const value = Number(credits);
  if (!Number.isFinite(value) || value <= 0) throw new RangeError('invalid_credit_grant');
  return Math.floor(value * CLIP_GRANT_PER_CREDIT);
}

function generationClipCost(credits) {
  const value = Number(credits);
  if (!Number.isFinite(value) || value <= 0) throw new RangeError('invalid_credit_cost');
  return Math.ceil(value * CLIP_COST_PER_CREDIT);
}

function hapwExchangeQuote(price, feeRate = HAPW_EXCHANGE_FEE_RATE) {
  const value = Number(price);
  const rate = Number(feeRate);
  if (!Number.isFinite(value) || value <= 0 || !Number.isFinite(rate) || rate < 0 || rate > 1) throw new RangeError('invalid_exchange_quote');
  const fee = Math.ceil(value * rate);
  return { price: value, fee, total: value + fee, feeRate: rate };
}

function hapwExchangeDailySnapshot(exchanges, dailyLimit = HAPW_EXCHANGE_DAILY_LIMIT, now = new Date()) {
  if (!Array.isArray(exchanges)) throw new TypeError('invalid_exchange_records');
  const limit = Number(dailyLimit);
  if (!Number.isInteger(limit) || limit <= 0) throw new RangeError('invalid_daily_exchange_limit');
  const date = new Date(now);
  if (Number.isNaN(date.getTime())) throw new RangeError('invalid_exchange_date');
  const day = date.toISOString().slice(0, 10);
  const reset = new Date(`${day}T00:00:00.000Z`);
  reset.setUTCDate(reset.getUTCDate() + 1);
  const usedToday = exchanges.filter(item => item && item.statusCode === 'completed' && String(item.createdAt || '').slice(0, 10) === day).length;
  return { date: day, timezone: 'UTC', resetsAt: reset.toISOString(), dailyLimit: limit, usedToday, remainingToday: Math.max(0, limit - usedToday), reached: usedToday >= limit };
}

function dexPoolSnapshot(pool) {
  const clipReserve = Number(pool.clipReserve);
  const usdtReserve = Number(pool.usdtReserve);
  if (!Number.isFinite(clipReserve) || clipReserve <= 0 || !Number.isFinite(usdtReserve) || usdtReserve <= 0) throw new RangeError('invalid_liquidity_pool');
  return {
    clipReserve,
    usdtReserve,
    clipPerUsdt: clipReserve / usdtReserve,
    usdtPerClip: usdtReserve / clipReserve,
    totalLiquidityUsdt: usdtReserve * 2,
    updatedAt: pool.updatedAt
  };
}

function syncGenerationAccount(account, redemptions, generations) {
  account.balance = redemptions
    .filter(item => item.status === '有效')
    .reduce((sum, item) => sum + item.creditsRemaining, 0);
  account.lifetimeGranted = redemptions.reduce((sum, item) => sum + item.creditsGranted, 0);
  account.lifetimeUsed = generations.reduce((sum, item) => sum + item.creditsUsed, 0);
  account.lifetimeClipGranted = redemptions.reduce((sum, item) => sum + (item.clipGranted || 0), 0);
  account.lifetimeClipSpent = generations.reduce((sum, item) => sum + (item.clipCost || 0), 0);
  return account;
}

module.exports = {
  GENERATION_FACTORS,
  CLIP_GRANT_PER_CREDIT,
  CLIP_COST_PER_CREDIT,
  HAPW_EXCHANGE_FEE_RATE,
  HAPW_EXCHANGE_DAILY_LIMIT,
  generationCost,
  redemptionClipGrant,
  generationClipCost,
  hapwExchangeQuote,
  hapwExchangeDailySnapshot,
  dexPoolSnapshot,
  syncGenerationAccount
};
