// External holdings are read from the partner; only confirmed transfers appear
// in the Clipli portfolio. This file never creates local sample assets.
const PLATFORM_API = '/api/v1/integrations/platform';

function emptyTableRow(columns) {
  return '<tr><td colspan="' + columns + '" class="empty-inline">' + t('noRecordsYet') + '</td></tr>';
}

function assetEmptyState(title = 'noAssetsYet') {
  return '<section class="asset-empty"><span class="empty-symbol" aria-hidden="true">⊞</span><div><h2>' + t(title) + '</h2><p>' + t('assetEmptyLead') + '</p></div><a class="button primary" href="#/bind">' + t('connectFirstPlatform') + ' →</a></section>';
}

async function renderAssets() {
  const data = await api('/api/assets');
  data.clip.balance = data.clip.availableBalance ?? data.clip.balance;
  const rows = data.assets.map(asset => '<tr><td><strong>' + local(asset, 'name') + '</strong><small>HAPW ' + escapeHtml(asset.tokenId) + '</small></td><td>' + escapeHtml(data.platforms.find(item => item.code === asset.external.providerCode)?.name || asset.external.providerCode) + '<small>' + escapeHtml(asset.external.providerAssetId) + '</small></td><td>' + local(asset, 'rightsHolder') + '</td><td><span class="status">' + local(asset, 'status') + '</span></td><td>' + escapeHtml(asset.external.transferId) + '</td><td><a class="text-link" href="#/studio">' + t(asset.redemptionStatus === 'redeemed' ? 'openStudio' : 'viewGeneration') + ' →</a></td></tr>').join('');
  const cards = [[t('hapwHoldings'), data.assets.length, 'HAPW', t('confirmedAssetsLead')], [t('generationCredits'), data.generationAccount.balance, t('creditsUnit'), t('generationCreditsLead')], [t('clipBalance'), data.clip.balance, 'CLIP', t('balanceFromRecords')]].map((item, index) => '<article class="token-card ' + ['hapw-card', 'credit-card', 'clip-card'][index] + '"><div class="token-heading"><span class="token-mark">' + ['W', 'C', 'P'][index] + '</span><h2>' + item[0] + '</h2></div><p>' + item[3] + '</p><div class="token-number">' + money(item[1]) + '<small> ' + item[2] + '</small></div></article>').join('');
  const reserves = data.assets.filter(asset => asset.exchangeAvailable && asset.redemptionStatus === 'available' && asset.clipPrice > 0);
  const policy = data.clip.hapwExchangePolicy;
  const reserveRows = reserves.map(asset => {
    const total = asset.clipPrice + Math.ceil(asset.clipPrice * data.clip.hapwExchangeFeeRate);
    const canExchange = !policy.reached && data.clip.balance >= total;
    return '<article class="reserve-row"><div><h3>' + local(asset, 'name') + '</h3><p>' + local(asset, 'authorizationScope') + '</p></div><strong>' + money(total) + ' CLIP <small>' + t('includesFee') + '</small></strong><button class="button secondary" data-hapw-exchange="' + escapeHtml(asset.id) + '" ' + (canExchange ? '' : 'disabled') + '>' + t(policy.reached ? 'reserveDailyLimit' : canExchange ? 'exchangeToHapw' : 'error_insufficient_clip') + '</button></article>';
  }).join('');
  const txRows = data.clipTransactions.map(item => '<tr><td>' + transactionTypeLabel(item) + '</td><td>' + local(item, 'counterparty') + '</td><td>' + (item.amount > 0 ? '+' : '') + money(item.amount) + ' CLIP</td><td>' + escapeHtml(item.txHash || '—') + '</td><td>' + escapeHtml(item.createdAt) + '</td></tr>').join('');
  const exerciseRows = data.exercises.map(item => '<tr><td>' + escapeHtml(data.assets.find(asset => asset.id === item.assetId)?.tokenId || item.assetId) + '</td><td>' + transferDirectionLabel(item) + '</td><td>' + statusLabel(item) + '</td><td>' + escapeHtml(item.createdAt) + '</td></tr>').join('');
  shell(pageHead(t('assetsEyebrow'), t('assetsTitle'), t('realAssetsOnly'), '<div class="filter-row"><a class="button primary" href="#/bind">' + t('navPlatforms') + '</a><button class="button" id="refresh-portfolio">' + t('refreshAssets') + '</button></div>') +
    (!data.assets.length ? assetEmptyState() : '<section class="notice">' + t('confirmedAssetsLead') + '</section>') +
    '<section class="ledger-grid three">' + cards + '</section>' +
    '<section class="data-section"><div class="data-head"><h2>' + t('confirmedPortfolio') + '</h2><a class="text-link" href="#/studio">' + t('openStudio') + ' →</a></div><div class="table-wrap"><table><thead><tr><th>' + t('asset') + '</th><th>' + t('assetSource') + '</th><th>' + t('rightsHolder') + '</th><th>' + t('status') + '</th><th>' + t('transferReceipt') + '</th><th>' + t('details') + '</th></tr></thead><tbody>' + (rows || emptyTableRow(6)) + '</tbody></table></div></section>' +
    '<section class="reserve-section"><header><div><h2>' + t('reserveTitle') + '</h2><p>' + t('reserveRealLead') + '</p></div><span class="fee-badge">5% · ' + t('exchangeFee') + '</span></header><div class="reserve-list">' + (reserveRows || '<p class="empty-inline">' + t('noReserveYet') + '</p>') + '</div></section>' +
    '<section class="data-section"><div class="data-head"><h2>' + t('recentTransfers') + '</h2><a class="text-link" href="#/transfer">' + t('transferAsset') + ' →</a></div><div class="table-wrap"><table><thead><tr><th>HAPW</th><th>' + t('exerciseTarget') + '</th><th>' + t('status') + '</th><th>' + t('date') + '</th></tr></thead><tbody>' + (exerciseRows || emptyTableRow(4)) + '</tbody></table></div></section>' +
    '<section class="data-section"><div class="data-head"><h2>' + t('clipHistory') + '</h2><a class="text-link" href="#/clip">' + t('mechanismDetails') + ' →</a></div><div class="table-wrap"><table><thead><tr><th>' + t('txType') + '</th><th>' + t('txAsset') + '</th><th>' + t('amount') + '</th><th>' + t('txHash') + '</th><th>' + t('date') + '</th></tr></thead><tbody>' + (txRows || emptyTableRow(5)) + '</tbody></table></div></section>');
  document.getElementById('refresh-portfolio').addEventListener('click', () => render());
  document.querySelectorAll('[data-hapw-exchange]').forEach(button => button.addEventListener('click', () => openHapwExchangeConfirm(reserves.find(asset => asset.id === button.dataset.hapwExchange), data.clip.hapwExchangeFeeRate, policy)));
  await mountHaiwenAssets();
}

async function renderBind() {
  const [profile, platformData] = await Promise.all([api('/api/profile'), api('/api/v1/platforms')]);
  const platforms = platformData.items;
  const configured = profile.externalPlatform?.configured === true;
  state.platformSandbox = profile.externalPlatform?.sandbox === true;
  const userId = profile.session?.userRef;
  let binding = null;
  let bindingError = '';
  if (configured && userId) {
    try { binding = (await api(PLATFORM_API + '/bindings?userId=' + encodeURIComponent(userId))).binding; }
    catch (error) { if (error.name === 'AbortError') throw error; if (error.code !== 'binding_not_found') bindingError = error.message; }
  }
  const others = platforms.filter(item => item.code !== 'haiwen').map(item => '<article class="platform-card upcoming"><span class="platform-monogram" aria-hidden="true">' + escapeHtml(item.name[0]) + '</span><div><h3>' + escapeHtml(item.name) + '</h3><p>' + t(item.code === 'opensea' ? 'openseaComingLead' : 'platformComingLead') + '</p></div><span class="status pending">' + t('integrationComing') + '</span><a class="text-link" href="' + safeExternalUrl(item.url) + '" target="_blank" rel="noopener noreferrer">' + t('visitPlatform') + ' ↗</a></article>').join('');
  const form = '<form id="haiwen-bind-form" class="form-grid"><label class="field"><span>' + t('haiwenUserId') + '</span><input name="externalUserId" autocomplete="username" required maxlength="100" aria-describedby="external-id-help"><small id="external-id-help">' + t('haiwenUserIdHelp') + '</small></label><label class="field"><span>' + t('bindingPhone') + '</span><input name="phone" type="tel" autocomplete="tel" required pattern="(?:\\+?86)?1[3-9][0-9]{9}" placeholder="13800000000"></label><div class="verification-row"><label class="field"><span>' + t('smsCode') + '</span><input name="smsCode" inputmode="numeric" autocomplete="one-time-code" pattern="[0-9]{4,8}" minlength="4" maxlength="8" required disabled></label><button id="send-binding-code" class="button" type="button">' + t('sendCode') + '</button></div><p id="binding-feedback" class="form-feedback" role="status" aria-live="polite"></p><label class="check"><input name="accepted" type="checkbox" required><span>' + t('acceptPlatformBinding') + '</span></label><button class="button primary wide" id="confirm-binding" type="submit" disabled>' + t('verifyAndBind') + '</button></form>';
  const bound = binding ? '<div class="binding-summary"><span class="status">✓ ' + t('platformBound') + '</span><dl><div><dt>' + t('haiwenUserId') + '</dt><dd>' + escapeHtml(binding.externalUserId) + '</dd></div><div><dt>' + t('bindingPhone') + '</dt><dd>' + escapeHtml(binding.phoneMasked || '—') + '</dd></div><div><dt>' + t('boundAt') + '</dt><dd>' + escapeHtml(binding.boundAt ? new Date(binding.boundAt).toLocaleString() : '—') + '</dd></div></dl><p>' + t('bindingDoesNotTransfer') + '</p><div class="filter-row"><button class="button secondary" id="load-external-holdings">' + t('viewExternalHoldings') + '</button><a class="button" href="#/assets">' + t('confirmedPortfolio') + '</a></div></div>' : '';
  shell('<div class="platform-page">' + pageHead('PLATFORMS', t('bindTitle'), t('bindLead'), '<a class="text-link" href="#/assets">' + t('backAssets') + ' →</a>') +
    '<ol class="platform-steps"><li><b>01</b>' + t('bindingStep1') + '</li><li><b>02</b>' + t('bindingStep2') + '</li><li><b>03</b>' + t('bindingStep3') + '</li></ol>' +
    '<section class="haiwen-panel"><header><span class="platform-monogram">海</span><div><h2>' + t('haiwenPlatform') + '</h2><p>' + t('haiwenReadyLead') + (state.platformSandbox ? ' · ' + t('sandboxConnection') : '') + '</p></div><button class="button" id="refresh-binding">' + t('refreshStatus') + '</button></header><div class="haiwen-content">' +
    (bindingError ? '<div class="notice" role="alert">' + escapeHtml(bindingError) + '<p>' + t('bindingStatusUnknown') + '</p></div>' : binding ? bound : configured ? form : '<div class="platform-unavailable"><h3>' + t('connectionUnavailable') + '</h3><p>' + t('connectionUnavailableLead') + '</p></div>') +
    '<aside class="binding-help"><h3>' + t('realAssetsOnly') + '</h3><p>' + t('bindingDoesNotTransfer') + '</p><p>' + t(state.platformSandbox ? 'sandboxTransferHelp' : 'transferHelp') + '</p><a class="text-link" href="' + safeExternalUrl(platforms.find(item => item.code === 'haiwen')?.url) + '" target="_blank" rel="noopener noreferrer">' + t('openHaiwen') + ' ↗</a></aside></div>' +
    '<div id="external-holdings" class="external-holdings" aria-live="polite"></div></section><section class="other-platforms"><h2>' + t('otherPlatforms') + '</h2><div class="platform-grid">' + others + '</div></section></div>');
  document.getElementById('refresh-binding').addEventListener('click', () => render());
  if (binding) document.getElementById('load-external-holdings').addEventListener('click', event => loadExternalHoldings(userId, 1, event.currentTarget));
  if (!binding && configured && !bindingError) bindHaiwenForm(userId);
}

function bindHaiwenForm(userId) {
  const form = document.getElementById('haiwen-bind-form');
  const send = document.getElementById('send-binding-code');
  const submit = document.getElementById('confirm-binding');
  const feedback = document.getElementById('binding-feedback');
  let challenge = null;
  let busy = false;
  let resendAt = state.bindingResendAt || 0;
  let bindRequestId = '';
  const show = (message, error = false) => { feedback.textContent = message; feedback.classList.toggle('is-error', error); };
  const invalidate = () => { challenge = null; bindRequestId = ''; form.smsCode.value = ''; form.smsCode.disabled = true; submit.disabled = true; show(''); };
  const tick = () => {
    const seconds = Math.max(0, Math.ceil((resendAt - Date.now()) / 1000));
    send.disabled = busy || seconds > 0;
    send.textContent = seconds ? t('resendCode') + ' (' + seconds + 's)' : t('sendCode');
    if (challenge && Date.parse(challenge.expiresAt) <= Date.now()) { invalidate(); show(t('error_verification_expired'), true); }
  };
  form.externalUserId.addEventListener('input', invalidate);
  form.phone.addEventListener('input', invalidate);
  send.addEventListener('click', async () => {
    if (busy || !form.externalUserId.reportValidity() || !form.phone.reportValidity()) return;
    busy = true; tick(); show(t('sendingCode'));
    const identity = { userId, externalUserId: form.externalUserId.value.trim(), phone: form.phone.value.trim() };
    form.externalUserId.readOnly = true; form.phone.readOnly = true;
    try {
      const result = await api(PLATFORM_API + '/verification-codes', {method: 'POST', body: JSON.stringify({...identity, requestId: operationId('binding-sms')})});
      challenge = {...result.challenge, ...identity};
      bindRequestId = operationId('platform-bind');
      state.bindingResendAt = resendAt = Date.now() + 60000;
      form.smsCode.disabled = false; submit.disabled = false;
      show(t('codeSent') + ' ' + (result.challenge.phoneMasked || '')); form.smsCode.focus();
    } catch (error) { show(error.message, true); }
    finally { busy = false; form.externalUserId.readOnly = false; form.phone.readOnly = false; tick(); }
  });
  form.addEventListener('submit', async event => {
    event.preventDefault();
    if (busy || !challenge || !form.reportValidity()) return;
    busy = true; submit.disabled = true; tick(); show(t('verifyingBinding'));
    try {
      const result = await api(PLATFORM_API + '/bindings', {method: 'POST', body: JSON.stringify({userId, externalUserId: challenge.externalUserId, phone: challenge.phone, smsCode: form.smsCode.value.trim(), verificationId: challenge.id, requestId: bindRequestId})});
      if (result.binding?.status !== 'bound') throw new Error(t('bindingStatusUnknown'));
      toast(t('platformBound'));
      if (state.route === '/bind') await render();
    } catch (error) { show(error.message, true); submit.disabled = false; }
    finally { busy = false; tick(); }
  });
  clearInterval(state.bindingTimer);
  state.bindingTimer = setInterval(tick, 1000); tick();
}

async function loadExternalHoldings(userId, page, button) {
  const container = document.getElementById('external-holdings');
  button.disabled = true;
  try {
    await loadHaiwenCatalog(container, {userId, page, sandbox: state.platformSandbox});
  } finally { button.disabled = false; }
}
