const { chromium } = require('playwright');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const baseURL = process.env.CLIPLI_TEST_URL || 'http://127.0.0.1:4175';
const output = path.resolve('artifacts/real-assets-ui');

(async () => {
  fs.mkdirSync(output, { recursive: true });
  const browser = await chromium.launch({ channel: 'chrome', headless: true });
  const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
  const errors = [];
  page.on('pageerror', error => errors.push(error.message));
  for (const route of ['/', '/assets', '/bind', '/studio', '/convert', '/transfer', '/wallet', '/clip', '/works', '/about']) {
    await page.goto(baseURL + '/#' + route);
    await page.locator('main h1').waitFor();
    assert(!await page.locator('main').innerText().then(text => /Cannot read|NaN|undefined|操作失败/.test(text)), route);
    assert(await page.locator('.nav-actions a[href="#/bind"]').isVisible());
    assert(await page.locator('.nav-actions a[href="#/wallet"]').isVisible());
    if (['/convert', '/transfer'].includes(route)) assert.equal(await page.locator('main button[type=submit]').count(), 0);
  }
  await page.goto(baseURL + '/#/assets');
  await page.getByRole('heading', { name: '还没有划转到 Clipli 的资产' }).waitFor();
  await page.screenshot({path: path.join(output, 'assets-desktop.png'), fullPage: true});
  // A configured partner is mocked only inside this browser test. No SMS or
  // external transfer requests leave the local test process.
  const profile = await page.request.get(baseURL + '/api/profile').then(response => response.json());
  profile.data.externalPlatform.configured = true;
  await page.route('**/api/profile', route => route.fulfill({json: profile}));
  let bound = false;
  let bindingUnavailable = false;
  let smsRequests = 0;
  let bindRequests = 0;
  const binding = {status: 'bound', externalUserId: '100001', phoneMasked: '138****8000', boundAt: '2026-09-08T00:00:00Z'};
  await page.route('**/api/v1/integrations/platform/**', async route => {
    const url = new URL(route.request().url());
    const method = route.request().method();
    if (url.pathname.endsWith('/bindings') && method === 'GET') return route.fulfill(bindingUnavailable ? {status: 502, json: {error: {code: 'external_platform_unavailable'}}} : bound ? {json: {data: {binding}}} : {status: 404, json: {error: {code: 'binding_not_found'}}});
    if (url.pathname.endsWith('/verification-codes')) {
      smsRequests++;
      const body = route.request().postDataJSON();
      assert.equal(body.externalUserId, '100001');
      assert.equal(body.phone, '13800138000');
      return route.fulfill({json: {data: {challenge: {id: 'test-challenge', phoneMasked: '138****8000', expiresAt: new Date(Date.now() + 300000).toISOString()}}}});
    }
    if (url.pathname.endsWith('/bindings') && method === 'POST') {
      bindRequests++;
      const body = route.request().postDataJSON();
      assert.equal(body.verificationId, 'test-challenge');
      if (body.smsCode !== '123456') return route.fulfill({status: 400, json: {error: {code: 'verification_code_invalid'}}});
      bound = true;
      return route.fulfill({json: {data: {binding}}});
    }
    if (url.pathname.endsWith('/templates')) return route.fulfill({json: {data: {items: [{template: {tplId: 100053, name: '接口测试资产'}, migrationReady: true}], page: 1, pageSize: 20, total: 1}}});
    if (url.pathname.endsWith('/asset-counts')) return route.fulfill({json: {data: {items: [{tplId: 100053, count: 2}]}}});
    throw new Error('Unexpected integration call: ' + method + ' ' + url.pathname);
  });
  await page.goto(baseURL + '/#/bind');
  await page.locator('#haiwen-bind-form').waitFor();
  assert(await page.locator('#confirm-binding').isDisabled());
  await page.locator('[name=externalUserId]').fill('100001');
  await page.locator('[name=phone]').fill('13800138000');
  await page.locator('#send-binding-code').click();
  await page.locator('[name=smsCode]:enabled').waitFor();
  assert(await page.locator('#send-binding-code').isDisabled());
  await page.locator('[name=smsCode]').fill('000000');
  await page.locator('[name=accepted]').check();
  await page.locator('#confirm-binding').click();
  await page.getByText('验证码不正确，请检查后重试。').waitFor();
  await page.locator('[name=smsCode]').fill('123456');
  await page.locator('#confirm-binding').click();
  await page.locator('.binding-summary').waitFor();
  assert.equal(smsRequests, 1); assert.equal(bindRequests, 2);
  await page.locator('#load-external-holdings').click();
  await page.getByText('接口测试资产').waitFor();
  // Merely reading external holdings must not add them to the Clipli balance.
  const assets = await page.request.get(baseURL + '/api/assets').then(response => response.json());
  assert.equal(assets.data.assets.length, 0);
  assert.equal(assets.data.clip.balance, 0);
  await page.screenshot({path: path.join(output, 'platforms-bound-desktop.png'), fullPage: true});
  bindingUnavailable = true;
  await page.locator('#refresh-binding').click();
  await page.getByText('暂时无法确认绑定状态，请刷新重试。').waitFor();
  assert.equal(await page.locator('#haiwen-bind-form').count(), 0);
  bindingUnavailable = false;
  bound = false;
  for (const width of [390, 320]) {
    await page.setViewportSize({width, height: 844});
    for (const lang of ['zh', 'en', 'es', 'ja', 'fr', 'ko']) {
      await page.goto(baseURL + '/#/bind');
      await page.reload();
      await page.locator('#haiwen-bind-form').waitFor();
      await page.locator('[data-language]').selectOption(lang);
      await page.locator('#haiwen-bind-form').waitFor();
      const dimensions = await page.evaluate(() => ({scroll: document.documentElement.scrollWidth, width: innerWidth}));
      assert(dimensions.scroll <= dimensions.width, width + 'px ' + lang + ': ' + JSON.stringify(dimensions));
      assert(await page.locator('.nav-actions a[href="#/bind"]').isVisible());
      assert(await page.locator('.nav-actions a[href="#/wallet"]').isVisible());
    }
  }
  await page.setViewportSize({width: 390, height: 844});
  await page.locator('[data-language]').selectOption('zh');
  await page.locator('#haiwen-bind-form').waitFor();
  await page.screenshot({path: path.join(output, 'platforms-mobile.png'), fullPage: true});
  await page.locator('[data-theme-toggle]').click();
  await page.locator('#haiwen-bind-form').waitFor();
  await page.screenshot({path: path.join(output, 'platforms-mobile-light.png'), fullPage: true});
  assert.deepEqual(errors, []);
  await browser.close();
  console.log('PASS: empty portfolios; ten routes; SMS validation; binding and refresh; holdings isolation; six languages at 320/390px; screenshots in ' + output);
})().catch(error => { console.error(error); process.exit(1); });
