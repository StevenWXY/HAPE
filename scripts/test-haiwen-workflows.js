const {chromium} = require('playwright');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const base = process.env.CLIPLI_TEST_URL || 'http://127.0.0.1:4175';
const output = path.resolve('artifacts/haiwen-workflows');

(async () => {
  fs.mkdirSync(output, {recursive: true});
  const browser = await chromium.launch({channel: 'chrome', headless: true});
  const page = await browser.newPage({viewport: {width: 1440, height: 1000}});
  const errors = [];
  page.on('pageerror', error => errors.push(error.message));
  const profile = await page.request.get(base + '/api/profile').then(response => response.json());
  profile.data.externalPlatform = {configured: true, sandbox: false, code: 'haiwen'};
  let requests = [], sent = 0, previews = 0, failedOnce = false, firstRequestID;
  const templates = Array.from({length: 34}, (_, index) => ({template: {tplId: 100001 + index, name: '接口测试资产 ' + (index + 1), description: '完整版权描述 <script>throw new Error("unsafe")</script>', image: 'https://cdn.example.test/haiwen.png', workId: 500001, worksName: '接口测试作品', worksTypeName: '影像', publishCount: 1000, authors: [{id: 1, name: '测试作者'}], owners: [{id: 2, name: '测试权利人'}]}, migrationReady: true, mapping: {creditYield: 150, version: 'v1'}}));
  await page.route('https://cdn.example.test/**', route => route.fulfill({contentType: 'image/png', body: fs.readFileSync(path.resolve('_icon_logo_.png'))}));
  await page.route('**/api/profile', route => route.fulfill({json: profile}));
  await page.route('**/api/v1/integrations/platform/**', async route => {
    const url = new URL(route.request().url());
    const method = route.request().method();
    const pageNum = Number(url.searchParams.get('page') || 1);
    if (url.pathname.endsWith('/bindings')) return route.fulfill({json: {data: {binding: {status: 'bound', externalUserId: 'external-fixture', phoneMasked: '138****0000'}}}});
    if (url.pathname.endsWith('/templates')) return route.fulfill({json: {data: {items: templates.slice((pageNum - 1) * 20, pageNum * 20), page: pageNum, pageSize: 20, total: 34}}});
    if (url.pathname.endsWith('/asset-counts')) return route.fulfill({json: {data: {items: url.searchParams.get('tplIds').split(',').map(id => ({tplId: Number(id), count: 2}))}}});
    if (url.pathname.endsWith('/works')) return route.fulfill({json: {data: {items: [{workId: 500001, worksName: '接口测试作品', showcase: ['https://cdn.example.test/haiwen.png'], authors: [{name: '测试作者'}], owners: [{name: '测试权利人'}], worksIntroduce: '完整作品介绍', publishNum: 34}], page: 1, pageSize: 20, total: 1}}});
    if (url.pathname.endsWith('/migrations/preview')) {
      previews++;
      return route.fulfill({json: {data: {ready: true, mapping: {creditYield: 150, version: 'v1'}, availableQuantity: 2, requestedQuantity: route.request().postDataJSON().num}}});
    }
    if (url.pathname.endsWith('/migration-requests') && method === 'GET') return route.fulfill({json: {data: {items: requests}}});
    if (url.pathname.endsWith('/migration-requests') && method === 'POST') {
      const body = route.request().postDataJSON(); sent++;
      assert.equal(body.num, 2); assert.equal(body.accepted, true);
      if (!failedOnce) { failedOnce = true; firstRequestID = body.requestId; return route.fulfill({status: 503, json: {error: {code: 'external_platform_unavailable'}}}); }
      assert.equal(body.requestId, firstRequestID);
      requests = [{id: 'request-fixture', userId: profile.data.session.userRef, externalUserId: 'external-fixture', name: '接口测试资产 21', tplId: body.tplId, quantity: body.num, requestNo: body.requestNo, requestId: body.requestId, status: 'pending_review', mappingVersion: 'v1', creditYield: 150, clipGrant: 60}];
      return route.fulfill({status: 201, json: {data: {request: requests[0]}}});
    }
    throw new Error('Unexpected integration request: ' + method + ' ' + url.pathname);
  });
  await page.goto(base + '/#/assets');
  await page.locator('.haiwen-template').first().waitFor();
  assert.equal(await page.locator('.haiwen-template').count(), 20);
  assert(await page.getByText('数量为海文发当前可核销数量', {exact: false}).isVisible());
  await page.locator('[data-catalog-page="2"]').click();
  await page.getByRole('heading', {name: '接口测试资产 34', exact: true}).waitFor();
  assert.equal(await page.locator('.haiwen-template').count(), 14);
  await page.locator('[data-preview-template="100021"]').click();
  await page.locator('dialog[open]').waitFor();
  await page.locator('dialog [type=submit]:enabled').waitFor();
  assert.equal(sent, 0);
  await page.locator('dialog [name=quantity]').fill('2');
  await page.locator('dialog [name=quantity]').press('Tab');
  await page.locator('dialog [type=submit]:enabled').waitFor();
  await page.locator('dialog [name=accepted]').check();
  await page.screenshot({path: path.join(output, 'exchange-preview-desktop.png')});
  await page.locator('dialog [type=submit]').click();
  await page.locator('dialog .form-feedback').filter({hasText: '稍后重试'}).waitFor();
  await page.locator('dialog [type=submit]').click();
  await page.locator('#migration-requests').getByText('待审核', {exact: true}).waitFor();
  assert.equal(sent, 2); assert(previews >= 2);
  assert.equal((await page.request.get(base + '/api/assets').then(response => response.json())).data.assets.length, 0);
  await page.goto(base + '/#/works');
  await page.getByRole('heading', {name: '接口测试作品', exact: true}).waitFor();
  await page.locator('[data-work-templates]').click();
  await page.locator('[data-preview-template="100001"]').waitFor();
  await page.screenshot({path: path.join(output, 'catalog-desktop.png'), fullPage: true});
  for (const width of [390, 320]) {
    await page.setViewportSize({width, height: 844});
    for (const lang of ['zh', 'en', 'es', 'ja', 'fr', 'ko']) {
      await page.locator('[data-language]').selectOption(lang);
      await page.locator('[data-preview-template="100001"]').waitFor();
      const dimensions = await page.evaluate(() => ({width: innerWidth, scroll: document.documentElement.scrollWidth}));
      assert(dimensions.scroll <= width, JSON.stringify({width, lang, dimensions}));
      assert(await page.evaluate(() => { const brand = document.querySelector('.brand-logo').getBoundingClientRect(), actions = document.querySelector('.nav-actions').getBoundingClientRect(); return brand.right <= actions.left || brand.bottom <= actions.top; }), 'Navigation overlaps the brand: ' + width + ' ' + lang);
    }
  }
  await page.locator('[data-language]').selectOption('zh');
  await page.locator('[data-preview-template="100001"]').waitFor();
  await page.screenshot({path: path.join(output, 'catalog-mobile.png'), fullPage: true});
  await page.locator('[data-preview-template="100001"]').click();
  await page.locator('dialog [type=submit]:enabled').waitFor();
  await page.screenshot({path: path.join(output, 'exchange-preview-mobile.png')});
  await page.keyboard.press('Escape');
  assert.equal(await page.locator('dialog').count(), 0);
  // Operator controls use browser-local fixtures; no admin key or upstream write is used.
  let reviews = 0;
  await page.route('**/api/v1/admin/**', async route => {
    const url = new URL(route.request().url());
    if (url.pathname.endsWith('/review')) {
      const body = route.request().postDataJSON();
      assert.equal(body.action, 'approve'); assert.equal(body.accepted, true); reviews++;
      requests[0].status = 'completed';
      return route.fulfill({json: {data: {request: requests[0]}}});
    }
    const items = url.pathname.endsWith('/migration-requests') ? requests : [];
    return route.fulfill({json: {data: {items}}});
  });
  await page.goto(base + '/admin');
  await page.locator('#admin-key').fill('browser-fixture-key');
  await page.locator('#auth-form [type=submit]').click();
  await page.locator('[data-review][data-action=approve]').waitFor();
  page.once('dialog', dialog => dialog.accept());
  await page.locator('[data-review][data-action=approve]').click();
  await page.locator('#migrations-body').getByText('completed', {exact: true}).waitFor();
  assert.equal(reviews, 1);
  assert.deepEqual(errors, []);
  await browser.close();
  console.log('PASS: 34 templates across pages; metadata/images; exchange preview/consent/retry; request history; operator review JSON; six languages at 320/390px; no real writes');
})().catch(error => { console.error(error); process.exit(1); });
