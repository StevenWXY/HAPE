const http = require('http');
const fs = require('fs');
const path = require('path');
const { URL } = require('url');
const store = require('./data');
const workSource = require('./work-source');
const {
  GENERATION_FACTORS,
  generationCost,
  redemptionClipGrant,
  generationClipCost,
  hapwExchangeQuote,
  HAPW_EXCHANGE_DAILY_LIMIT,
  hapwExchangeDailySnapshot,
  dexPoolSnapshot,
  syncGenerationAccount
} = require('./mechanism');

const ROOT = path.join(__dirname, '..');
const PUBLIC = path.join(ROOT, 'public');
const PORT = Number(process.env.PORT || 4173);
const HOST = process.env.HOST || '127.0.0.1';
const SECURITY_HEADERS = {
  'Content-Security-Policy': "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'",
  'Referrer-Policy': 'strict-origin-when-cross-origin',
  'X-Content-Type-Options': 'nosniff',
  'X-Frame-Options': 'DENY'
};

const json = (res, status, payload) => {
  res.writeHead(status, { ...SECURITY_HEADERS, 'Content-Type': 'application/json; charset=utf-8', 'Cache-Control': 'no-store' });
  res.end(JSON.stringify(payload));
};

const error = (res, status, code, message) => json(res, status, { error: { code, message } });

const readBody = (req) => new Promise((resolve, reject) => {
  let body = '';
  req.on('data', chunk => {
    body += chunk;
    if (body.length > 1e6) req.destroy();
  });
  req.on('end', () => {
    if (!body) return resolve({});
    try { resolve(JSON.parse(body)); } catch { reject(new Error('invalid_json')); }
  });
  req.on('error', reject);
});

const assetName = (asset) => `${asset.name} ${asset.tokenId}`;
const validRequestId = value => typeof value === 'string' && value.length >= 8 && value.length <= 100;

function hapwExchangePolicy() {
  const reserveAssets = store.assets.filter(item => item.kind === 'HAPW' && item.clipPrice > 0);
  return {
    ...hapwExchangeDailySnapshot(store.hapwExchanges, HAPW_EXCHANGE_DAILY_LIMIT),
    inventoryTotal: reserveAssets.length,
    inventoryAvailable: reserveAssets.filter(item => item.exchangeAvailable).length
  };
}

function handleApi(req, res, url) {
  if (req.method === 'GET' && url.pathname === '/api/overview') {
    return json(res, 200, { data: {
      counts: { authorized: 18420, overseas: 2908, visible: 86, redeemed: store.hapwRedemptions.length, credits: store.generationAccount.balance, clipGranted: store.generationAccount.lifetimeClipGranted },
      featuredAsset: store.assets[0],
      steps: ['hold', 'redeem', 'grant', 'generate']
    } });
  }
  if (req.method === 'GET' && url.pathname === '/api/works') {
    return workSource.listWorks()
      .then(works => json(res, 200, { data: works }))
      .catch(() => error(res, 502, 'work_source_failed', 'Unable to load works'));
  }
  if (req.method === 'GET' && url.pathname.startsWith('/api/works/')) {
    const workId = decodeURIComponent(url.pathname.slice('/api/works/'.length));
    return workSource.listWorks()
      .then(works => {
        const work = works.find(item => item.id === workId);
        return work
          ? json(res, 200, { data: work })
          : error(res, 404, 'work_not_found', 'Work not found');
      })
      .catch(() => error(res, 502, 'work_source_failed', 'Unable to load work'));
  }
  if (req.method === 'GET' && url.pathname === '/api/assets') {
    syncGenerationAccount(store.generationAccount, store.hapwRedemptions, store.generations);
    const assets = store.assets.filter(item => item.kind === 'HAPW' && item.owner === 'Clipli');
    const holdings = assets.filter(item => item.redemptionStatus !== 'redeemed');
    return json(res, 200, { data: {
      assets,
      stats: { holdings: holdings.length, transferable: holdings.filter(item => item.transferable).length, recent: store.transfers.length, totalValue: holdings.reduce((sum, item) => sum + item.value, 0) },
      transfers: store.transfers,
      exercises: store.transfers,
      platforms: store.externalPlatforms,
      clip: { ...store.clip, dexPool: dexPoolSnapshot(store.clip.dexPool), hapwExchangePolicy: hapwExchangePolicy() },
      clipTransactions: store.clipTransactions,
      generationAccount: store.generationAccount,
      hapwRedemptions: store.hapwRedemptions,
      generations: store.generations,
      hapwExchanges: store.hapwExchanges
    } });
  }
  if (req.method === 'GET' && url.pathname === '/api/studio') {
    syncGenerationAccount(store.generationAccount, store.hapwRedemptions, store.generations);
    return json(res, 200, { data: {
      assets: store.assets.filter(item => item.kind === 'HAPW' && item.owner === 'Clipli'),
      generationAccount: store.generationAccount,
      redemptions: store.hapwRedemptions,
      generations: store.generations,
      clip: { ...store.clip, dexPool: dexPoolSnapshot(store.clip.dexPool) }
    } });
  }
  if (req.method === 'GET' && url.pathname === '/api/profile') return json(res, 200, { data: { ...store.profile, clipBalance: store.clip.balance } });

  if (req.method === 'POST' && url.pathname === '/api/hapw/redemptions') {
    return readBody(req).then(body => {
      if (!validRequestId(body.requestId)) return error(res, 400, 'invalid_request_id', 'A valid operation id is required');
      const existing = store.hapwRedemptions.find(item => item.requestId === body.requestId);
      if (existing) {
        syncGenerationAccount(store.generationAccount, store.hapwRedemptions, store.generations);
        return json(res, 200, { data: { redemption: existing, generationAccount: store.generationAccount } });
      }
      const asset = store.assets.find(item => item.id === body.assetId && item.owner === 'Clipli');
      if (!asset || asset.redemptionStatus !== 'available') return error(res, 400, 'hapw_not_redeemable', 'This HAPW is not available for redemption');
      if (body.accepted !== true) return error(res, 400, 'redemption_not_confirmed', 'Confirm the irreversible HAPW redemption');
      asset.redemptionStatus = 'redeemed';
      asset.transferable = false;
      asset.status = '已核销';
      asset.statusEn = 'Redeemed';
      asset.statusKo = '상각 완료';
      const clipGranted = redemptionClipGrant(asset.creditYield);
      const now = new Date().toISOString().slice(0, 16).replace('T', ' ');
      const redemption = {
        id: `redeem-${Date.now()}`,
        assetId: asset.id,
        receipt: `Clipli-LIC-${asset.tokenId.replace('#', '')}-${Date.now().toString().slice(-4)}`,
        creditsGranted: asset.creditYield,
        creditsRemaining: asset.creditYield,
        clipGranted,
        requestId: body.requestId,
        status: '有效',
        statusEn: 'Active',
        statusKo: '유효',
        createdAt: now
      };
      store.hapwRedemptions.unshift(redemption);
      store.clip.balance += clipGranted;
      store.clipTransactions.unshift({
        id: `clip-grant-${Date.now()}`,
        typeCode: 'redemptionGrant',
        type: 'HAPW 核销领取',
        typeEn: 'HAPW redemption grant',
        typeKo: 'HAPW 상각 지급',
        amount: clipGranted,
        counterparty: `HAPW ${asset.tokenId} · ${asset.name}`,
        counterpartyEn: `HAPW ${asset.tokenId} · ${asset.nameEn}`,
        counterpartyKo: `HAPW ${asset.tokenId} · ${asset.nameKo || asset.nameEn}`,
        statusCode: 'completed', status: '已完成', statusEn: 'Completed', statusKo: '완료',
        txHash: '0xgrant…demo', createdAt: now
      });
      syncGenerationAccount(store.generationAccount, store.hapwRedemptions, store.generations);
      return json(res, 201, { data: { redemption, generationAccount: store.generationAccount, clipBalance: store.clip.balance } });
    }).catch(() => error(res, 400, 'invalid_json', 'Invalid request body'));
  }

  if (req.method === 'POST' && url.pathname === '/api/generations') {
    return readBody(req).then(body => {
      if (!validRequestId(body.requestId)) return error(res, 400, 'invalid_request_id', 'A valid operation id is required');
      const existing = store.generations.find(item => item.requestId === body.requestId);
      if (existing) {
        syncGenerationAccount(store.generationAccount, store.hapwRedemptions, store.generations);
        return json(res, 200, { data: { generation: existing, generationAccount: store.generationAccount, clipBalance: store.clip.balance } });
      }
      const asset = store.assets.find(item => item.id === body.assetId);
      const duration = Number(body.duration);
      const quality = String(body.quality || 'standard');
      const redemption = store.hapwRedemptions.find(item => item.assetId === body.assetId && item.status === '有效');
      if (!asset || !redemption) return error(res, 400, 'license_required', 'Redeem the matching HAPW before using this material');
      if (![15, 30, 60].includes(duration) || !GENERATION_FACTORS[quality] || body.accepted !== true) return error(res, 400, 'invalid_generation', 'Select duration and quality, then confirm the material license');
      const cost = generationCost(duration, quality);
      const clipCost = generationClipCost(cost);
      if (store.generationAccount.balance < cost || redemption.creditsRemaining < cost) return error(res, 400, 'insufficient_generation_credits', 'Insufficient creation credits for this HAPW license');
      if (store.clip.balance < clipCost) return error(res, 400, 'insufficient_clip', 'Insufficient CLIP for this generation');

      const validViews = 12400 + store.generations.length * 3100;
      const now = new Date().toISOString().slice(0, 16).replace('T', ' ');
      const job = {
        id: `video-${Date.now()}`,
        assetId: asset.id,
        title: `${asset.name} · AI 试片 ${store.generations.length + 1}`,
        titleEn: `${asset.nameEn} · AI Cut ${store.generations.length + 1}`,
        titleKo: `${asset.nameKo || asset.nameEn} · AI 테스트 ${store.generations.length + 1}`,
        duration,
        quality,
        creditsUsed: cost,
        clipCost,
        validViews,
        requestId: body.requestId,
        status: '已生成',
        statusEn: 'Generated',
        statusKo: '생성 완료',
        createdAt: now
      };
      redemption.creditsRemaining -= cost;
      store.clip.balance -= clipCost;
      store.generations.unshift(job);
      syncGenerationAccount(store.generationAccount, store.hapwRedemptions, store.generations);
      store.clipTransactions.unshift({
        id: `clip-generation-${Date.now()}`,
        typeCode: 'generationFee',
        type: 'AI 视频生成费',
        typeEn: 'AI video generation fee',
        typeKo: 'AI 영상 생성 수수료',
        amount: -clipCost,
        counterparty: job.title,
        counterpartyEn: job.titleEn,
        counterpartyKo: job.titleKo,
        statusCode: 'completed',
        status: '已完成',
        statusEn: 'Completed',
        statusKo: '완료',
        txHash: '0xplay…demo',
        createdAt: now
      });
      return json(res, 201, { data: { generation: job, generationAccount: store.generationAccount, clipBalance: store.clip.balance } });
    }).catch(() => error(res, 400, 'invalid_json', 'Invalid request body'));
  }

  if (req.method === 'POST' && url.pathname === '/api/conversions') {
    return readBody(req).then(body => {
      if (!validRequestId(body.requestId)) return error(res, 400, 'invalid_request_id', 'A valid operation id is required');
      const existing = store.conversions.find(item => item.requestId === body.requestId);
      if (existing) return json(res, 200, { data: existing });
      const asset = store.assets.find(item => item.id === body.assetId);
      const days = Number(body.days);
      if (!asset || !asset.transferable) return error(res, 400, 'asset_unavailable', 'Asset is not available for conversion');
      if (![30, 90, 180].includes(days) || !body.region || body.accepted !== true) return error(res, 400, 'invalid_conversion', 'Please select a region, duration and accept the terms');
      if (store.clip.balance < 18) return error(res, 400, 'insufficient_clip', 'Insufficient CLIP balance');
      const conversion = { id: `conversion-${Date.now()}`, requestId: body.requestId, assetId: asset.id, asset: assetName(asset), region: body.region, days, fee: 18, statusCode: 'submitted', status: '已提交', statusEn: 'Submitted', statusKo: '제출됨', createdAt: new Date().toISOString().slice(0, 10) };
      store.conversions.unshift(conversion);
      store.clip.balance -= 18;
      store.clipTransactions.unshift({
        id: `clip-tx-${Date.now()}`,
        typeCode: 'license',
        type: '授权手续费',
        typeEn: 'License fee',
        amount: -18,
        counterparty: `HAPW ${asset.tokenId} · ${asset.name}`,
        counterpartyEn: `HAPW ${asset.tokenId} · ${asset.nameEn}`,
        counterpartyKo: `HAPW ${asset.tokenId} · ${asset.nameKo || asset.nameEn}`,
        statusCode: 'completed',
        status: '已完成',
        statusEn: 'Completed',
        statusKo: '완료',
        txHash: '0xdemo…clipli',
        createdAt: new Date().toISOString().slice(0, 16).replace('T', ' ')
      });
      asset.status = '授权处理中'; asset.statusEn = 'License processing';
      return json(res, 201, { data: conversion });
    }).catch(() => error(res, 400, 'invalid_json', 'Invalid request body'));
  }

  if (req.method === 'POST' && ['/api/exercises', '/api/transfers'].includes(url.pathname)) {
    return readBody(req).then(body => {
      if (!validRequestId(body.requestId)) return error(res, 400, 'invalid_request_id', 'A valid operation id is required');
      const existing = store.transfers.find(item => item.requestId === body.requestId);
      if (existing) return json(res, 200, { data: existing });
      const asset = store.assets.find(item => item.id === body.assetId);
      const platform = store.externalPlatforms.find(item => item.code === body.platformCode);
      if (!asset || !asset.transferable) return error(res, 400, 'asset_unavailable', 'Asset is not available for exercise');
      if (!platform || body.accepted !== true) return error(res, 400, 'invalid_exercise', 'Select a third-party platform and accept the exercise confirmation');
      const transfer = { id: `exercise-${Date.now()}`, requestId: body.requestId, assetId: asset.id, platformCode: platform.code, direction: `Clipli → ${platform.name}`, directionEn: `Clipli → ${platform.name}`, directionKo: `Clipli → ${platform.name}`, value: asset.value, statusCode: 'awaitingSignature', status: '待签名', statusEn: 'Awaiting signature', statusKo: '서명 대기', createdAt: new Date().toISOString().slice(0, 10) };
      store.transfers.unshift(transfer);
      return json(res, 201, { data: transfer });
    }).catch(() => error(res, 400, 'invalid_json', 'Invalid request body'));
  }

  if (req.method === 'POST' && url.pathname === '/api/clip/hapw-exchanges') {
    return readBody(req).then(body => {
      if (!validRequestId(body.requestId)) return error(res, 400, 'invalid_request_id', 'A valid operation id is required');
      const existing = store.hapwExchanges.find(item => item.requestId === body.requestId);
      if (existing) return json(res, 200, { data: { exchange: existing, clipBalance: store.clip.balance } });
      const asset = store.assets.find(item => item.id === body.assetId);
      if (!asset || !asset.exchangeAvailable || asset.redemptionStatus !== 'available') return error(res, 400, 'hapw_reserve_unavailable', 'This reserve HAPW is unavailable');
      if (body.accepted !== true) return error(res, 400, 'exchange_not_confirmed', 'Confirm the CLIP to HAPW exchange');
      const policy = hapwExchangePolicy();
      if (policy.reached) return error(res, 429, 'hapw_exchange_daily_limit', 'The daily HAPW exchange limit has been reached');
      const quote = hapwExchangeQuote(asset.clipPrice, store.clip.hapwExchangeFeeRate);
      if (store.clip.balance < quote.total) return error(res, 400, 'insufficient_clip', 'Insufficient CLIP balance');
      const now = new Date().toISOString().slice(0, 16).replace('T', ' ');
      const exchange = { id: `hapw-exchange-${Date.now()}`, requestId: body.requestId, assetId: asset.id, ...quote, statusCode: 'completed', status: '已完成', statusEn: 'Completed', statusKo: '완료', createdAt: now };
      asset.exchangeAvailable = false;
      store.clip.balance -= quote.total;
      store.hapwExchanges.unshift(exchange);
      store.clipTransactions.unshift({
        id: `clip-hapw-${Date.now()}`, typeCode: 'hapwExchange', type: 'HAPW 储备兑换', typeEn: 'HAPW reserve exchange', typeKo: 'HAPW 준비금 교환', amount: -quote.total,
        counterparty: `HAPW ${asset.tokenId} · 本金 ${quote.price} + 手续费 ${quote.fee}`,
        counterpartyEn: `HAPW ${asset.tokenId} · price ${quote.price} + fee ${quote.fee}`,
        counterpartyKo: `HAPW ${asset.tokenId} · 가격 ${quote.price} + 수수료 ${quote.fee}`,
        statusCode: 'completed', status: '已完成', statusEn: 'Completed', statusKo: '완료', txHash: '0xhapw…demo', createdAt: now
      });
      return json(res, 201, { data: { exchange, asset, clipBalance: store.clip.balance } });
    }).catch(() => error(res, 400, 'invalid_json', 'Invalid request body'));
  }

  if (req.method === 'POST' && url.pathname === '/api/profile/wallet') {
    return readBody(req).then(body => {
      const providers = ['MetaMask', 'Coinbase Wallet', 'WalletConnect', 'Venly'];
      if (!providers.includes(body.provider)) return error(res, 400, 'invalid_provider', 'Unsupported wallet provider');
      store.profile.walletProvider = body.provider;
      store.profile.wallet = '0x7E…4A91';
      return json(res, 200, { data: store.profile });
    }).catch(() => error(res, 400, 'invalid_json', 'Invalid request body'));
  }

  if (req.method === 'POST' && url.pathname === '/api/profile/overseas') {
    return readBody(req).then(body => {
      const account = String(body.account || '').trim();
      if (!/^[A-Za-z0-9_-]{4,64}$/.test(account)) return error(res, 400, 'invalid_account', 'Enter a valid overseas account');
      store.profile.overseasAccount = account;
      return json(res, 200, { data: store.profile });
    }).catch(() => error(res, 400, 'invalid_json', 'Invalid request body'));
  }

  if (req.method === 'DELETE' && url.pathname === '/api/profile/overseas') {
    store.profile.overseasAccount = '';
    return json(res, 200, { data: store.profile });
  }

  if (req.method === 'PATCH' && url.pathname === '/api/profile/settings') {
    return readBody(req).then(body => {
      ['walletSign', 'expiryReminder'].forEach(key => { if (typeof body[key] === 'boolean') store.profile.settings[key] = body[key]; });
      return json(res, 200, { data: store.profile });
    }).catch(() => error(res, 400, 'invalid_json', 'Invalid request body'));
  }

  return error(res, 404, 'not_found', 'API route not found');
}

function serveStatic(req, res, url) {
  let pathname;
  try { pathname = decodeURIComponent(url.pathname); } catch { return error(res, 400, 'invalid_path', 'Invalid path'); }
  if (pathname === '/') pathname = '/index.html';
  const filePath = path.resolve(PUBLIC, `.${pathname}`);
  if (filePath !== PUBLIC && !filePath.startsWith(`${PUBLIC}${path.sep}`)) return error(res, 403, 'forbidden', 'Forbidden');
  fs.readFile(filePath, (readErr, content) => {
    if (readErr) {
      if (!path.extname(pathname)) return fs.readFile(path.join(PUBLIC, 'index.html'), (fallbackErr, fallback) => fallbackErr ? error(res, 404, 'not_found', 'Not found') : sendFile(res, 'index.html', fallback));
      return error(res, 404, 'not_found', 'Not found');
    }
    sendFile(res, pathname, content);
  });
}

function sendFile(res, pathname, content) {
  const types = { '.html': 'text/html; charset=utf-8', '.css': 'text/css; charset=utf-8', '.js': 'text/javascript; charset=utf-8', '.json': 'application/json; charset=utf-8', '.svg': 'image/svg+xml' };
  res.writeHead(200, { ...SECURITY_HEADERS, 'Content-Type': types[path.extname(pathname)] || 'application/octet-stream', 'Cache-Control': 'no-store' });
  res.end(content);
}

const server = http.createServer((req, res) => {
  const url = new URL(req.url, `http://${req.headers.host || 'localhost'}`);
  if (url.pathname.startsWith('/api/')) return handleApi(req, res, url);
  return serveStatic(req, res, url);
});

server.listen(PORT, HOST, () => console.log(`Clipli running at http://${HOST}:${PORT}`));
