const test = require('node:test');
const assert = require('node:assert/strict');
const { normalizeWork, normalizeWorks } = require('../server/adapters/works/normalize');

test('external works are normalized to the stable frontend contract', () => {
  const work = normalizeWork({ id: 'partner-1', title: 'Partner work', views: '42', accent: 'unknown', haiwenUrl: 'https://example.com/work' }, 0);
  assert.equal(work.views, 42);
  assert.equal(work.accent, 'coral');
  assert.equal(work.haiwenUrl, 'https://example.com/work');
  assert.equal(work.source, 'external');
});

test('unsafe links and unusable records are rejected', () => {
  assert.equal(normalizeWork({ id: 'bad', title: 'Bad link', haiwenUrl: 'javascript:alert(1)' }, 0).haiwenUrl, '');
  assert.deepEqual(normalizeWorks([null, { id: 'missing-title' }]), []);
});

test('localized partner fields preserve Korean content', () => {
  const work = normalizeWork({ id: 'partner-ko', title: '작품', titleKo: '한국어 작품', summaryKo: '한국어 소개' }, 0);
  assert.equal(work.titleKo, '한국어 작품');
  assert.equal(work.summaryKo, '한국어 소개');
});
