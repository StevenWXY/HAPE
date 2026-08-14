const ACCENTS = new Set(['coral', 'amber', 'teal', 'gold']);

function text(value, fallback = '') {
  return String(value ?? fallback).trim().slice(0, 2000);
}

function externalUrl(value) {
  try {
    const url = new URL(String(value));
    return ['http:', 'https:'].includes(url.protocol) ? url.href : '';
  } catch {
    return '';
  }
}

function normalizeWork(item, index) {
  if (!item || typeof item !== 'object') return null;
  const id = text(item.id || `external-${index}`, `external-${index}`).slice(0, 100);
  const title = text(item.title || item.titleEn);
  if (!id || !title) return null;
  const normalized = {
    id,
    title,
    titleEn: text(item.titleEn, title),
    category: text(item.category, '作品'),
    categoryEn: text(item.categoryEn, 'Work'),
    duration: /^\d{2}:\d{2}$/.test(text(item.duration)) ? text(item.duration) : '00:30',
    views: Math.max(0, Math.floor(Number(item.views) || 0)),
    creator: text(item.creator, 'Clipli Creator'),
    creatorEn: text(item.creatorEn, item.creator || 'Clipli Creator'),
    summary: text(item.summary),
    summaryEn: text(item.summaryEn, item.summary),
    description: text(item.description, item.summary),
    descriptionEn: text(item.descriptionEn, item.summaryEn || item.summary),
    format: text(item.format, 'AI 视频'),
    formatEn: text(item.formatEn, 'AI video'),
    linkedAssetId: text(item.linkedAssetId).slice(0, 100),
    haiwenUrl: externalUrl(item.haiwenUrl),
    accent: ACCENTS.has(item.accent) ? item.accent : 'coral',
    source: text(item.source, 'external').slice(0, 40)
  };
  for (const suffix of ['Es', 'Ja', 'Fr', 'Ko']) {
    for (const key of ['title', 'category', 'creator', 'summary', 'description', 'format']) {
      if (item[key + suffix]) normalized[key + suffix] = text(item[key + suffix]);
    }
  }
  return normalized;
}

function normalizeWorks(items) {
  if (!Array.isArray(items)) throw new TypeError('invalid_work_source');
  return items.map(normalizeWork).filter(Boolean);
}

module.exports = { normalizeWork, normalizeWorks };
