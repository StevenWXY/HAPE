const {chromium} = require('playwright');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const base = process.env.CLIPLI_TEST_URL || 'http://127.0.0.1:4176';

(async () => {
  const totals = {};
  for (const resource of ['works', 'templates']) {
    const seen = new Set(); let total = 0, images = 0;
    for (let page = 1; ; page++) {
      const response = await fetch(base + '/api/v1/integrations/platform/' + resource + '?page=' + page + '&pageSize=20');
      assert(response.ok, 'Live catalog read failed: ' + response.status);
      const {data} = await response.json(); total = data.total;
      for (const item of data.items) {
        const id = resource === 'templates' ? item.template.tplId : item.workId;
        assert(!seen.has(id), 'Duplicate source ID across pages'); seen.add(id);
        if (resource === 'templates' ? item.template.image : item.showcase?.some(Boolean)) images++;
      }
      if (seen.size >= total) break;
      assert(data.items.length > 0, 'Pagination ended before source total');
    }
    assert.equal(seen.size, total); totals[resource] = {total, images};
  }
  const output = path.resolve('artifacts/haiwen-live'); fs.mkdirSync(output, {recursive: true});
  const browser = await chromium.launch({channel: 'chrome', headless: true});
  const page = await browser.newPage({viewport: {width: 1440, height: 1000}});
  const errors = []; page.on('pageerror', error => errors.push(error.message));
  for (const route of ['works', 'haiwen']) {
    await page.goto(base + '/#/' + route);
    await page.locator('.haiwen-template').first().waitFor();
    await page.locator('.haiwen-image img').evaluateAll(imgs => Promise.all(imgs.slice(0, 2).map(img => img.decode())));
    await page.screenshot({path: path.join(output, route + '-desktop.png')});
    assert(await page.locator('.haiwen-image img').evaluateAll(imgs => imgs.some(img => img.complete && img.naturalWidth > 0)), 'Source images are blank');
  }
  await page.setViewportSize({width: 390, height: 844});
  await page.screenshot({path: path.join(output, 'templates-mobile.png')});
  assert(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth));
  assert.deepEqual(errors, []);
  await browser.close();
  console.log('PASS: live read-only catalog parity and image rendering ' + JSON.stringify(totals));
})().catch(error => { console.error(error); process.exit(1); });
