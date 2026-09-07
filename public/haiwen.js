const haiwenCopy = {
  catalog: ['海文发资产', 'HAIWEN assets', 'Activos HAIWEN', '海文発の資産', 'Actifs HAIWEN', '해문발 자산'],
  redeemable: ['当前可兑换', 'Redeemable now', 'Canjeables ahora', '現在交換可能', 'Échangeables', '현재 교환 가능'],
  scope: ['数量为海文发当前可核销数量，不含挂单、冻结、锁定、已失效或已核销资产。', 'Counts exclude listed, frozen, locked, expired and previously redeemed assets.', 'Los recuentos excluyen activos publicados, congelados, bloqueados, vencidos y canjeados.', '出品中・凍結・ロック・失効・核销済みの資産を除いた数量です。', 'Les quantités excluent les actifs en vente, gelés, bloqués, expirés et déjà échangés.', '판매 중, 동결, 잠금, 만료 및 이미 상각된 자산은 제외됩니다.'],
  authors: ['作者', 'Authors', 'Autores', '作者', 'Auteurs', '작가'],
  type: ['作品类型', 'Work type', 'Tipo de obra', '作品の種類', 'Type d’œuvre', '작품 유형'],
  owners: ['权利人', 'Rights holders', 'Titulares', '権利者', 'Ayants droit', '권리자'],
  published: ['发行数量', 'Issued quantity', 'Cantidad emitida', '発行数', 'Quantité émise', '발행 수량'],
  work: ['所属作品', 'Work', 'Obra', '作品', 'Œuvre', '작품'],
  templates: ['版权模板', 'Copyright templates', 'Plantillas de derechos', '著作権テンプレート', 'Modèles de droits', '저작권 템플릿'],
  preview: ['兑换', 'Exchange', 'Canjear', '交換', 'Échanger', '교환'],
  quantity: ['兑换数量', 'Quantity', 'Cantidad', '交換数', 'Quantité', '교환 수량'],
  request: ['提交兑换申请', 'Request exchange', 'Solicitar canje', '交換を申請', 'Demander l’échange', '교환 신청'],
  consent: ['同意经审核后核销所选海文发资产，兑换不可撤销。', 'I agree to the irreversible write-off of the selected HAIWEN assets after approval.', 'Acepto la cancelación irreversible de los activos seleccionados tras la aprobación.', '承認後に選択した海文発資産を取消不能で核销することに同意します。', 'J’accepte la radiation irréversible des actifs sélectionnés après approbation.', '승인 후 선택한 해문발 자산의 되돌릴 수 없는 상각에 동의합니다.'],
  requests: ['兑换申请', 'Exchange requests', 'Solicitudes de canje', '交換申請', 'Demandes d’échange', '교환 신청'],
  pending_review: ['待审核', 'Awaiting review', 'Pendiente de revisión', '審査待ち', 'En attente', '심사 대기'],
  processing: ['处理中', 'Processing', 'Procesando', '処理中', 'En cours', '처리 중'],
  completed: ['已完成', 'Completed', 'Completado', '完了', 'Terminé', '완료'],
  rejected: ['未通过', 'Rejected', 'Rechazado', '却下', 'Refusé', '거절됨'],
  attention_required: ['待运营处理', 'Needs operator attention', 'Requiere revisión', '運営の確認待ち', 'À vérifier', '운영자 확인 필요'],
  airdrops: ['CLIP 空投', 'CLIP airdrops', 'Airdrops CLIP', 'CLIPエアドロップ', 'Airdrops CLIP', 'CLIP 에어드롭'],
  reserved: ['空投预留', 'Reserved for airdrops', 'Reservado para airdrops', 'エアドロップ予約額', 'Réservé aux airdrops', '에어드롭 예약'],
  queued: ['等待链上发放', 'Awaiting transfer', 'Pendiente de envío', '送金待ち', 'En attente d’envoi', '전송 대기'],
  submitted: ['链上确认中', 'Confirming on chain', 'Confirmando en cadena', 'チェーン確認中', 'Confirmation en cours', '체인 확인 중'],
  confirmed: ['链上已确认', 'Confirmed on chain', 'Confirmado en cadena', 'チェーン確認済み', 'Confirmé sur la chaîne', '체인 확인 완료'],
  failed: ['发放失败', 'Transfer failed', 'Envío fallido', '送金失敗', 'Échec du transfert', '전송 실패'],
  cancelled: ['已取消', 'Cancelled', 'Cancelado', '取消済み', 'Annulé', '취소됨'],
  cancelDrop: ['取消空投', 'Cancel airdrop', 'Cancelar airdrop', 'エアドロップ取消', 'Annuler l’airdrop', '에어드롭 취소'],
  claim: ['领取至钱包', 'Claim to wallet', 'Recibir en cartera', 'ウォレットで受け取る', 'Recevoir au portefeuille', '지갑으로 받기'],
  ledgerReward: ['计入站内余额', 'Credited to platform balance', 'Saldo de la plataforma', 'サービス残高に加算', 'Crédité au solde', '플랫폼 잔액에 적립'],
  requestSent: ['申请已提交，等待审核', 'Request submitted for review', 'Solicitud enviada', '申請を提出しました', 'Demande envoyée', '심사 신청 완료'],
  refresh: ['刷新', 'Refresh', 'Actualizar', '更新', 'Actualiser', '새로고침'],
  terms: ['审核通过后，海文发资产将核销，并在 Clipli 发放对应授权和奖励。', 'After approval, HAIWEN assets are written off and the matching license and rewards are issued in Clipli.', 'Tras la aprobación, se cancelan los activos y se emiten la licencia y recompensas en Clipli.', '承認後、海文発資産を核销し、Clipliで対応する許諾と報酬を発行します。', 'Après approbation, les actifs sont radiés et la licence et les récompenses sont émises dans Clipli.', '승인 후 해문발 자산을 상각하고 Clipli에서 이용 허가와 보상을 지급합니다.']
};
function hc(key) { return haiwenCopy[key]?.[['zh','en','es','ja','fr','ko'].indexOf(state.lang)] || haiwenCopy[key]?.[1] || t(key); }
function partyNames(items) { return escapeHtml((items || []).map(item => item.name).filter(Boolean).join(' / ') || '—'); }
function injectedWallet(name) {
  const providers = window.ethereum?.providers || (window.ethereum ? [window.ethereum] : []);
  return providers.find(provider => name === 'MetaMask' ? provider.isMetaMask && !provider.isCoinbaseWallet : name === 'Coinbase Wallet' && provider.isCoinbaseWallet);
}
function watchWallet(provider, name) {
  if (!provider?.on) return;
  state.activeWalletProvider = provider;
  state.watchedWallets ||= new WeakSet();
  if (state.watchedWallets.has(provider)) return;
  state.watchedWallets.add(provider);
  const update = async () => {
    if (state.activeWalletProvider !== provider) return;
    try {
      const [accounts, chainId] = await Promise.all([provider.request({method: 'eth_accounts'}), provider.request({method: 'eth_chainId'})]);
      if (!accounts?.[0] || chainId !== '0x38') await api('/api/v1/wallet/connect', {method: 'DELETE'});
      else await api('/api/v1/wallet/connect', {method: 'POST', body: JSON.stringify({provider: name, address: accounts[0], chainId})});
      if (['/wallet', '/assets', '/studio'].includes(state.route)) await render();
    } catch (error) { toast(error.message, 'error'); }
  };
  provider.on('accountsChanged', update); provider.on('chainChanged', update);
}
function sourceDescription(value) {
  const document = new DOMParser().parseFromString(String(value || ''), 'text/html');
  document.querySelectorAll('script, style, iframe, object').forEach(node => node.remove());
  return escapeHtml(document.body.textContent.trim() || '—');
}
function sourceImage(url, name) {
  let safe = '';
  try { const value = new URL(url); if (value.protocol === 'https:' && !value.username && !value.password) safe = value.href; } catch {}
  return '<div class="haiwen-image">' + (safe ? '<img src="' + escapeHtml(safe) + '" alt="' + escapeHtml(name) + '" loading="lazy" referrerpolicy="no-referrer">' : '<span aria-hidden="true">H</span>') + '</div>';
}
function catalogPagination(page, total, size) {
  const last = Math.max(1, Math.ceil(total / size));
  return '<div class="holdings-pagination"><button class="button" data-catalog-page="' + (page - 1) + '" aria-label="' + t('previousPage') + '" ' + (page <= 1 ? 'disabled' : '') + '>←</button><span>' + page + ' / ' + last + ' · ' + money(total) + '</span><button class="button" data-catalog-page="' + (page + 1) + '" aria-label="' + t('nextPage') + '" ' + (page >= last ? 'disabled' : '') + '>→</button></div>';
}
function catalogFailure(container, error, retry) {
  if (error.name === 'AbortError' || !container.isConnected) return;
  container.innerHTML = '<p class="notice" role="alert">' + escapeHtml(error.message) + '</p><button class="button" data-retry>' + hc('refresh') + '</button>';
  container.querySelector('[data-retry]').addEventListener('click', retry);
}
async function loadHaiwenCatalog(container, options = {}) {
  const {userId, page = 1, workId = 0, sandbox = false} = options;
  container.innerHTML = '<p class="empty-inline" role="status">' + t('loadingExternalHoldings') + '</p>';
  try {
    const query = new URLSearchParams({page, pageSize: 20});
    if (workId) query.set('workId', workId);
    const catalog = await api(PLATFORM_API + '/templates?' + query);
    const ids = catalog.items.map(item => item.template.tplId);
    const counts = userId && ids.length ? await api(PLATFORM_API + '/users/' + encodeURIComponent(userId) + '/asset-counts?tplIds=' + ids.join(',')) : {items: []};
    if (!container.isConnected) return;
    const cards = catalog.items.map(item => {
      const template = item.template;
      const count = counts.items.find(value => value.tplId === template.tplId)?.count;
      return '<article class="haiwen-template">' + sourceImage(template.image, template.name) + '<div class="haiwen-template-body"><small>HAIWEN · ' + template.tplId + '</small><h3>' + escapeHtml(template.name) + '</h3><p class="source-description">' + sourceDescription(template.description) + '</p><dl><div><dt>' + hc('work') + '</dt><dd>' + escapeHtml(template.worksName) + '</dd></div><div><dt>' + hc('authors') + '</dt><dd>' + partyNames(template.authors) + '</dd></div><div><dt>' + hc('owners') + '</dt><dd>' + partyNames(template.owners) + '</dd></div><div><dt>' + hc('type') + '</dt><dd>' + escapeHtml(template.worksTypeName || '—') + '</dd></div><div><dt>' + hc('published') + '</dt><dd>' + money(template.publishCount) + '</dd></div></dl><footer><strong>' + hc('redeemable') + ': ' + (count === undefined ? '—' : money(count)) + '</strong>' + (!userId ? '<a class="button" href="#/bind">' + t('navPlatforms') + '</a>' : '<button class="button secondary" data-preview-template="' + template.tplId + '" ' + (!item.migrationReady || !count || sandbox ? 'disabled' : '') + '>' + (sandbox ? t('sandboxConnection') : !item.migrationReady ? t('mappingPending') : hc('preview')) + '</button>') + '</footer></div></article>';
    }).join('');
    container.innerHTML = '<div class="data-head"><h2>' + hc('catalog') + '</h2><button class="button" data-refresh>' + hc('refresh') + '</button></div><p class="notice">' + hc('scope') + (sandbox ? ' ' + t('sandboxHoldingsLead') : '') + '</p><div class="haiwen-template-grid">' + (cards || '<p class="empty-inline">' + t('noRecordsYet') + '</p>') + '</div>' + catalogPagination(catalog.page, catalog.total, catalog.pageSize);
    container.querySelector('[data-refresh]').addEventListener('click', () => loadHaiwenCatalog(container, options));
    container.querySelectorAll('[data-catalog-page]').forEach(button => button.addEventListener('click', () => loadHaiwenCatalog(container, {...options, page: Number(button.dataset.catalogPage)})));
    container.querySelectorAll('[data-preview-template]').forEach(button => button.addEventListener('click', () => openHaiwenExchange(userId, catalog.items.find(item => item.template.tplId === Number(button.dataset.previewTemplate)), counts.items.find(item => item.tplId === Number(button.dataset.previewTemplate)).count)));
    container.querySelectorAll('img').forEach(img => img.addEventListener('error', () => { img.hidden = true; img.parentElement.textContent = 'H'; }, {once: true}));
  } catch (error) { catalogFailure(container, error, () => loadHaiwenCatalog(container, options)); }
}

async function renderHaiwenPage() {
  const profile = await api('/api/profile');
  shell(pageHead('HAIWEN', hc('catalog'), '', '<a class="text-link" href="#/works">' + t('browseWorks') + ' →</a>') + '<section id="haiwen-catalog" class="data-section haiwen-catalog-section"></section>');
  let userId = null;
  try { await api(PLATFORM_API + '/bindings?userId=' + encodeURIComponent(profile.session.userRef)); userId = profile.session.userRef; }
  catch (error) { if (error.name === 'AbortError') throw error; }
  await loadHaiwenCatalog(document.getElementById('haiwen-catalog'), {userId, sandbox: profile.externalPlatform?.sandbox, workId: state.haiwenWorkId || 0});
}

async function renderHaiwenWorks(profile, page = 1) {
  shell(pageHead('HAIWEN', t('worksTitle'), '', '<a class="button" href="#/haiwen">' + hc('templates') + ' →</a>') + '<section id="haiwen-works" class="data-section haiwen-catalog-section"></section>');
  const container = document.getElementById('haiwen-works');
  container.innerHTML = '<p class="empty-inline">' + t('loadingExternalHoldings') + '</p>';
  try {
    const catalog = await api(PLATFORM_API + '/works?page=' + page + '&pageSize=20');
    container.innerHTML = '<div class="haiwen-template-grid">' + (catalog.items.map(work => '<article class="haiwen-template">' + sourceImage(work.showcase?.find(Boolean), work.worksName) + '<div class="haiwen-template-body"><small>HAIWEN · ' + work.workId + '</small><h2>' + escapeHtml(work.worksName) + '</h2><p class="source-description">' + sourceDescription(work.worksIntroduce) + '</p><dl><div><dt>' + hc('authors') + '</dt><dd>' + partyNames(work.authors) + '</dd></div><div><dt>' + hc('owners') + '</dt><dd>' + partyNames(work.owners) + '</dd></div><div><dt>' + hc('published') + '</dt><dd>' + money(work.publishNum) + '</dd></div></dl><button class="button" data-work-templates="' + work.workId + '">' + hc('templates') + ' →</button></div></article>').join('') || '<p class="empty-inline">' + t('noWorksYet') + '</p>') + '</div>' + catalogPagination(catalog.page, catalog.total, catalog.pageSize);
    container.querySelectorAll('[data-catalog-page]').forEach(button => button.addEventListener('click', () => renderHaiwenWorks(profile, Number(button.dataset.catalogPage))));
    container.querySelectorAll('[data-work-templates]').forEach(button => button.addEventListener('click', () => { state.haiwenWorkId = Number(button.dataset.workTemplates); location.hash = '#/haiwen'; }));
    document.querySelector('a[href="#/haiwen"]').addEventListener('click', () => { state.haiwenWorkId = 0; });
  } catch (error) { catalogFailure(container, error, () => renderHaiwenWorks(profile, page)); }
}

function openHaiwenExchange(userId, item, available) {
  const input = {tplId: item.template.tplId, num: 1, requestId: operationId('haiwen-request'), requestNo: operationId('haiwen-write-off')};
  const previousFocus = document.activeElement;
  modalRoot.innerHTML = '<dialog class="haiwen-dialog"><form><h2>' + escapeHtml(item.template.name) + '</h2><p>' + hc('terms') + '</p><label class="field"><span>' + hc('quantity') + '</span><input name="quantity" type="number" min="1" max="' + Math.min(100, available) + '" step="1" value="1" required></label><dl class="exchange-rewards"></dl><p class="form-feedback" role="status"></p><label class="check"><input type="checkbox" name="accepted" required><span>' + hc('consent') + '</span></label><div class="confirm-actions"><button class="button" type="button" data-close>' + t('cancel') + '</button><button class="button primary" type="submit" disabled>' + hc('request') + '</button></div></form></dialog>';
  const dialog = modalRoot.querySelector('dialog');
  const form = dialog.querySelector('form');
  const submit = form.querySelector('[type="submit"]');
  const feedback = form.querySelector('[role="status"]');
  let revision = 0, sending = false;
  const close = () => { if (sending) return; dialog.close(); dialog.remove(); previousFocus?.focus(); };
  dialog.querySelector('[data-close]').addEventListener('click', close);
  dialog.addEventListener('cancel', event => { event.preventDefault(); close(); });
  dialog.addEventListener('click', event => { if (event.target === dialog) close(); });
  const preview = async () => {
    const current = ++revision; submit.disabled = true; form.accepted.checked = false;
    if (!form.quantity.reportValidity()) return;
    input.num = Number(form.quantity.value);
    feedback.textContent = t('loadingExternalHoldings');
    try {
      const result = await api(PLATFORM_API + '/users/' + encodeURIComponent(userId) + '/migrations/preview', {method: 'POST', body: JSON.stringify(input)});
      if (!dialog.isConnected || current !== revision) return;
      if (!result.ready) { feedback.textContent = t(result.reason === 'mapping_required' ? 'mappingPending' : 'error_external_assets_insufficient'); return; }
      const profile = await api('/api/profile');
      if (!dialog.isConnected || current !== revision) return;
      input.walletAddress = profile.wallet || ''; input.chainId = profile.walletChainId || '';
      input.mappingVersion = result.mapping.version; input.creditYield = result.mapping.creditYield;
      const credits = result.mapping.creditYield * input.num, clip = Math.floor(result.mapping.creditYield * 0.4) * input.num;
      form.querySelector('dl').innerHTML = '<div><dt>' + t('generationCredits') + '</dt><dd>' + money(credits) + '</dd></div><div><dt>CLIP</dt><dd>' + money(clip) + '</dd></div><div><dt>' + hc('airdrops') + '</dt><dd>' + escapeHtml(profile.wallet || hc('ledgerReward')) + '</dd></div>';
      feedback.textContent = ''; submit.disabled = false;
    } catch (error) { if (current === revision) feedback.textContent = error.message; }
  };
  form.quantity.addEventListener('change', preview);
  form.addEventListener('submit', async event => {
    event.preventDefault(); if (sending || !form.reportValidity() || submit.disabled) return;
    sending = true; submit.disabled = true; form.quantity.disabled = true;
    try {
      await api(PLATFORM_API + '/users/' + encodeURIComponent(userId) + '/migration-requests', {method: 'POST', body: JSON.stringify({...input, accepted: true})});
      sending = false; close(); toast(hc('requestSent')); location.hash = '#/assets'; if (state.route === '/assets') await render();
    } catch (error) { feedback.textContent = error.message; sending = false; submit.disabled = false; form.quantity.disabled = false; }
  });
  dialog.showModal(); preview();
}

async function mountHaiwenAssets() {
  const profile = await api('/api/profile');
  const section = document.createElement('section'); section.className = 'data-section haiwen-catalog-section';
  document.querySelector('.ledger-grid').after(section);
  section.innerHTML = '<div class="data-head"><h2>' + hc('catalog') + '</h2><a class="button" href="#/haiwen">' + hc('templates') + ' →</a></div><div id="portfolio-haiwen"></div><div id="migration-requests"></div><div id="portfolio-airdrops"></div>';
  if (profile.externalPlatform?.configured) {
    try {
      await api(PLATFORM_API + '/bindings?userId=' + encodeURIComponent(profile.session.userRef));
      await loadHaiwenCatalog(section.querySelector('#portfolio-haiwen'), {userId: profile.session.userRef, sandbox: profile.externalPlatform?.sandbox});
    } catch (error) {
      if (error.name === 'AbortError') throw error;
      section.querySelector('#portfolio-haiwen').innerHTML = error.code === 'binding_not_found' ? '<a class="button" href="#/bind">' + t('navPlatforms') + '</a>' : '<p class="notice" role="alert">' + escapeHtml(error.message) + '</p>';
    }
  }
  try {
    const requests = await api(PLATFORM_API + '/users/' + encodeURIComponent(profile.session.userRef) + '/migration-requests');
    section.querySelector('#migration-requests').innerHTML = '<h2>' + hc('requests') + '</h2>' + (requests.items.map(item => '<article class="haiwen-request"><div><strong>' + escapeHtml(item.name) + ' × ' + item.quantity + '</strong><small>' + escapeHtml(item.requestNo) + '</small><small>' + escapeHtml(item.reason || '') + '</small></div><span class="status">' + hc(item.status) + '</span></article>').join('') || '<p class="empty-inline">' + t('noRecordsYet') + '</p>');
  } catch (error) { if (error.name !== 'AbortError') section.querySelector('#migration-requests').textContent = error.message; }
  const drops = profile.airdrops || [];
  const wallet = profile.wallet;
  const air = section.querySelector('#portfolio-airdrops');
  air.innerHTML = '<div class="data-head"><h2>' + hc('airdrops') + '</h2><span>' + hc('reserved') + ': ' + money(profile.clip?.reserved || 0) + ' CLIP</span></div>' + (drops.map(item => '<article class="haiwen-request"><div><strong>' + money(item.amount) + ' CLIP · ' + hc(item.status) + '</strong><small>' + escapeHtml(item.walletAddress) + '</small><small>' + escapeHtml(item.txHash || item.failureReason || '') + '</small></div>' + (item.status === 'queued' && item.allocationSource === 'redemption-entitlement' ? '<button class="button" data-cancel-airdrop="' + escapeHtml(item.id) + '">' + hc('cancelDrop') + '</button>' : '') + '</article>').join('') || '<p class="empty-inline">' + t('noRecordsYet') + '</p>');
  if (wallet) {
    const assets = await api('/api/assets');
    const claims = assets.assets.filter(item => item.redemptionStatus === 'redeemed' && !drops.some(drop => drop.assetId === item.id && !['failed', 'cancelled'].includes(drop.status)));
    air.insertAdjacentHTML('beforeend', claims.map(item => '<article class="haiwen-request"><strong>' + local(item, 'name') + '</strong><button class="button" data-claim-airdrop="' + escapeHtml(item.id) + '">' + hc('claim') + '</button></article>').join(''));
  }
  air.querySelectorAll('[data-cancel-airdrop], [data-claim-airdrop]').forEach(button => button.addEventListener('click', async () => {
    button.disabled = true;
    try {
      const path = button.dataset.cancelAirdrop ? '/api/v1/wallet/airdrops/' + encodeURIComponent(button.dataset.cancelAirdrop) + '/cancel' : '/api/v1/wallet/airdrops';
      await api(path, {method: 'POST', body: JSON.stringify({assetId: button.dataset.claimAirdrop, requestId: operationId('wallet-airdrop')})});
      await render();
    } catch (error) { toast(error.message, 'error'); button.disabled = false; }
  }));
}
